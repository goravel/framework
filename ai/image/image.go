package image

import (
	"io"

	sharedattachment "github.com/goravel/framework/ai/attachment"
	contractsai "github.com/goravel/framework/contracts/ai"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
)

type Option func(*sharedattachment.Metadata)

func WithFilename(filename string) Option {
	return func(metadata *sharedattachment.Metadata) {
		metadata.Filename = filename
	}
}

func WithMimeType(mimeType string) Option {
	return func(metadata *sharedattachment.Metadata) {
		metadata.MimeType = mimeType
	}
}

func New(content []byte, options ...Option) contractsai.Attachment {
	return sharedattachment.FromBytes(contractsai.AttachmentKindImage, content, resolveMetadata(options))
}

func FromReader(reader io.Reader, options ...Option) contractsai.Attachment {
	return sharedattachment.FromReader(contractsai.AttachmentKindImage, reader, resolveMetadata(options))
}

func FromPath(path string, options ...Option) contractsai.Attachment {
	return sharedattachment.FromPath(contractsai.AttachmentKindImage, path, resolveMetadata(options))
}

func FromStorage(storage contractsfilesystem.Driver, path string, options ...Option) contractsai.Attachment {
	return sharedattachment.FromStorage(contractsai.AttachmentKindImage, storage, path, resolveMetadata(options))
}

func resolveMetadata(options []Option) sharedattachment.Metadata {
	metadata := sharedattachment.Metadata{}
	for _, option := range options {
		if option != nil {
			option(&metadata)
		}
	}

	return metadata
}
