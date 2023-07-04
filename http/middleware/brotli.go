package middleware

import (
	gbrotli "github.com/anargu/gin-brotli"
	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/http"
)

func Brotli() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		switch ctx := ctx.(type) {
		case *http.GinContext:
			newBrotli(gbrotli.Options{
				WriterOptions: brotli.WriterOptions{
					Quality: facades.Config().GetInt("brotli.compression_level"),
					LGWin:   facades.Config().GetInt("brotli.lgwin_level"),
				},
				SkipExtensions: gbrotli.DefaultCompression.SkipExtensions,
			})(ctx.Instance())
		}

		ctx.Request().Next()
	}
}

type brotliWrapper struct {
	options gbrotli.Options
}

func (b brotliWrapper) build() gin.HandlerFunc {
	brotliMiddleware := gbrotli.Brotli(b.options)
	return func(ctx *gin.Context) {
		brotliMiddleware(ctx)
	}
}

func newBrotli(options gbrotli.Options) gin.HandlerFunc {
	return brotliWrapper{options}.build()
}
