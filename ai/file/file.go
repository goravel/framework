package file

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
	"time"

	"github.com/gabriel-vasile/mimetype"

	contractsai "github.com/goravel/framework/contracts/ai"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/facades"
)

type resolver func(context.Context) ([]byte, string, string, error)

var attachmentHTTPClient = &http.Client{Timeout: 30 * time.Second}

const attachmentURLMaxBytes = 20 << 20

type resolved struct {
	kind     contractsai.AttachmentKind
	filename string
	mimeType string
	resolver resolver

	once    sync.Once
	content []byte
	err     error
}

func WithMimeType(mimeType string) contractsai.AttachmentOption {
	return func(options *contractsai.AttachmentOptions) {
		options.MimeType = mimeType
	}
}

func WithDisk(disk string) contractsai.AttachmentOption {
	return func(options *contractsai.AttachmentOptions) {
		options.Disk = disk
	}
}

func DocumentFromByte(content []byte, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromBytes(contractsai.AttachmentKindFile, content, resolveOptions(options))
}

func DocumentFromString(content string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromString(contractsai.AttachmentKindFile, content, resolveOptions(options))
}

func DocumentFromBase64(content string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromBase64(contractsai.AttachmentKindFile, content, resolveOptions(options))
}

func DocumentFromReader(reader io.Reader, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromReader(contractsai.AttachmentKindFile, reader, resolveOptions(options))
}

func DocumentFromPath(path string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromPath(contractsai.AttachmentKindFile, path, resolveOptions(options))
}

func DocumentFromStorage(path string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromStorage(contractsai.AttachmentKindFile, path, resolveOptions(options))
}

func DocumentFromURL(rawURL string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromURL(contractsai.AttachmentKindFile, rawURL, resolveOptions(options))
}

func DocumentFromUpload(file contractsfilesystem.File, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromUpload(contractsai.AttachmentKindFile, file, resolveOptions(options))
}

func ImageFromByte(content []byte, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromBytes(contractsai.AttachmentKindImage, content, resolveOptions(options))
}

func ImageFromBase64(content string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromBase64(contractsai.AttachmentKindImage, content, resolveOptions(options))
}

func ImageFromReader(reader io.Reader, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromReader(contractsai.AttachmentKindImage, reader, resolveOptions(options))
}

func ImageFromPath(path string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromPath(contractsai.AttachmentKindImage, path, resolveOptions(options))
}

func ImageFromStorage(path string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromStorage(contractsai.AttachmentKindImage, path, resolveOptions(options))
}

func ImageFromURL(rawURL string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromURL(contractsai.AttachmentKindImage, rawURL, resolveOptions(options))
}

func ImageFromUpload(file contractsfilesystem.File, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromUpload(contractsai.AttachmentKindImage, file, resolveOptions(options))
}

func resolveOptions(options []contractsai.AttachmentOption) contractsai.AttachmentOptions {
	metadata := contractsai.AttachmentOptions{}
	for _, option := range options {
		if option != nil {
			option(&metadata)
		}
	}

	return metadata
}

func newAttachment(kind contractsai.AttachmentKind, resolver resolver, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return &resolved{
		kind:     kind,
		mimeType: metadata.MimeType,
		resolver: resolver,
	}
}

func fromBytes(kind contractsai.AttachmentKind, content []byte, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return newAttachment(kind, func(context.Context) ([]byte, string, string, error) {
		return bytes.Clone(content), "", "", nil
	}, metadata)
}

func fromReader(kind contractsai.AttachmentKind, reader io.Reader, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return newAttachment(kind, func(context.Context) ([]byte, string, string, error) {
		content, err := io.ReadAll(reader)
		return content, "", "", err
	}, metadata)
}

func fromString(kind contractsai.AttachmentKind, content string, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return fromBytes(kind, []byte(content), metadata)
}

func fromBase64(kind contractsai.AttachmentKind, content string, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return newAttachment(kind, func(context.Context) ([]byte, string, string, error) {
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, "", "", err
		}

		return decoded, "", "", nil
	}, metadata)
}

func fromPath(kind contractsai.AttachmentKind, path string, metadata contractsai.AttachmentOptions) contractsai.Attachment {
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
	}, metadata)
}

func fromStorage(kind contractsai.AttachmentKind, path string, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return newAttachment(kind, func(ctx context.Context) ([]byte, string, string, error) {
		storage := facades.Storage()
		var driver contractsfilesystem.Driver = storage
		if metadata.Disk != "" {
			driver = storage.Disk(metadata.Disk)
		}
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

func fromURL(kind contractsai.AttachmentKind, rawURL string, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return newAttachment(kind, func(ctx context.Context) ([]byte, string, string, error) {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, "", "", err
		}

		response, err := attachmentHTTPClient.Do(request)
		if err != nil {
			return nil, "", "", err
		}
		defer errors.Ignore(response.Body.Close)

		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return nil, "", "", errors.AIAttachmentUrlResponseNotOK.Args(response.StatusCode)
		}
		if response.ContentLength > attachmentURLMaxBytes {
			return nil, "", "", errors.AIAttachmentUrlResponseTooLarge.Args(attachmentURLMaxBytes)
		}

		content, err := io.ReadAll(io.LimitReader(response.Body, attachmentURLMaxBytes+1))
		if err != nil {
			return nil, "", "", err
		}
		if int64(len(content)) > attachmentURLMaxBytes {
			return nil, "", "", errors.AIAttachmentUrlResponseTooLarge.Args(attachmentURLMaxBytes)
		}

		mimeType := response.Header.Get("Content-Type")
		if parsedMimeType, _, err := mime.ParseMediaType(mimeType); err == nil {
			mimeType = parsedMimeType
		}

		return content, resolveURLFilename(rawURL), mimeType, nil
	}, metadata)
}

func fromUpload(kind contractsai.AttachmentKind, file contractsfilesystem.File, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return newAttachment(kind, func(context.Context) ([]byte, string, string, error) {
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

func (r *resolved) FileName() string { return r.filename }

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
