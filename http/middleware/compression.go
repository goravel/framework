package middleware

import (
	"strings"

	gbrotli "github.com/anargu/gin-brotli"
	"github.com/andybalholm/brotli"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/http"
)

func newGzipMiddleware() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression) // http.ConfigFacade.GetInt("compression.encoders.gzip.compression_level"))
}

func newBrotliMiddleware() gin.HandlerFunc {
	return gbrotli.Brotli(gbrotli.Options{
		WriterOptions: brotli.WriterOptions{
			Quality: brotli.DefaultCompression, // http.ConfigFacade.GetInt("compression.encoders.compression_level"),
			LGWin:   gbrotli.DefaultCompression.LGWin,
		},
		SkipExtensions: gbrotli.DefaultCompression.SkipExtensions,
	})
}

func prefersBrotli(r httpcontract.Request) bool {
	encodings := strings.Split(r.Header("Accept-Encoding"), ",")
	for _, e := range encodings {
		if e == "br" {
			return true
		}
	}

	return false
}

func Compression() httpcontract.Middleware {
	gzipMiddleware := newGzipMiddleware()
	brotliMiddleware := newBrotliMiddleware()
	return func(ctx httpcontract.Context) {
		switch ctx := ctx.(type) {
		case *http.GinContext:
			instance := ctx.Instance()
			if prefersBrotli(ctx.Request()) {
				brotliMiddleware(instance)
			} else {
				gzipMiddleware(instance)
			}
		}

		ctx.Request().Next()
	}
}
