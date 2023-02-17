package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
)

type GinGroup struct {
	instance          gin.IRouter
	originPrefix      string
	prefix            string
	originMiddlewares []httpcontract.Middleware
	middlewares       []httpcontract.Middleware
	lastMiddlewares   []httpcontract.Middleware
}

func NewGinGroup(instance gin.IRouter, prefix string, originMiddlewares []httpcontract.Middleware, lastMiddlewares []httpcontract.Middleware) route.Route {
	return &GinGroup{
		instance:          instance,
		originPrefix:      prefix,
		originMiddlewares: originMiddlewares,
		lastMiddlewares:   lastMiddlewares,
	}
}

func (r *GinGroup) Group(handler route.GroupFunc) {
	var middlewares []httpcontract.Middleware
	middlewares = append(middlewares, r.originMiddlewares...)
	middlewares = append(middlewares, r.middlewares...)
	r.middlewares = []httpcontract.Middleware{}
	prefix := pathToGinPath(r.originPrefix + "/" + r.prefix)
	r.prefix = ""

	handler(NewGinGroup(r.instance, prefix, middlewares, r.lastMiddlewares))
}

func (r *GinGroup) Prefix(addr string) route.Route {
	r.prefix += "/" + addr

	return r
}

func (r *GinGroup) Middleware(middlewares ...httpcontract.Middleware) route.Route {
	r.middlewares = append(r.middlewares, middlewares...)

	return r
}

func (r *GinGroup) Any(relativePath string, handler httpcontract.HandlerFunc) {
	r.getGinRoutesWithMiddlewares().Any(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(handler)}...)
}

func (r *GinGroup) Get(relativePath string, handler httpcontract.HandlerFunc) {
	r.getGinRoutesWithMiddlewares().GET(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(handler)}...)
}

func (r *GinGroup) Post(relativePath string, handler httpcontract.HandlerFunc) {
	r.getGinRoutesWithMiddlewares().POST(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(handler)}...)
}

func (r *GinGroup) Delete(relativePath string, handler httpcontract.HandlerFunc) {
	r.getGinRoutesWithMiddlewares().DELETE(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(handler)}...)
}

func (r *GinGroup) Patch(relativePath string, handler httpcontract.HandlerFunc) {
	r.getGinRoutesWithMiddlewares().PATCH(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(handler)}...)
}

func (r *GinGroup) Put(relativePath string, handler httpcontract.HandlerFunc) {
	r.getGinRoutesWithMiddlewares().PUT(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(handler)}...)
}

func (r *GinGroup) Options(relativePath string, handler httpcontract.HandlerFunc) {
	r.getGinRoutesWithMiddlewares().OPTIONS(pathToGinPath(relativePath), []gin.HandlerFunc{handlerToGinHandler(handler)}...)
}

func (r *GinGroup) Static(relativePath, root string) {
	r.getGinRoutesWithMiddlewares().Static(pathToGinPath(relativePath), root)
}

func (r *GinGroup) StaticFile(relativePath, filepath string) {
	r.getGinRoutesWithMiddlewares().StaticFile(pathToGinPath(relativePath), filepath)
}

func (r *GinGroup) StaticFS(relativePath string, fs http.FileSystem) {
	r.getGinRoutesWithMiddlewares().StaticFS(pathToGinPath(relativePath), fs)
}

func (r *GinGroup) getGinRoutesWithMiddlewares() gin.IRoutes {
	prefix := pathToGinPath(r.originPrefix + "/" + r.prefix)
	r.prefix = ""
	ginGroup := r.instance.Group(prefix)

	var middlewares []gin.HandlerFunc
	ginOriginMiddlewares := middlewaresToGinHandlers(r.originMiddlewares)
	ginMiddlewares := middlewaresToGinHandlers(r.middlewares)
	ginLastMiddlewares := middlewaresToGinHandlers(r.lastMiddlewares)
	middlewares = append(middlewares, ginOriginMiddlewares...)
	middlewares = append(middlewares, ginMiddlewares...)
	middlewares = append(middlewares, ginLastMiddlewares...)
	r.middlewares = []httpcontract.Middleware{}
	if len(middlewares) > 0 {
		return ginGroup.Use(middlewares...)
	} else {
		return ginGroup
	}
}
