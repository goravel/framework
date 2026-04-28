package image

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"

	frameworkai "github.com/goravel/framework/ai"
	contractsai "github.com/goravel/framework/contracts/ai"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/errors"
)

func New(content []byte, options ...frameworkai.AttachmentOption) contractsai.Attachment {
	return frameworkai.NewAttachment(contractsai.AttachmentKindImage, func(context.Context) ([]byte, string, string, error) {
		return bytes.Clone(content), "", "", nil
	}, options...)
}

func FromReader(reader io.Reader, options ...frameworkai.AttachmentOption) contractsai.Attachment {
	return frameworkai.NewAttachment(contractsai.AttachmentKindImage, func(context.Context) ([]byte, string, string, error) {
		content, err := io.ReadAll(reader)
		return content, "", "", err
	}, options...)
}

func FromPath(path string, options ...frameworkai.AttachmentOption) contractsai.Attachment {
	return frameworkai.NewAttachment(contractsai.AttachmentKindImage, func(context.Context) ([]byte, string, string, error) {
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

func FromStorage(storage contractsfilesystem.Driver, path string, options ...frameworkai.AttachmentOption) contractsai.Attachment {
	return frameworkai.NewAttachment(contractsai.AttachmentKindImage, func(ctx context.Context) ([]byte, string, string, error) {
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
