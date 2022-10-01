package route

import (
	"net/http"

	httpcontract "github.com/goravel/framework/contracts/http"
)

type GroupFunc func(routes Route)

type Engine interface {
	Route
	Run(addr string) error
	Group(GroupFunc)
	Prefix(addr string) Engine
}

type Route interface {
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

	Middleware(...httpcontract.Middleware) Route
}
