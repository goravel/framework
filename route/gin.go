package route

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gookit/color"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/foundation"
	frameworkhttp "github.com/goravel/framework/http"
)

type Gin struct {
	route.Route
	instance *gin.Engine
}

func NewGin() route.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	return &Gin{instance: engine, Route: NewGinGroup(engine.Group("/"))}
}

func (r *Gin) Run(addr string) error {
	rootApp := foundation.Application{}
	if facades.Config.GetBool("app.debug") && !rootApp.RunningInConsole() {
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

type GinGroup struct {
	instance    gin.IRouter
	prefix      string
	middlewares []httpcontract.Middleware
}

func NewGinGroup(routeGroup gin.IRouter) route.Route {
	return &GinGroup{instance: routeGroup}
}

func (r *GinGroup) Group(handler route.GroupFunc) {
	handler(NewGinGroup(r.getGinRoutesWithMiddlewares()))
}

func (r *GinGroup) Prefix(addr string) route.Route {
	r.prefix += "/" + addr

	return r
}

func (r *GinGroup) Middleware(handlers ...httpcontract.Middleware) route.Route {
	r.middlewares = append(r.middlewares, handlers...)

	return r
}

func (r *GinGroup) Any(relativePath string, handler httpcontract.HandlerFunc) {
	r.getGinRoutesWithMiddlewares().Any(r.pathToGinPath(relativePath), []gin.HandlerFunc{HandlerToGinHandler(handler)}...)
}

func (r *GinGroup) Get(relativePath string, handler httpcontract.HandlerFunc) {
	r.getGinRoutesWithMiddlewares().GET(r.pathToGinPath(relativePath), []gin.HandlerFunc{HandlerToGinHandler(handler)}...)
}

func (r *GinGroup) Post(relativePath string, handler httpcontract.HandlerFunc) {
	r.getGinRoutesWithMiddlewares().POST(r.pathToGinPath(relativePath), []gin.HandlerFunc{HandlerToGinHandler(handler)}...)
}

func (r *GinGroup) Delete(relativePath string, handler httpcontract.HandlerFunc) {
	r.getGinRoutesWithMiddlewares().DELETE(r.pathToGinPath(relativePath), []gin.HandlerFunc{HandlerToGinHandler(handler)}...)
}

func (r *GinGroup) Patch(relativePath string, handler httpcontract.HandlerFunc) {
	r.getGinRoutesWithMiddlewares().PATCH(r.pathToGinPath(relativePath), []gin.HandlerFunc{HandlerToGinHandler(handler)}...)
}

func (r *GinGroup) Put(relativePath string, handler httpcontract.HandlerFunc) {
	r.getGinRoutesWithMiddlewares().PUT(r.pathToGinPath(relativePath), []gin.HandlerFunc{HandlerToGinHandler(handler)}...)
}
func (r *GinGroup) Options(relativePath string, handler httpcontract.HandlerFunc) {
	r.getGinRoutesWithMiddlewares().OPTIONS(r.pathToGinPath(relativePath), []gin.HandlerFunc{HandlerToGinHandler(handler)}...)
}

func (r *GinGroup) Static(relativePath, root string) {
	r.getGinRoutesWithMiddlewares().Static(r.pathToGinPath(relativePath), root)
}

func (r *GinGroup) StaticFile(relativePath, filepath string) {
	r.getGinRoutesWithMiddlewares().StaticFile(r.pathToGinPath(relativePath), filepath)
}

func (r *GinGroup) StaticFS(relativePath string, fs http.FileSystem) {
	r.getGinRoutesWithMiddlewares().StaticFS(r.pathToGinPath(relativePath), fs)
}

func (r *GinGroup) getGinRoutesWithMiddlewares() gin.IRouter {
	ginGroup := r.instance.Group(r.pathToGinPath(r.prefix))
	ginMiddlewares := r.middlewaresToGinHandlers(r.middlewares)
	ginGroup.Use(ginMiddlewares...)

	return ginGroup
}

func (r *GinGroup) pathToGinPath(relativePath string) string {
	ginPath := bracketToColon(mergeSlashForPath(relativePath))
	r.prefix = ""

	return ginPath
}

func (r *GinGroup) middlewaresToGinHandlers(middlewares []httpcontract.Middleware) []gin.HandlerFunc {
	var ginHandlers []gin.HandlerFunc
	for _, middleware := range middlewares {
		ginHandlers = append(ginHandlers, MiddlewareToGinHandler(middleware))
	}

	r.middlewares = []httpcontract.Middleware{}

	return ginHandlers
}

func colonToBracket(relativePath string) string {

	arr := strings.Split(relativePath, "/")
	var newArr []string
	for _, item := range arr {
		if strings.HasPrefix(item, ":") {
			item = "{" + strings.ReplaceAll(item, ":", "") + "}"
		}
		newArr = append(newArr, item)
	}

	return strings.Join(newArr, "/")
}

func bracketToColon(relativePath string) string {
	compileRegex := regexp.MustCompile("\\{(.*?)\\}")
	matchArr := compileRegex.FindAllStringSubmatch(relativePath, -1)

	for _, item := range matchArr {
		relativePath = strings.ReplaceAll(relativePath, item[0], ":"+item[1])
	}

	return relativePath
}

func mergeSlashForPath(path string) string {
	path = strings.ReplaceAll(path, "//", "/")

	return strings.ReplaceAll(path, "//", "/")
}

func HandlerToGinHandler(handler httpcontract.HandlerFunc) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		facades.Response = frameworkhttp.NewGinResponse(ginCtx)
		handler(frameworkhttp.NewGinRequest(ginCtx))
	}
}

func MiddlewareToGinHandler(handler httpcontract.Middleware) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		facades.Response = frameworkhttp.NewGinResponse(ginCtx)
		handler(frameworkhttp.NewGinRequest(ginCtx))
	}
}
