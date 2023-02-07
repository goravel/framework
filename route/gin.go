package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gookit/color"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
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
	outputRoutes(r.instance.Routes())
	color.Greenln("Listening and serving HTTP on " + addr)

	return r.instance.Run([]string{addr}...)
}

func (r *Gin) RunTLS(addr, certFile, keyFile string) error {
	outputRoutes(r.instance.Routes())
	color.Greenln("Listening and serving HTTPS on " + addr)

	return r.instance.RunTLS(addr, certFile, keyFile)
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
