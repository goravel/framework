package openai

import (
	"bytes"
	"context"

	contractsai "github.com/goravel/framework/contracts/ai"
)

type response struct {
	text      string
	usage     *usage
	toolCalls []contractsai.ToolCall
}

type storedFileResponse struct {
	id string
}

type imageResponse struct {
	mimeType string
	content  []byte
	usage    *usage
}

type fileResponse struct {
	id       string
	mimeType string
	content  []byte
}

func (r *response) Text() string                      { return r.text }
func (r *response) Usage() contractsai.Usage          { return r.usage }
func (r *response) ToolCalls() []contractsai.ToolCall { return r.toolCalls }
func (r *response) Then(callback func(contractsai.Response)) contractsai.Response {
	if callback == nil {
		return r
	}

	callback(r)

	return r
}

func (r *storedFileResponse) ID() string { return r.id }

func (r *imageResponse) Content(context.Context) ([]byte, error) { return bytes.Clone(r.content), nil }

func (r *imageResponse) MimeType() string { return r.mimeType }

func (r *imageResponse) Usage() contractsai.Usage { return r.usage }

func (r *imageResponse) Then(callback func(contractsai.ImageResponse)) contractsai.ImageResponse {
	if callback == nil {
		return r
	}

	callback(r)

	return r
}

func (r *fileResponse) ID() string { return r.id }

func (r *fileResponse) MimeType() string { return r.mimeType }

func (r *fileResponse) Content(context.Context) ([]byte, error) { return bytes.Clone(r.content), nil }

type usage struct{ input, output, total int }

func (r *usage) Input() int  { return r.input }
func (r *usage) Output() int { return r.output }
func (r *usage) Total() int  { return r.total }
