package route

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gookit/color"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/facades"
	goravelhttp "github.com/goravel/framework/http"
)

type Gin struct {
	route.Route
	instance *gin.Engine
}

func NewGin() *Gin {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	if debugLog := getDebugLog(); debugLog != nil {
		engine.Use(debugLog)
	}

	return &Gin{instance: engine, Route: NewGinGroup(
		engine.Group("/"),
		"",
		[]httpcontract.Middleware{},
		[]httpcontract.Middleware{goravelhttp.GinResponseMiddleware()},
	)}
}

func (r *Gin) Run(addr string) error {
	if facades.Config.GetBool("app.debug") && !runningInConsole() {
		routes := r.instance.Routes()
		for _, item := range routes {
			fmt.Printf("%-10s %s\n", item.Method, colonToBracket(item.Path))
		}
	}

	color.Greenln("Listening and serving HTTP on " + addr)

	return r.instance.Run([]string{addr}...)
}

func (r *Gin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.instance.ServeHTTP(w, req)
}

func (r *Gin) GlobalMiddleware(handlers ...httpcontract.Middleware) {
	if len(handlers) > 0 {
		r.instance.Use(middlewaresToGinHandlers(handlers)...)
	}
	r.Route = NewGinGroup(
		r.instance.Group("/"),
		"",
		[]httpcontract.Middleware{},
		[]httpcontract.Middleware{goravelhttp.GinResponseMiddleware()},
	)
}
