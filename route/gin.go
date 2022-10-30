package route

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

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
	if debugLog := getDebugLog(); debugLog != nil {
		engine.Use(debugLog)
	}

	return &Gin{instance: engine, Route: NewGinGroup(
		engine.Group("/"),
		"",
		[]httpcontract.Middleware{},
	)}
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

func (r *Gin) GlobalMiddleware(handlers ...httpcontract.Middleware) {
	r.instance.Use(middlewaresToGinHandlers(handlers)...)
	r.Route = NewGinGroup(
		r.instance.Group("/"),
		"",
		[]httpcontract.Middleware{},
	)
}

type GinGroup struct {
	instance          gin.IRouter
	originPrefix      string
	originMiddlewares []httpcontract.Middleware
	prefix            string
	middlewares       []httpcontract.Middleware
}

func NewGinGroup(instance gin.IRouter, prefix string, originMiddlewares []httpcontract.Middleware) route.Route {
	return &GinGroup{
		instance:          instance,
		originPrefix:      prefix,
		originMiddlewares: originMiddlewares,
	}
}

func (r *GinGroup) Group(handler route.GroupFunc) {
	var middlewares []httpcontract.Middleware
	middlewares = append(middlewares, r.originMiddlewares...)
	middlewares = append(middlewares, r.middlewares...)
	r.middlewares = []httpcontract.Middleware{}
	prefix := pathToGinPath(r.originPrefix + "/" + r.prefix)
	r.prefix = ""

	handler(NewGinGroup(r.instance, prefix, middlewares))
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
	middlewares = append(middlewares, ginOriginMiddlewares...)
	middlewares = append(middlewares, ginMiddlewares...)
	r.middlewares = []httpcontract.Middleware{}
	if len(middlewares) > 0 {
		return ginGroup.Use(middlewares...)
	} else {
		return ginGroup
	}
}

func pathToGinPath(relativePath string) string {
	return bracketToColon(mergeSlashForPath(relativePath))
}

func middlewaresToGinHandlers(middlewares []httpcontract.Middleware) []gin.HandlerFunc {
	var ginHandlers []gin.HandlerFunc
	for _, item := range middlewares {
		ginHandlers = append(ginHandlers, middlewareToGinHandler(item))
	}

	return ginHandlers
}

func handlerToGinHandler(handler httpcontract.HandlerFunc) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		handler(frameworkhttp.NewGinContext(ginCtx))
	}
}

func middlewareToGinHandler(handler httpcontract.Middleware) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		handler(frameworkhttp.NewGinContext(ginCtx))
	}
}

func getDebugLog() gin.HandlerFunc {
	logFormatter := func(param gin.LogFormatterParams) string {
		var statusColor, methodColor, resetColor string
		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		if param.Latency > time.Minute {
			// Truncate in a golang < 1.8 safe way
			param.Latency = param.Latency - param.Latency%time.Second
		}
		return fmt.Sprintf("[HTTP] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			statusColor, param.StatusCode, resetColor,
			param.Latency,
			param.ClientIP,
			methodColor, param.Method, resetColor,
			param.Path,
			param.ErrorMessage,
		)
	}

	if facades.Config.GetBool("app.debug") {
		return gin.LoggerWithFormatter(logFormatter)
	}

	return nil
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
