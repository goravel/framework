package ai

import (
	"bytes"
	"context"
	"strings"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/str"
)

type textResponse struct {
	text      string
	usage     contractsai.Usage
	toolCalls []contractsai.ToolCall
}

type imageResponse struct {
	mimeType string
	content  []byte
	usage    contractsai.Usage
	name     string
	storer   contractsai.ImageStorer
}

type audioResponse struct {
	mimeType string
	content  []byte
	usage    contractsai.Usage
	name     string
	storer   audioStorer
}

type fileResponse struct {
	id       string
	mimeType string
	content  []byte
}

type usage struct{ input, output, total int }

func NewTextResponse(text string, usage contractsai.Usage, toolCalls []contractsai.ToolCall) contractsai.AgentResponse {
	return &textResponse{text: text, usage: usage, toolCalls: toolCalls}
}

func NewImageResponse(content []byte, mimeType string, usage contractsai.Usage) contractsai.ImageResponse {
	storer := NewImageStorer()
	return &imageResponse{content: content, mimeType: mimeType, usage: usage, storer: storer}
}

func NewAudioResponse(content []byte, mimeType string, usage contractsai.Usage) contractsai.AudioResponse {
	return &audioResponse{content: content, mimeType: mimeType, usage: usage, storer: audioStorer{}}
}

func NewFileResponse(id, mimeType string, content []byte) contractsai.FileResponse {
	return &fileResponse{id: id, mimeType: mimeType, content: content}
}

func NewUsage(input, output, total int) contractsai.Usage {
	return &usage{input: input, output: output, total: total}
}

func (r *textResponse) Text() string                      { return r.text }
func (r *textResponse) Usage() contractsai.Usage          { return r.usage }
func (r *textResponse) ToolCalls() []contractsai.ToolCall { return r.toolCalls }
func (r *textResponse) Then(callback func(contractsai.AgentResponse)) contractsai.AgentResponse {
	if callback == nil {
		return r
	}

	callback(r)

	return r
}

func (r *imageResponse) Content() ([]byte, error) { return bytes.Clone(r.content), nil }

func (r *imageResponse) MimeType() string { return r.mimeType }

func (r *imageResponse) Store(disk ...string) (string, error) {
	resolvedDisk, err := resolveImageStoreDisk(disk)
	if err != nil {
		return "", err
	}

	return r.storer.Store(r.content, r.storageName(), resolvedDisk)
}

func (r *imageResponse) StoreAs(path string, disk ...string) (string, error) {
	resolvedDisk, err := resolveImageStoreDisk(disk)
	if err != nil {
		return "", err
	}

	return r.storer.StoreAs(r.content, path, resolvedDisk)
}

func (r *imageResponse) Usage() contractsai.Usage { return r.usage }

func (r *imageResponse) Then(callback func(contractsai.ImageResponse)) contractsai.ImageResponse {
	if callback == nil {
		return r
	}

	callback(r)

	return r
}

func (r *audioResponse) Content() ([]byte, error) { return bytes.Clone(r.content), nil }

func (r *audioResponse) MimeType() string { return r.mimeType }

func (r *audioResponse) Store(disk ...string) (string, error) {
	resolvedDisk, err := resolveAudioStoreDisk(disk)
	if err != nil {
		return "", err
	}

	return r.storer.Store(r.content, r.storageName(), resolvedDisk)
}

func (r *audioResponse) StoreAs(path string, disk ...string) (string, error) {
	resolvedDisk, err := resolveAudioStoreDisk(disk)
	if err != nil {
		return "", err
	}

	return r.storer.StoreAs(r.content, path, resolvedDisk)
}

func (r *audioResponse) Usage() contractsai.Usage { return r.usage }

func (r *audioResponse) Then(callback func(contractsai.AudioResponse)) contractsai.AudioResponse {
	if callback == nil {
		return r
	}

	callback(r)

	return r
}

func (r *audioResponse) storageName() string {
	if r.name != "" {
		return r.name
	}

	extension := ".mp3"
	switch strings.ToLower(r.mimeType) {
	case "audio/wav", "audio/x-wav":
		extension = ".wav"
	case "audio/flac":
		extension = ".flac"
	case "audio/aac":
		extension = ".aac"
	case "audio/opus", "audio/ogg":
		extension = ".opus"
	case "audio/l16", "audio/pcm", "audio/x-pcm":
		extension = ".pcm"
	}

	r.name = str.Random(40) + extension

	return r.name
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

func resolveImageStoreDisk(disk []string) (string, error) {
	if len(disk) > 1 {
		return "", errors.AIImageStoreTooManyPaths
	}

	if len(disk) == 0 {
		return "", nil
	}

	return disk[0], nil
}

func resolveAudioStoreDisk(disk []string) (string, error) {
	if len(disk) > 1 {
		return "", errors.AIAudioStoreTooManyPaths
	}

	if len(disk) == 0 {
		return "", nil
	}

	return disk[0], nil
}

func (r *fileResponse) ID() string { return r.id }

func (r *fileResponse) MimeType() string { return r.mimeType }

func (r *fileResponse) Content(context.Context) ([]byte, error) { return bytes.Clone(r.content), nil }

func (r *usage) Input() int  { return r.input }
func (r *usage) Output() int { return r.output }
func (r *usage) Total() int  { return r.total }
