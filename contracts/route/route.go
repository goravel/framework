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
	// GlobalMiddleware registers global middleware to be applied to all routes of the router.
	GlobalMiddleware(middlewares ...contractshttp.Middleware)
	// Run starts the HTTP server and listens for incoming connections on the specified host.
	Run(host ...string) error
	// RunTLS starts the HTTPS server with the provided TLS configuration and listens on the specified host.
	RunTLS(host ...string) error
	// RunTLSWithCert starts the HTTPS server with the provided certificate and key files and listens on the specified host and port.
	RunTLSWithCert(host, certFile, keyFile string) error
	// ServeHTTP serves HTTP requests.
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
}

//go:generate mockery --name=Router
type Router interface {
	// Group creates a new router group with the specified handler.
	Group(handler GroupFunc)
	// Prefix adds a common prefix to the routes registered with the router.
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
	// Resource registers RESTful routes for a resource controller.
	Resource(relativePath string, controller contractshttp.ResourceController)

	// Static registers a new route with path prefix to serve static files from the provided root directory.
	Static(relativePath, root string)
	// StaticFile registers a new route with a specific path to serve a static file from the filesystem.
	StaticFile(relativePath, filepath string)
	// StaticFS registers a new route with a path prefix to serve static files from the provided file system.
	StaticFS(relativePath string, fs http.FileSystem)
}
