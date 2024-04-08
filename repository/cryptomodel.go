package repository

type EncryptTask struct {
	Key   []byte
	Data  []byte
	IV    []byte
	Result chan EncryptResult
}

type EncryptResult struct {
	Ciphertext []byte
	IV         []byte
	Err        error
}

type DecryptTask struct {
	Key        []byte
	Ciphertext []byte
	IV         []byte
	Result     chan DecryptResult
}

type DecryptResult struct {
	Plaintext []byte
	Err       error
}

type SsoCheckRequest struct {
	UserID   string
	SsoToken string
	Result   chan SsoCheckResult
}

type SsoCheckResult struct {
	Valid bool
	Err   error
}