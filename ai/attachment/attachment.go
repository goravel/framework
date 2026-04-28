package attachment

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

type Metadata struct {
	Filename string
	MimeType string
}

type Resolver func(context.Context) ([]byte, string, string, error)

type resolved struct {
	kind     contractsai.AttachmentKind
	filename string
	mimeType string
	resolver Resolver

	once    sync.Once
	content []byte
	err     error
}

func New(kind contractsai.AttachmentKind, resolver Resolver, metadata Metadata) contractsai.Attachment {
	return &resolved{
		kind:     kind,
		filename: metadata.Filename,
		mimeType: metadata.MimeType,
		resolver: resolver,
	}
}

func FromBytes(kind contractsai.AttachmentKind, content []byte, metadata Metadata) contractsai.Attachment {
	return New(kind, func(context.Context) ([]byte, string, string, error) {
		return bytes.Clone(content), "", "", nil
	}, metadata)
}

func FromReader(kind contractsai.AttachmentKind, reader io.Reader, metadata Metadata) contractsai.Attachment {
	return New(kind, func(context.Context) ([]byte, string, string, error) {
		content, err := io.ReadAll(reader)
		return content, "", "", err
	}, metadata)
}

func FromPath(kind contractsai.AttachmentKind, path string, metadata Metadata) contractsai.Attachment {
	return New(kind, func(context.Context) ([]byte, string, string, error) {
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
	}, metadata)
}

func FromStorage(kind contractsai.AttachmentKind, storage contractsfilesystem.Driver, path string, metadata Metadata) contractsai.Attachment {
	return New(kind, func(ctx context.Context) ([]byte, string, string, error) {
		driver := storage
		if storageWithContext := driver.WithContext(ctx); storageWithContext != nil {
			driver = storageWithContext
		}

		content, err := driver.GetBytes(path)
		if err != nil {
			return nil, "", "", err
		}

		mimeType, err := driver.MimeType(path)
		if err != nil {
			mimeType = mimetype.Detect(content).String()
		}

		return content, filepath.Base(path), mimeType, nil
	}, metadata)
}

func (r *resolved) Kind() contractsai.AttachmentKind { return r.kind }

func (r *resolved) Filename() string { return r.filename }

func (r *resolved) MimeType() string { return r.mimeType }

func (r *resolved) Content(ctx context.Context) ([]byte, error) {
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
