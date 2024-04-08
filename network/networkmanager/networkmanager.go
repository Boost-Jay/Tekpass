package networkmanager

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	https "weston.io/Apex-Agent/network/httpconfig"
	router "weston.io/Apex-Agent/network/tekpassrouters"

)

func StartServer() {
	tlsConfig, err := https.GetRequireHttp()
	if err != nil {
		log.Fatalf("Failed to get TLS configuration: %v", err)
	}

	app := router.InitRouter()

	app.GET("/ws", func(c *gin.Context) {
		router.HandleWebSocket(c.Writer, c.Request)
	})

	server := &http.Server{
		Addr:      ":8080",
		TLSConfig: tlsConfig,
		Handler:   app,
	}

	err = server.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatal(err)
	}
}
