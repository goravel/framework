package ai

import contractshttp "github.com/goravel/framework/contracts/http"

type StreamEventType string

const (
	StreamEventTypeTextDelta StreamEventType = "text_delta"
	StreamEventTypeToolCall  StreamEventType = "tool_call"
	StreamEventTypeDone      StreamEventType = "done"
	StreamEventTypeError     StreamEventType = "error"
)

type StreamEvent struct {
	Type      StreamEventType
	Delta     string
	Usage     Usage
	Error     string
	ToolCalls []ToolCall
}

type RenderFunc func(w contractshttp.StreamWriter, event StreamEvent) error

type StreamOption func(options *StreamOptions)

type StreamOptions struct {
	Code   int
	Render RenderFunc
}

type StreamableResponse interface {
	Each(callback func(StreamEvent) error) error
	Then(callback func(Response)) StreamableResponse
	HTTPResponse(ctx contractshttp.Context, options ...StreamOption) contractshttp.Response
}
