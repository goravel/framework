package attachment

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"mime"
	"net/http"
	urlpkg "net/url"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/gabriel-vasile/mimetype"

	contractsai "github.com/goravel/framework/contracts/ai"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/errors"
)

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

func WithFilename(filename string) contractsai.AttachmentOption {
	return func(options *contractsai.AttachmentOptions) {
		options.Filename = filename
	}
}

func WithMimeType(mimeType string) contractsai.AttachmentOption {
	return func(options *contractsai.AttachmentOptions) {
		options.MimeType = mimeType
	}
}

func New(kind contractsai.AttachmentKind, resolver Resolver, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return &resolved{
		kind:     kind,
		filename: metadata.Filename,
		mimeType: metadata.MimeType,
		resolver: resolver,
	}
}

func FromBytes(kind contractsai.AttachmentKind, content []byte, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return New(kind, func(context.Context) ([]byte, string, string, error) {
		return bytes.Clone(content), "", "", nil
	}, metadata)
}

func FromReader(kind contractsai.AttachmentKind, reader io.Reader, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return New(kind, func(context.Context) ([]byte, string, string, error) {
		content, err := io.ReadAll(reader)
		return content, "", "", err
	}, metadata)
}

func FromString(kind contractsai.AttachmentKind, content string, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return FromBytes(kind, []byte(content), metadata)
}

func FromBase64(kind contractsai.AttachmentKind, content string, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return New(kind, func(context.Context) ([]byte, string, string, error) {
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, "", "", err
		}

		return decoded, "", "", nil
	}, metadata)
}

func FromPath(kind contractsai.AttachmentKind, path string, metadata contractsai.AttachmentOptions) contractsai.Attachment {
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

func FromStorage(kind contractsai.AttachmentKind, storage contractsfilesystem.Driver, path string, metadata contractsai.AttachmentOptions) contractsai.Attachment {
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

func FromUrl(kind contractsai.AttachmentKind, rawURL string, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return New(kind, func(ctx context.Context) ([]byte, string, string, error) {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, "", "", err
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, "", "", err
		}
		defer errors.Ignore(response.Body.Close)

		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return nil, "", "", errors.AIAttachmentUrlResponseNotOK.Args(response.StatusCode)
		}

		content, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, "", "", err
		}

		mimeType := response.Header.Get("Content-Type")
		if parsedMimeType, _, err := mime.ParseMediaType(mimeType); err == nil {
			mimeType = parsedMimeType
		}

		return content, resolveURLFilename(rawURL), mimeType, nil
	}, metadata)
}

func FromUpload(kind contractsai.AttachmentKind, file contractsfilesystem.File, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return New(kind, func(context.Context) ([]byte, string, string, error) {
		content, err := os.ReadFile(file.File())
		if err != nil {
			return nil, "", "", err
		}

		mimeType, err := file.MimeType()
		if err != nil {
			mimeType = mimetype.Detect(content).String()
		}

		return content, file.GetClientOriginalName(), mimeType, nil
	}, metadata)
}

func resolveURLFilename(rawURL string) string {
	parsedURL, err := urlpkg.Parse(rawURL)
	if err != nil {
		return ""
	}

	filename := path.Base(parsedURL.Path)
	if filename == "." || filename == "/" {
		return ""
	}

	return filename
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
