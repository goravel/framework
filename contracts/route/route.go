package route

import (
	"net/http"

	httpcontract "github.com/goravel/framework/contracts/http"
)

type GroupFunc func(routes Route)

//go:generate mockery --name=Engine
type Engine interface {
	Route
	Run(host ...string) error
	RunTLS(host ...string) error
	RunTLSWithCert(host, certFile, keyFile string) error
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
	GlobalMiddleware(middlewares ...httpcontract.Middleware)
}

//go:generate mockery --name=Route
type Route interface {
	Group(handler GroupFunc)
	Prefix(addr string) Route
	Middleware(middlewares ...httpcontract.Middleware) Route

	Any(relativePath string, handler httpcontract.HandlerFunc)
	Get(relativePath string, handler httpcontract.HandlerFunc)
	Post(relativePath string, handler httpcontract.HandlerFunc)
	Delete(relativePath string, handler httpcontract.HandlerFunc)
	Patch(relativePath string, handler httpcontract.HandlerFunc)
	Put(relativePath string, handler httpcontract.HandlerFunc)
	Options(relativePath string, handler httpcontract.HandlerFunc)

	Static(relativePath, root string)
	StaticFile(relativePath, filepath string)
	StaticFS(relativePath string, fs http.FileSystem)
}
