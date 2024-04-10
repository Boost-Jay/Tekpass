package network

import (
	"crypto/rand"
	"encoding/base64"

	// "encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/zishang520/socket.io/v2/socket"
)

var Client *socket.Socket

var RandBytes []byte

func HandleWebSocket(client *socket.Socket) {
	log.Println("WebSocket connection established")

	client.On("tekpass", func(datas ...any) {
		randBytes := make([]byte, 64)
		if _, err := rand.Read(randBytes); err != nil {
			log.Println("Error generating random bytes:", err)
			return
		}

		var localip = getLocalIPAddress()
		if localip == "" {
			log.Println("Error getting local IP address")
			return
		} else {
			log.Printf("Local IP address: %s", localip)
		}

		response := fmt.Sprintf("https://tekpass.com.tw/sso?receiver=%s:8080&token=%s", localip, base64.RawURLEncoding.EncodeToString(randBytes))
		// data := map[string]string{
		// 	"event": "acknowledge",
		// 	"data":  response,
		// }

		// jsonData, err := json.Marshal(data)
		// if err != nil {
		// 	log.Println("Error marshaling JSON:", err)
		// 	return
		// }

		RandBytes = randBytes

		client.Emit("acknowledge", response)
	})

	client.On("disconnect", func(...any) {
		log.Println("WebSocket connection closed")
	})

	Client = client
}

// 獲取本地 IP 地址
func getLocalIPAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
