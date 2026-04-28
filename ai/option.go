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

func WithAttachments(attachments ...contractsai.Attachment) contractsai.ConversationOption {
	return func(options *contractsai.ConversationOptions) {
		options.Attachments = append(options.Attachments, filterNilAttachments(attachments)...)
	}
}

func WithAttachment(attachment contractsai.Attachment) contractsai.ConversationOption {
	return WithAttachments(attachment)
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

func filterNilAttachments(attachments []contractsai.Attachment) []contractsai.Attachment {
	if len(attachments) == 0 {
		return nil
	}

	filtered := make([]contractsai.Attachment, 0, len(attachments))

	for _, attachment := range attachments {
		if isNilAttachment(attachment) {
			continue
		}

		filtered = append(filtered, attachment)
	}

	if len(filtered) == 0 {
		return nil
	}

	return filtered
}

func isNilMiddleware(middleware contractsai.Middleware) bool {
	return isNilInterface(middleware)
}

func isNilAttachment(attachment contractsai.Attachment) bool {
	return isNilInterface(attachment)
}

func isNilInterface(value any) bool {
	if value == nil {
		return true
	}

	reflectValue := reflect.ValueOf(value)
	switch reflectValue.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflectValue.IsNil()
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
