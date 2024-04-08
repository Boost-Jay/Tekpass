package tekpassrouters

import (
	"github.com/gin-gonic/gin"
	apiv1 "weston.io/Apex-Agent/handle"
)

type rawValue string

const (
	Check     rawValue = "/cgi-bin/TekpassCheck"
	Result    rawValue = "/cgi-bin/TekpassResult"
	VersionNo rawValue = "/"
	// Auth      rawValue = "/tekpass_auth"
	// Token     rawValue = "/tekpass_token"
)

func InitRouter() *gin.Engine {
	app := gin.Default()

	tekpassApiv1 := app.Group("")
	{
		tekpassApiv1.POST(string(Check), apiv1.CheckHandler)
		tekpassApiv1.POST(string(Result), apiv1.ResultHandler)
		tekpassApiv1.GET(string(VersionNo), apiv1.VersionNoHandler)
		// tekpassApiv1.GET(string(Auth), apiv1.AuthHandler)
		// tekpassApiv1.GET(string(Token), apiv1.TokenHandler)
	}

	return app
}