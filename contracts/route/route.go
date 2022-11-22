package route

import (
	"net/http"

	httpcontract "github.com/goravel/framework/contracts/http"
)

type GroupFunc func(routes Route)

//go:generate mockery --name=Engine
type Engine interface {
	Route
	Run(addr string) error
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	GlobalMiddleware(...httpcontract.Middleware)
}

//go:generate mockery --name=Route
type Route interface {
	Group(GroupFunc)
	Prefix(addr string) Route
	Middleware(...httpcontract.Middleware) Route

	Any(string, httpcontract.HandlerFunc)
	Get(string, httpcontract.HandlerFunc)
	Post(string, httpcontract.HandlerFunc)
	Delete(string, httpcontract.HandlerFunc)
	Patch(string, httpcontract.HandlerFunc)
	Put(string, httpcontract.HandlerFunc)
	Options(string, httpcontract.HandlerFunc)

	Static(string, string)
	StaticFile(string, string)
	StaticFS(string, http.FileSystem)
}
