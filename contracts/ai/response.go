package ai

import contractshttp "github.com/goravel/framework/contracts/http"

// AgentResponse exposes generated text and provider metadata.
type AgentResponse interface {
	Text() string
	Usage() Usage
	// ToolCalls returns any tool invocations the model requested.
	// Returns nil or an empty slice when the model returns plain text.
	ToolCalls() []ToolCall
	// Then registers a callback against the resolved response.
	Then(callback func(AgentResponse)) AgentResponse
}

type StreamableAgentResponse interface {
	Each(callback func(StreamEvent) error) error
	Then(callback func(AgentResponse)) StreamableAgentResponse
	HTTPResponse(ctx contractshttp.Context, options ...StreamOption) contractshttp.Response
}

// Usage contains token statistics for a response.
type Usage interface {
	Input() int
	Output() int
	Total() int
}

// ImageResponse exposes generated image bytes and provider metadata.
type ImageResponse interface {
	Content() ([]byte, error)
	MimeType() string
	Store(disk ...string) (string, error)
	StoreAs(path string, disk ...string) (string, error)
	Usage() Usage
	Then(callback func(ImageResponse)) ImageResponse
}

// AudioResponse exposes generated audio bytes and provider metadata.
type AudioResponse interface {
	Content() ([]byte, error)
	MimeType() string
	Store(disk ...string) (string, error)
	StoreAs(path string, disk ...string) (string, error)
	Usage() Usage
	Then(callback func(AudioResponse)) AudioResponse
}
