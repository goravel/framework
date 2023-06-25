package middleware

import (
	"errors"
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/spf13/cast"
)

func Recover(config ...Config) httpcontract.Middleware {

	// Set default config
	cfg := configDefault(config...)

	return func(ctx httpcontract.Context) {

		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(ctx) {
			ctx.Request().Next()
			return
		}

		// Catch panics
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					// Set error that will call the global error handler
					ctx.Request().AbortWithError(err.(error))
					return
				}
				if cast.ToString(r) != "" {
					// Set error that will call the global error handler
					ctx.Request().AbortWithError(errors.New(cast.ToString(r)))
					return
				}

				// Set error that will call the global error handler
				ctx.Request().AbortWithError(errors.New("Unknown error"))
				return
			}
		}()

		// Return err if exist, else move to next handler
		ctx.Request().Next()
	}
}

// Config defines the config for middleware.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(c httpcontract.Context) bool
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	Next: nil,
}

// Helper function to set default values
func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		return ConfigDefault
	}

	// Override default config
	cfg := config[0]

	return cfg
}
