package access

import "context"

//go:generate mockery --name=Gate
type Gate interface {
	WithContext(ctx context.Context) Gate
	Allows(ability string, arguments map[string]any) bool
	Denies(ability string, arguments map[string]any) bool
	Inspect(ability string, arguments map[string]any) Response
	Define(ability string, callback func(ctx context.Context, arguments map[string]any) Response)
	Any(abilities []string, arguments map[string]any) bool
	None(abilities []string, arguments map[string]any) bool
	Before(callback func(ctx context.Context, ability string, arguments map[string]any) Response)
	After(callback func(ctx context.Context, ability string, arguments map[string]any, result Response) Response)
}

type Response interface {
	Allowed() bool
	Message() string
}

func NewAllowResponse() Response {
	return &ResponseImpl{allowed: true}
}

func NewDenyResponse(message string) Response {
	return &ResponseImpl{message: message}
}
