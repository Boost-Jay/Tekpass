package networkmanager

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/zishang520/socket.io/v2/socket"
	"weston.io/Apex-Agent/network/socket"
	router "weston.io/Apex-Agent/network/tekpassrouters"
)

func StartServer() {
	app := router.InitRouter()

	io := socket.NewServer(nil, nil)
	io.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		network.HandleWebSocket(client)
	})

	app.GET("/socket.io/*any", gin.WrapH(io.ServeHandler(nil)))

	// 設置 TLS 設定
	app.Use(func(c *gin.Context) {
		c.Request.URL.Scheme = "https"
		c.Next()
	})

	// 使用 RunTLS 啟動伺服器
	err := app.RunTLS(":8080", "ca.pem", "privatekey.pem")
	if err != nil {
		log.Fatal(err)
	}
}
