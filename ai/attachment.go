package ai

import (
	"bytes"
	"context"
	"sync"

	"github.com/gabriel-vasile/mimetype"

	contractsai "github.com/goravel/framework/contracts/ai"
)

type AttachmentOption func(*attachment)

type AttachmentResolver func(context.Context) ([]byte, string, string, error)

type attachment struct {
	kind     contractsai.AttachmentKind
	filename string
	mimeType string
	resolver AttachmentResolver

	once    sync.Once
	content []byte
	err     error
}

func WithFilename(filename string) AttachmentOption {
	return func(attachment *attachment) {
		attachment.filename = filename
	}
}

func WithMimeType(mimeType string) AttachmentOption {
	return func(attachment *attachment) {
		attachment.mimeType = mimeType
	}
}

func NewAttachment(kind contractsai.AttachmentKind, resolver AttachmentResolver, options ...AttachmentOption) contractsai.Attachment {
	attachment := &attachment{kind: kind, resolver: resolver}
	for _, option := range options {
		option(attachment)
	}

	return attachment
}

func (r *attachment) Kind() contractsai.AttachmentKind { return r.kind }

func (r *attachment) Filename() string { return r.filename }

func (r *attachment) MimeType() string { return r.mimeType }

func (r *attachment) Content(ctx context.Context) ([]byte, error) {
	r.once.Do(func() {
		content, filename, mimeType, err := r.resolver(ctx)
		if err != nil {
			r.err = err
			return
		}

		r.content = content
		if r.filename == "" {
			r.filename = filename
		}
		if r.mimeType == "" {
			r.mimeType = mimeType
		}
		if r.mimeType == "" {
			r.mimeType = mimetype.Detect(r.content).String()
		}
	})
	if r.err != nil {
		return nil, r.err
	}

	return bytes.Clone(r.content), nil
}
