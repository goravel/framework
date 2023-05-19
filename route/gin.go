package route

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	goravelhttp "github.com/goravel/framework/http"
)

type Gin struct {
	route.Route
	config   config.Config
	instance *gin.Engine
}

func NewGin(config config.Config) *Gin {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	if debugLog := getDebugLog(config); debugLog != nil {
		engine.Use(debugLog)
	}

	return &Gin{
		Route: NewGinGroup(engine.Group("/"),
			"",
			[]httpcontract.Middleware{},
			[]httpcontract.Middleware{goravelhttp.GinResponseMiddleware()},
		),
		config:   config,
		instance: engine,
	}
}

func (r *Gin) Fallback(handler httpcontract.HandlerFunc) {
	r.instance.NoRoute(handlerToGinHandler(handler))
}

func (r *Gin) GlobalMiddleware(middlewares ...httpcontract.Middleware) {
	if len(middlewares) > 0 {
		r.instance.Use(middlewaresToGinHandlers(middlewares)...)
	}
	r.Route = NewGinGroup(
		r.instance.Group("/"),
		"",
		[]httpcontract.Middleware{},
		[]httpcontract.Middleware{goravelhttp.GinResponseMiddleware()},
	)
}

func (r *Gin) Run(host ...string) error {
	if len(host) == 0 {
		defaultHost := r.config.GetString("http.host")
		if defaultHost == "" {
			return errors.New("host can't be empty")
		}

		defaultPort := r.config.GetString("http.port")
		if defaultPort == "" {
			return errors.New("port can't be empty")
		}
		completeHost := defaultHost + ":" + defaultPort
		host = append(host, completeHost)
	}

	r.outputRoutes()
	color.Greenln("[HTTP] Listening and serving HTTP on " + host[0])

	return r.instance.Run([]string{host[0]}...)
}

func (r *Gin) RunTLS(host ...string) error {
	if len(host) == 0 {
		defaultHost := r.config.GetString("http.tls.host")
		if defaultHost == "" {
			return errors.New("host can't be empty")
		}

		defaultPort := r.config.GetString("http.tls.port")
		if defaultPort == "" {
			return errors.New("port can't be empty")
		}
		completeHost := defaultHost + ":" + defaultPort
		host = append(host, completeHost)
	}

	certFile := r.config.GetString("http.tls.ssl.cert")
	keyFile := r.config.GetString("http.tls.ssl.key")

	return r.RunTLSWithCert(host[0], certFile, keyFile)
}

func (r *Gin) RunTLSWithCert(host, certFile, keyFile string) error {
	if host == "" {
		return errors.New("host can't be empty")
	}
	if certFile == "" || keyFile == "" {
		return errors.New("certificate can't be empty")
	}

	r.outputRoutes()
	color.Greenln("[HTTPS] Listening and serving HTTPS on " + host)

	return r.instance.RunTLS(host, certFile, keyFile)
}

func (r *Gin) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	r.instance.ServeHTTP(writer, request)
}

func (r *Gin) outputRoutes() {
	if r.config.GetBool("app.debug") && !runningInConsole() {
		for _, item := range r.instance.Routes() {
			fmt.Printf("%-10s %s\n", item.Method, colonToBracket(item.Path))
		}
	}
}
