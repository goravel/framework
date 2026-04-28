package document

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
	return sharedattachment.FromBytes(contractsai.AttachmentKindFile, content, resolveMetadata(options))
}

func FromReader(reader io.Reader, options ...Option) contractsai.Attachment {
	return sharedattachment.FromReader(contractsai.AttachmentKindFile, reader, resolveMetadata(options))
}

func FromPath(path string, options ...Option) contractsai.Attachment {
	return sharedattachment.FromPath(contractsai.AttachmentKindFile, path, resolveMetadata(options))
}

func FromStorage(storage contractsfilesystem.Driver, path string, options ...Option) contractsai.Attachment {
	return sharedattachment.FromStorage(contractsai.AttachmentKindFile, storage, path, resolveMetadata(options))
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
