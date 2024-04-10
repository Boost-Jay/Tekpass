package networkmanager

import (
	"log"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/zishang520/socket.io/v2/socket"
	"weston.io/Apex-Agent/network/socket"
	router "weston.io/Apex-Agent/network/tekpassrouters"

	
)

var ip = getLocalIPAddress()
var port = "8080"
var host = ip + ":" + port

func StartServer() {
	app := router.InitRouter()

	io := socket.NewServer(nil, nil)
	io.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		network.HandleWebSocket(client, host)
	})

	app.GET("/socket.io/*any", gin.WrapH(io.ServeHandler(nil)))

	// 設置 TLS 設定
	app.Use(func(c *gin.Context) {
		c.Request.URL.Scheme = "https"
		c.Next()
	})

	// 使用 RunTLS 啟動伺服器
	err := app.RunTLS(host, "ca.pem", "privatekey.pem")
	if err != nil {
		log.Fatal(err)
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
