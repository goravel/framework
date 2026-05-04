package openai

import (
	"bytes"
	"context"

	frameworkai "github.com/goravel/framework/ai"
	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/support/str"
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
	name     string
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

func (r *imageResponse) Content() ([]byte, error) { return bytes.Clone(r.content), nil }

func (r *imageResponse) MimeType() string { return r.mimeType }

func (r *imageResponse) Store(disk ...string) (string, error) {
	return frameworkai.StoreImage(r.content, r.storageName(), disk...)
}

func (r *imageResponse) StoreAs(path string, disk ...string) (string, error) {
	return frameworkai.StoreImageContentAs(r.content, path, disk...)
}

func (r *imageResponse) Usage() contractsai.Usage { return r.usage }

func (r *imageResponse) Then(callback func(contractsai.ImageResponse)) contractsai.ImageResponse {
	if callback == nil {
		return r
	}

	callback(r)

	return r
}

func (r *imageResponse) storageName() string {
	if r.name != "" {
		return r.name
	}

	extension := ".png"
	switch r.mimeType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/webp":
		extension = ".webp"
	}

	r.name = str.Random(40) + extension

	return r.name
}

func (r *fileResponse) ID() string { return r.id }

func (r *fileResponse) MimeType() string { return r.mimeType }

func (r *fileResponse) Content(context.Context) ([]byte, error) { return bytes.Clone(r.content), nil }

type usage struct{ input, output, total int }

func (r *usage) Input() int  { return r.input }
func (r *usage) Output() int { return r.output }
func (r *usage) Total() int  { return r.total }
