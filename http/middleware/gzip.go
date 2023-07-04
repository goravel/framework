package middleware

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/http"
)

func Gzip() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		switch ctx := ctx.(type) {
		case *http.GinContext:
			newGzip(facades.Config().GetInt("gzip.compression_level"))(ctx.Instance())
		}

		ctx.Request().Next()
	}
}

type gzipWrapper struct {
	compressionLevel int
	options          []gzip.Option
}

func (g gzipWrapper) build() gin.HandlerFunc {
	gzipMiddleware := gzip.Gzip(g.compressionLevel, g.options...)
	return func(ctx *gin.Context) {
		gzipMiddleware(ctx)
	}
}

func newGzip(compressionLevel int, options ...gzip.Option) gin.HandlerFunc {
	return gzipWrapper{compressionLevel, options}.build()
}
