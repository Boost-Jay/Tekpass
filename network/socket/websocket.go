package network

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/zishang520/socket.io/v2/socket"
)

var Client *socket.Socket

var RandBytes []byte

func HandleWebSocket(client *socket.Socket, host string) {
	log.Println("WebSocket connection established")

	client.On("tekpass", func(datas ...any) {
		randBytes := make([]byte, 64)
		if _, err := rand.Read(randBytes); err != nil {
			log.Println("Error generating random bytes:", err)
			return
		}

		response := fmt.Sprintf("https://tekpass.com.tw/sso?receiver=%s&token=%s", host, base64.RawURLEncoding.EncodeToString(randBytes))

		RandBytes = randBytes

		client.Emit("acknowledge", response)
	})

	client.On("disconnect", func(...any) {
		log.Println("WebSocket connection closed")
	})

	Client = client
}
