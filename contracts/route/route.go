package route

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
)

type GroupFunc func(routes Route)

//go:generate mockery --name=Engine
type Engine interface {
	Route
	Fallback(handler contractshttp.HandlerFunc)
	GlobalMiddleware(middlewares ...contractshttp.Middleware)
	Run(host ...string) error
	RunTLS(host ...string) error
	RunTLSWithCert(host, certFile, keyFile string) error
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
}

//go:generate mockery --name=Route
type Route interface {
	Group(handler GroupFunc)
	Prefix(addr string) Route
	Middleware(middlewares ...contractshttp.Middleware) Route

	Any(relativePath string, handler contractshttp.HandlerFunc)
	Get(relativePath string, handler contractshttp.HandlerFunc)
	Post(relativePath string, handler contractshttp.HandlerFunc)
	Delete(relativePath string, handler contractshttp.HandlerFunc)
	Patch(relativePath string, handler contractshttp.HandlerFunc)
	Put(relativePath string, handler contractshttp.HandlerFunc)
	Options(relativePath string, handler contractshttp.HandlerFunc)

	Static(relativePath, root string)
	StaticFile(relativePath, filepath string)
	StaticFS(relativePath string, fs http.FileSystem)
}
