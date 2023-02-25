package route

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	frameworkhttp "github.com/goravel/framework/http"
)

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
	compileRegex := regexp.MustCompile(`{(.*?)}`)
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

func runningInConsole() bool {
	args := os.Args

	return len(args) >= 2 && args[1] == "artisan"
}

func outputRoutes(routes gin.RoutesInfo) {
	if facades.Config.GetBool("app.debug") && !runningInConsole() {
		for _, item := range routes {
			fmt.Printf("%-10s %s\n", item.Method, colonToBracket(item.Path))
		}
	}
}
