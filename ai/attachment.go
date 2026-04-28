package ai

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/gabriel-vasile/mimetype"

	contractsai "github.com/goravel/framework/contracts/ai"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/errors"
)

type AttachmentOption func(*attachment)

type attachment struct {
	kind     contractsai.AttachmentKind
	filename string
	mimeType string
	resolver func(context.Context) ([]byte, string, string, error)

	once    sync.Once
	content []byte
	err     error
}

func Image(content []byte, options ...AttachmentOption) contractsai.Attachment {
	return newAttachment(contractsai.AttachmentKindImage, func(context.Context) ([]byte, string, string, error) {
		return bytes.Clone(content), "", "", nil
	}, options...)
}

func ImageFromReader(reader io.Reader, options ...AttachmentOption) contractsai.Attachment {
	return newAttachment(contractsai.AttachmentKindImage, func(context.Context) ([]byte, string, string, error) {
		content, err := io.ReadAll(reader)
		return content, "", "", err
	}, options...)
}

func ImageFromPath(path string, options ...AttachmentOption) contractsai.Attachment {
	return newAttachmentFromPath(contractsai.AttachmentKindImage, path, options...)
}

func ImageFromStorage(storage contractsfilesystem.Driver, path string, options ...AttachmentOption) contractsai.Attachment {
	return newAttachmentFromStorage(contractsai.AttachmentKindImage, storage, path, options...)
}

func File(content []byte, options ...AttachmentOption) contractsai.Attachment {
	return newAttachment(contractsai.AttachmentKindFile, func(context.Context) ([]byte, string, string, error) {
		return bytes.Clone(content), "", "", nil
	}, options...)
}

func FileFromReader(reader io.Reader, options ...AttachmentOption) contractsai.Attachment {
	return newAttachment(contractsai.AttachmentKindFile, func(context.Context) ([]byte, string, string, error) {
		content, err := io.ReadAll(reader)
		return content, "", "", err
	}, options...)
}

func FileFromPath(path string, options ...AttachmentOption) contractsai.Attachment {
	return newAttachmentFromPath(contractsai.AttachmentKindFile, path, options...)
}

func FileFromStorage(storage contractsfilesystem.Driver, path string, options ...AttachmentOption) contractsai.Attachment {
	return newAttachmentFromStorage(contractsai.AttachmentKindFile, storage, path, options...)
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

func newAttachmentFromPath(kind contractsai.AttachmentKind, path string, options ...AttachmentOption) contractsai.Attachment {
	return newAttachment(kind, func(context.Context) ([]byte, string, string, error) {
		file, err := os.Open(path)
		if err != nil {
			return nil, "", "", err
		}
		defer errors.Ignore(file.Close)

		content, err := io.ReadAll(file)
		if err != nil {
			return nil, "", "", err
		}

		return content, filepath.Base(path), mimetype.Detect(content).String(), nil
	}, options...)
}

func newAttachmentFromStorage(kind contractsai.AttachmentKind, storage contractsfilesystem.Driver, path string, options ...AttachmentOption) contractsai.Attachment {
	return newAttachment(kind, func(ctx context.Context) ([]byte, string, string, error) {
		if storageWithContext := storage.WithContext(ctx); storageWithContext != nil {
			storage = storageWithContext
		}

		content, err := storage.GetBytes(path)
		if err != nil {
			return nil, "", "", err
		}

		mimeType, err := storage.MimeType(path)
		if err != nil {
			mimeType = mimetype.Detect(content).String()
		}

		return content, filepath.Base(path), mimeType, nil
	}, options...)
}

func newAttachment(kind contractsai.AttachmentKind, resolver func(context.Context) ([]byte, string, string, error), options ...AttachmentOption) contractsai.Attachment {
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
