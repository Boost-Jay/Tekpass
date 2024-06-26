package handle

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	network "weston.io/Apex-Agent/network/socket"
	"weston.io/Apex-Agent/repository"
)

var (
	// xlh       = "2TNsA6a2C6JQLnZl26dNF7RDv3lBUokiUZdRZED_szx2VsVxKUODgT1DOwgTrs1Zr1IVtunk6d8vNqaB5zW-BhNDYK9HZ1THjZSLuRZ0eO-qPSUuLClQS3p7JMLoGVN24QBSrDUmxBM"
	// xlhPin    = "004309"
	// xlhUser   = "User-1670460972576"
	apexid   string
	ssoID    string
	ssoToken string
	userData string
)

// 驗證 SSO token 和用戶 ID。
func CheckHandler(c *gin.Context) {
	log.Println("/cgi-bin/TekpassCheck")

	var reqBody map[string]string
	err := c.BindJSON(&reqBody)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	ssoToken = reqBody["sso_token"]

	log.Printf("sso_id length: %d", len(ssoID))
	log.Printf("sso_token length: %d", len(ssoToken))

	plain, err := base64.StdEncoding.DecodeString(ssoID)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}
	log.Printf("plain: %v", plain)

	qrKey := network.RandBytes[:32]
	qrIV := network.RandBytes[32:48]

	decryptTasks := []repository.DecryptTask{
		{
			Key:        qrKey,
			Ciphertext: plain,
			IV:         qrIV,
			Result:     make(chan repository.DecryptResult),
		},
	}

	repository.ConcurrentDecrypt(decryptTasks)
	decryptResult := <-decryptTasks[0].Result
	if decryptResult.Err != nil {
		c.AbortWithError(500, decryptResult.Err)
		return
	}

	userID := decryptResult.Plaintext[:len(decryptResult.Plaintext)]
	log.Printf("userId: %s", string(userID))

	if ssoToken == "" || string(userID) == "" {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("%v error %s %s", reqBody, string(userID), ssoToken))
		return
	}

	ssoCheckRequests := []repository.SsoCheckRequest{
		{
			UserID:   string(userID),
			SsoToken: ssoToken,
			Result:   make(chan repository.SsoCheckResult),
		},
	}

	repository.ConcurrentSsoCheck(ssoCheckRequests)
	ssoCheckResult := <-ssoCheckRequests[0].Result
	if ssoCheckResult.Err != nil {
		log.Println(ssoCheckResult.Err)
		c.AbortWithError(500, ssoCheckResult.Err)
		return
	} else {
		log.Println(ssoCheckRequests[0].UserID, "已經登入")
	}

	if ssoCheckResult.Valid {
		if apexid == "" {
			apexid = string(userID)
			c.JSON(http.StatusOK, textBody("", http.StatusOK))
			return
		} else {
			if string(userID) != apexid {
				c.AbortWithStatus(400)
				return
			} else {
				c.JSON(http.StatusOK, textBody("", http.StatusOK))
				return
			}
		}
	} else {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

}

// 接收 SSO ID 和用戶數據。
func ResultHandler(c *gin.Context) {
	log.Println("/cgi-bin/TekpassResult")

	// 限制請求體大小為 10KB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10240)

	var reqBody map[string]string
	err := c.BindJSON(&reqBody)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	ssoID = reqBody["sso_id"]
	userData = reqBody["sso_token"]
	log.Printf("sso_id length: %d", len(ssoID))
	log.Printf("userData length: %d", len(userData))

	network.Client.Emit("response", userData)

	c.JSON(http.StatusOK, gin.H{"sso_id": ssoID, "sso_token": userData})
}

func VersionNoHandler(c *gin.Context) {
	c.String(http.StatusOK, "version 0.3")
}

func textBody(body any, status int) map[string]any {
	return map[string]any{
		"isBase64Encoded": false,
		"statusCode":      status,
		"headers":         commonHeaders(),
		"body":            body,
	}
}

func commonHeaders() map[string]any {
	return map[string]any{}
}
