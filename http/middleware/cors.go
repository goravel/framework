package middleware

import (
	nethttp "net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

func Cors() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		switch ctx := ctx.(type) {
		case *http.GinContext:
			allowedMethods := facades.Config.Get("cors.allowed_methods").([]string)
			if len(allowedMethods) == 1 && allowedMethods[0] == "*" {
				allowedMethods = []string{nethttp.MethodPost, nethttp.MethodGet, nethttp.MethodOptions, nethttp.MethodPut, nethttp.MethodDelete}
			}

			New(Options{
				AllowedMethods:      allowedMethods,
				AllowedOrigins:      facades.Config.Get("cors.allowed_origins").([]string),
				AllowedHeaders:      facades.Config.Get("cors.allowed_headers").([]string),
				ExposedHeaders:      facades.Config.Get("cors.exposed_headers").([]string),
				MaxAge:              facades.Config.GetInt("cors.max_age"),
				AllowCredentials:    facades.Config.GetBool("cors.supports_credentials"),
				AllowPrivateNetwork: true,
			})(ctx.Instance())
		}

		ctx.Request().Next()
	}
}

// Options is a configuration container to setup the CORS middleware.
type Options = cors.Options

// corsWrapper is a wrapper of cors.Cors handler which preserves information
// about configured 'optionPassthrough' option.
type corsWrapper struct {
	*cors.Cors
	optionPassthrough bool
}

// build transforms wrapped cors.Cors handler into Gin middleware.
func (c corsWrapper) build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c.HandlerFunc(ctx.Writer, ctx.Request)
		if !c.optionPassthrough &&
			ctx.Request.Method == nethttp.MethodOptions &&
			ctx.GetHeader("Access-Control-Request-Method") != "" {
			// Abort processing next Gin middlewares.
			ctx.AbortWithStatus(nethttp.StatusNoContent)
		}
	}
}

// AllowAll creates a new CORS Gin middleware with permissive configuration
// allowing all origins with all standard methods with any header and
// credentials.
func AllowAll() gin.HandlerFunc {
	return corsWrapper{Cors: cors.AllowAll()}.build()
}

// Default creates a new CORS Gin middleware with default options.
func Default() gin.HandlerFunc {
	return corsWrapper{Cors: cors.Default()}.build()
}

// New creates a new CORS Gin middleware with the provided options.
func New(options Options) gin.HandlerFunc {
	return corsWrapper{cors.New(options), options.OptionsPassthrough}.build()
}
