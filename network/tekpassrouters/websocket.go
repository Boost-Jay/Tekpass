package tekpassrouters

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {

	// 建立websocket連線
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	log.Println("WebSocket connection established")

	// 事件循環
	for {
		// 讀取消息
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		log.Printf("Received message from WebSocket client: %s", string(message))

		event := string(message)
		if event == "tekpass" {
			randBytes := make([]byte, 64)
			if _, err := rand.Read(randBytes); err != nil {
				log.Println("Error generating random bytes:", err)
				continue
			}

			var localip = getLocalIPAddress()
			if localip == "" {
				log.Println("Error getting local IP address")
				continue
			} else {
				log.Printf("Local IP address: %s", localip)
			}

			response := fmt.Sprintf("https://tekpass.com.tw/sso?receiver=%s:8080&token=%s", localip, base64.RawURLEncoding.EncodeToString(randBytes))
			data := map[string]string{
				"event": "acknowledge",
				"data":  response,
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Println("Error marshaling JSON:", err)
				continue
			}

			// 發送訊息回客戶端
			err = conn.WriteMessage(messageType, jsonData)
			if err != nil {
				log.Println("Error sending message:", err)
				break
			}
		}
	}
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