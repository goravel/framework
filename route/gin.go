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
			fmt.Printf("%-10s %s\n", item.Method, item.Path)
		}
	}

	color.Greenln("Listening and serving HTTP on " + addr)

	return r.instance.Run([]string{addr}...)
}

func (r *Gin) Group(handler route.GroupFunc) {
	handler(r.Route)
}

func (r *Gin) Prefix(addr string) route.Engine {
	return &Gin{instance: r.instance, Route: NewGinGroup(r.instance.Group(addr))}
}

type GinGroup struct {
	instance gin.IRoutes
}

func NewGinGroup(routeGroup gin.IRoutes) route.Route {
	return &GinGroup{instance: routeGroup}
}

func (r *GinGroup) Any(relativePath string, handler httpcontract.HandlerFunc) {
	r.instance.Any(BracketToColon(relativePath), []gin.HandlerFunc{HandlerToGinHandler(handler)}...)
}

func (r *GinGroup) Get(relativePath string, handler httpcontract.HandlerFunc) {
	r.instance.GET(BracketToColon(relativePath), []gin.HandlerFunc{HandlerToGinHandler(handler)}...)
}

func (r *GinGroup) Post(relativePath string, handler httpcontract.HandlerFunc) {
	r.instance.POST(BracketToColon(relativePath), []gin.HandlerFunc{HandlerToGinHandler(handler)}...)
}

func (r *GinGroup) Delete(relativePath string, handler httpcontract.HandlerFunc) {
	r.instance.DELETE(BracketToColon(relativePath), []gin.HandlerFunc{HandlerToGinHandler(handler)}...)
}

func (r *GinGroup) Patch(relativePath string, handler httpcontract.HandlerFunc) {
	r.instance.PATCH(BracketToColon(relativePath), []gin.HandlerFunc{HandlerToGinHandler(handler)}...)
}

func (r *GinGroup) Put(relativePath string, handler httpcontract.HandlerFunc) {
	r.instance.PUT(BracketToColon(relativePath), []gin.HandlerFunc{HandlerToGinHandler(handler)}...)
}

func (r *GinGroup) Options(relativePath string, handler httpcontract.HandlerFunc) {
	r.instance.OPTIONS(BracketToColon(relativePath), []gin.HandlerFunc{HandlerToGinHandler(handler)}...)
}

func (r *GinGroup) Middleware(handlers ...httpcontract.Middleware) route.Route {
	var ginHandlers []gin.HandlerFunc
	for _, handler := range handlers {
		ginHandlers = append(ginHandlers, MiddlewareToGinHandler(handler))
	}
	r.instance.Use(ginHandlers...)

	return r
}

func (r *GinGroup) Static(relativePath, root string) {
	r.instance.Static(relativePath, root)
}

func (r *GinGroup) StaticFile(relativePath, filepath string) {
	r.instance.StaticFile(relativePath, filepath)
}

func (r *GinGroup) StaticFS(relativePath string, fs http.FileSystem) {
	r.instance.StaticFS(relativePath, fs)
}

func ColonToBracket(relativePath string) string {
	arr := strings.Split(relativePath, "/")
	for _, item := range arr {
		if strings.HasPrefix(item, ":") {
			item = "{" + strings.ReplaceAll(item, ":", "") + "}"
		}
	}

	path := strings.Join(arr, "/")
	if strings.HasPrefix(relativePath, "/") {
		path = "/" + path
	}
	if strings.HasSuffix(relativePath, "/") {
		path += "/"
	}

	fmt.Print(relativePath, path)

	return path
}

func BracketToColon(relativePath string) string {
	compileRegex := regexp.MustCompile("\\{(.*?)\\}")
	matchArr := compileRegex.FindAllStringSubmatch(relativePath, -1)

	for _, item := range matchArr {
		relativePath = strings.ReplaceAll(relativePath, item[0], ":"+item[1])
	}

	return relativePath
}

func HandlerToGinHandler(handler httpcontract.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		facades.Response = frameworkhttp.NewGinResponse(ctx)
		handler(frameworkhttp.NewGinRequest(ctx))
	}
}

func MiddlewareToGinHandler(handler httpcontract.Middleware) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		facades.Response = frameworkhttp.NewGinResponse(ctx)
		handler(frameworkhttp.NewGinRequest(ctx))
	}
}
