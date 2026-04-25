package ai

import (
	"reflect"

	contractsai "github.com/goravel/framework/contracts/ai"
)

func WithProvider(provider string) contractsai.Option {
	return func(options *contractsai.Options) {
		options.Provider = provider
	}
}

func WithModel(model string) contractsai.Option {
	return func(options *contractsai.Options) {
		options.Model = model
	}
}

func WithMiddleware(middlewares ...contractsai.Middleware) contractsai.Option {
	return func(options *contractsai.Options) {
		options.Middlewares = append(options.Middlewares, filterNilMiddlewares(middlewares)...)
	}
}

func filterNilMiddlewares(middlewares []contractsai.Middleware) []contractsai.Middleware {
	filtered := make([]contractsai.Middleware, 0, len(middlewares))

	for _, middleware := range middlewares {
		if isNilMiddleware(middleware) {
			continue
		}

		filtered = append(filtered, middleware)
	}

	return filtered
}

func isNilMiddleware(middleware contractsai.Middleware) bool {
	if middleware == nil {
		return true
	}

	value := reflect.ValueOf(middleware)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func WithStreamCode(code int) contractsai.StreamOption {
	return func(options *contractsai.StreamOptions) {
		options.Code = code
	}
}

func WithStreamRender(render contractsai.RenderFunc) contractsai.StreamOption {
	return func(options *contractsai.StreamOptions) {
		options.Render = render
	}
}
