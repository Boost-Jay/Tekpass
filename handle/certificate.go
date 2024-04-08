package handle

import (
    "log"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"weston.io/Apex-Agent/repository"

)

var (
	// xlh       = "2TNsA6a2C6JQLnZl26dNF7RDv3lBUokiUZdRZED_szx2VsVxKUODgT1DOwgTrs1Zr1IVtunk6d8vNqaB5zW-BhNDYK9HZ1THjZSLuRZ0eO-qPSUuLClQS3p7JMLoGVN24QBSrDUmxBM"
	// xlhPin    = "004309"
	// xlhUser   = "User-1670460972576"
	randBytes []byte
	apexid    string
	ssoID     string
	ssoToken  string
	userData  string
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

	qrKey := randBytes[:32]
	qrIV := randBytes[32:48]

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

	userID := decryptResult.Plaintext[:len(decryptResult.Plaintext)-16]
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
	}

	if ssoCheckResult.Valid {
		if apexid == "" {
			apexid = string(userID)
			extraData := gin.H{
				"message": "SSO token and user ID are valid",
				"user_id": string(userID),
			}
			c.JSON(http.StatusOK, extraData)
			return
		} else {
			if string(userID) != apexid {
				c.AbortWithStatus(400)
				return
			} else {
				extraData := gin.H{
					"message": "SSO token and user ID are valid",
					"user_id": string(userID),
				}
				c.JSON(http.StatusOK, extraData)
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

	c.JSON(http.StatusOK, gin.H{"sso_id": ssoID, "sso_token": userData})
}

func VersionNoHandler(c *gin.Context) {
	c.String(http.StatusOK, "version 0.3")
}