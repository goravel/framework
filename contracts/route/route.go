package route

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
)

type GroupFunc func(router Router)

//go:generate mockery --name=Route
type Route interface {
	Router
	// Fallback registers a handler to be executed when no other route was matched.
	Fallback(handler contractshttp.HandlerFunc)
	GlobalMiddleware(middlewares ...contractshttp.Middleware)
	Run(host ...string) error
	RunTLS(host ...string) error
	RunTLSWithCert(host, certFile, keyFile string) error
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
}

//go:generate mockery --name=Router
type Router interface {
	// Group creates a new router group.
	Group(handler GroupFunc)
	// Prefix adds a prefix to the router.
	Prefix(addr string) Router
	// Middleware sets the middleware for the router.
	Middleware(middlewares ...contractshttp.Middleware) Router

	// Any registers a new route responding to all verbs.
	Any(relativePath string, handler contractshttp.HandlerFunc)
	// Get registers a new GET route with the router.
	Get(relativePath string, handler contractshttp.HandlerFunc)
	// Post registers a new POST route with the router.
	Post(relativePath string, handler contractshttp.HandlerFunc)
	// Delete registers a new DELETE route with the router.
	Delete(relativePath string, handler contractshttp.HandlerFunc)
	// Patch registers a new PATCH route with the router.
	Patch(relativePath string, handler contractshttp.HandlerFunc)
	// Put registers a new PUT route with the router.
	Put(relativePath string, handler contractshttp.HandlerFunc)
	// Options registers a new OPTIONS route with the router.
	Options(relativePath string, handler contractshttp.HandlerFunc)
	// Resource route a resource to a controller.
	Resource(relativePath string, controller contractshttp.ResourceController)

	Static(relativePath, root string)
	StaticFile(relativePath, filepath string)
	StaticFS(relativePath string, fs http.FileSystem)
}
