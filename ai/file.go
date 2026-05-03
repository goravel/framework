package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"mime"
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

type resolver func(context.Context) ([]byte, string, string, error)

type fileUploader interface {
	putFile(ctx context.Context, file contractsai.StorableFile, options ...contractsai.Option) (contractsai.StoredFileResponse, error)
}

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
	return fromBytes(contractsai.AttachmentKindFile, content, resolveAttachmentOptions(options))
}

func DocumentFromString(content string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromString(contractsai.AttachmentKindFile, content, resolveAttachmentOptions(options))
}

func DocumentFromBase64(content string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromBase64(contractsai.AttachmentKindFile, content, resolveAttachmentOptions(options))
}

func DocumentFromReader(reader io.Reader, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromReader(contractsai.AttachmentKindFile, reader, resolveAttachmentOptions(options))
}

func DocumentFromPath(path string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromPath(contractsai.AttachmentKindFile, path, resolveAttachmentOptions(options))
}

func DocumentFromStorage(path string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromStorage(contractsai.AttachmentKindFile, path, resolveAttachmentOptions(options))
}

func DocumentFromURL(rawURL string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromURL(contractsai.AttachmentKindFile, rawURL, resolveAttachmentOptions(options))
}

func DocumentFromUpload(file contractsfilesystem.File, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromUpload(contractsai.AttachmentKindFile, file, resolveAttachmentOptions(options))
}

func ImageFromByte(content []byte, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromBytes(contractsai.AttachmentKindImage, content, resolveAttachmentOptions(options))
}

func ImageFromBase64(content string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromBase64(contractsai.AttachmentKindImage, content, resolveAttachmentOptions(options))
}

func ImageFromReader(reader io.Reader, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromReader(contractsai.AttachmentKindImage, reader, resolveAttachmentOptions(options))
}

func ImageFromPath(path string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromPath(contractsai.AttachmentKindImage, path, resolveAttachmentOptions(options))
}

func ImageFromStorage(path string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromStorage(contractsai.AttachmentKindImage, path, resolveAttachmentOptions(options))
}

func ImageFromURL(rawURL string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromURL(contractsai.AttachmentKindImage, rawURL, resolveAttachmentOptions(options))
}

func ImageFromUpload(file contractsfilesystem.File, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return fromUpload(contractsai.AttachmentKindImage, file, resolveAttachmentOptions(options))
}

func resolveAttachmentOptions(options []contractsai.AttachmentOption) contractsai.AttachmentOptions {
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
		if storageFacade == nil {
			return nil, "", "", errors.StorageFacadeNotSet
		}

		var driver contractsfilesystem.Driver = storageFacade
		if metadata.Disk != "" {
			driver = storageFacade.Disk(metadata.Disk)
		}
		driver = driver.WithContext(ctx)

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
		if httpFacade == nil {
			return nil, "", "", errors.HttpFacadeNotSet
		}

		response, err := httpFacade.WithContext(ctx).Get(rawURL)
		if err != nil {
			return nil, "", "", err
		}

		if !response.Successful() {
			return nil, "", "", errors.AIAttachmentUrlResponseNotOK.Args(response.Status())
		}

		stream, err := response.Stream()
		if err != nil {
			return nil, "", "", err
		}
		defer errors.Ignore(stream.Close)

		content, err := io.ReadAll(stream)
		if err != nil {
			return nil, "", "", err
		}

		mimeType := response.Header("Content-Type")
		if parsedMimeType, _, err := mime.ParseMediaType(mimeType); err == nil {
			mimeType = parsedMimeType
		}

		return content, resolveURLFilename(rawURL), mimeType, nil
	}, metadata)
}

func fromUpload(kind contractsai.AttachmentKind, file contractsfilesystem.File, metadata contractsai.AttachmentOptions) contractsai.Attachment {
	return newAttachment(kind, func(context.Context) ([]byte, string, string, error) {
		path := file.File()
		opened, err := os.Open(path)
		if err != nil {
			return nil, "", "", err
		}
		defer errors.Ignore(opened.Close)

		content, err := io.ReadAll(opened)
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

// PutFile uploads a file through the configured AI provider and returns the stored provider file reference.
func PutFile(ctx context.Context, file contractsai.StorableFile, options ...contractsai.Option) (contractsai.StoredFileResponse, error) {
	if fileUploaderFacade == nil {
		return nil, errors.AIFacadeNotSet
	}

	return fileUploaderFacade.putFile(ctx, file, options...)
}

func (r *resolved) Put(ctx context.Context, options ...contractsai.Option) (contractsai.StoredFileResponse, error) {
	return PutFile(ctx, r, options...)
}

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
