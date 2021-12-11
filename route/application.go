package route

import (
	"github.com/gin-gonic/gin"
	"github.com/goravel/framework/support/facades"
)

type Application struct {
}

func (app *Application) Init() *gin.Engine {
	if facades.Config.GetBool("app.debug") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	route := gin.New()

	return route
}
