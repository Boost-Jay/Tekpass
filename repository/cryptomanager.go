package repository

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"log"
	"os/exec"
	"sync"
)

/*
AES-GCM 時，通常建議使用 96 位元 IV = 12-bytes
GCM 在內部運作 CTR，需要 16-byte counter.。 IV 提供了其中的 12 個
另外 4 個是實際的 block-wise counter。
加密算法是128，则對應的key是16位
如果加密算法是256，则對應的key是32位。
https://medium.com/@tony.infisical/guide-to-nodes-crypto-module-for-encryption-decryption-65c077176980
*/

// 使用給定的 key 和 iv 加密 data。如果 iv 為 nil，將生成一個新的隨機 IV。
func AesGCMEncrypt(key, data, iv []byte) ([]byte, []byte, error) {

	// 創建一個 cipher.Block 接口類型的參數，該參數表示一個對稱密鑰分組密碼。
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	if iv == nil {
		iv = make([]byte, 16)
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return nil, nil, err
		}
	}
	// AES 塊加密算法與 GCM 模式結合，返回 AEAD 接口，自動處理認證標籤的生成和驗證。
	aesGCM, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, nil, err
	}

	// 使用 GCM 模式進行加密
	// 將 IV 前置到密文之前，並返回一個包含 IV 和密文的單一字節序列。
	ciphertext := aesGCM.Seal(nil, iv, data, nil)
	return ciphertext, iv, nil
}

func ConcurrentEncrypt(tasks []EncryptTask) {
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go func(task EncryptTask) {
			defer wg.Done()
			ciphertext, iv, err := AesGCMEncrypt(task.Key, task.Data, task.IV)
			task.Result <- EncryptResult{Ciphertext: ciphertext, IV: iv, Err: err}
		}(task)
	}
	wg.Wait()
}

// 使用給定的 key 和 iv 解密 ciphertext。如果 iv 為 nil，將生成一個新的隨機 IV。
func AesGCMDecrypt(key, ciphertext, iv []byte) ([]byte, error) {
	// 創建一個 cipher.Block 接口類型的參數，該參數表示一個對稱密鑰分組密碼。
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if iv == nil {
		iv = make([]byte, 16)
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return nil, err
		}
	}

	// AES 塊加密算法與 GCM 模式結合，返回 AEAD 接口，自動處理認證標籤的生成和驗證。
	aesgcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, err
	}

	// 使用 GCM 模式進行解密
	plaintext, err := aesgcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func ConcurrentDecrypt(tasks []DecryptTask) {
	// var wg sync.WaitGroup
	for _, task := range tasks {
		// wg.Add(1)
		go func(task DecryptTask) {
			// defer wg.Done()
			plaintext, err := AesGCMDecrypt(task.Key, task.Ciphertext, task.IV)
			task.Result <- DecryptResult{Plaintext: plaintext, Err: err}
		}(task)
	}
	// wg.Wait()
}

// 驗證 SSO token
func SsoCheck(userID, ssoToken string, result chan<- SsoCheckResult) {
	cmd := exec.Command("/bin/sso-check-linux-amd64", "https://2hwi7j8zb7.execute-api.ap-northeast-1.amazonaws.com/default/apex-v2", "Free5GC-1699862217954", "m6v8-1hSYujwO2xRne1QYK1EHlwzRu4tfCc0rMepSVxE_ViyVnJPJeDyJ_mwn-DBw-PEKVaK10yEGjiGgCAi1itDhi442v4bQGrn3mRbxJJVbsb4MMTKBSzMhum8u4a-G6mxgZrbMWbiJmr_7AzNatm8_RlGQ5y9", "free5gc", userID, ssoToken)
	output, err := cmd.Output()
	if err != nil {
		result <- SsoCheckResult{Valid: false, Err: err}
		return
	}

	log.Printf("stdout: %s", string(output))
	result <- SsoCheckResult{Valid: true, Err: nil}
}

func ConcurrentSsoCheck(requests []SsoCheckRequest) {
	var wg sync.WaitGroup
	for _, req := range requests {
		wg.Add(1)
		go func(req SsoCheckRequest) {
			defer wg.Done()
			SsoCheck(req.UserID, req.SsoToken, req.Result)
		}(req)
	}
	wg.Wait()
}