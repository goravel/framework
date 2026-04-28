package document

import (
	"io"

	"github.com/goravel/framework/ai/attachment"
	contractsai "github.com/goravel/framework/contracts/ai"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
)

var (
	WithFilename = attachment.WithFilename
	WithMimeType = attachment.WithMimeType
)

func FromByte(content []byte, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return attachment.FromBytes(contractsai.AttachmentKindFile, content, resolveOptions(options))
}

func FromString(content string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return attachment.FromString(contractsai.AttachmentKindFile, content, resolveOptions(options))
}

func FromBase64(content string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return attachment.FromBase64(contractsai.AttachmentKindFile, content, resolveOptions(options))
}

func FromReader(reader io.Reader, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return attachment.FromReader(contractsai.AttachmentKindFile, reader, resolveOptions(options))
}

func FromPath(path string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return attachment.FromPath(contractsai.AttachmentKindFile, path, resolveOptions(options))
}

func FromStorage(path string, disk string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	storage := facades.Storage()
	var driver contractsfilesystem.Driver = storage
	if disk != "" {
		driver = storage.Disk(disk)
	}

	return attachment.FromStorage(contractsai.AttachmentKindFile, driver, path, resolveOptions(options))
}

func FromUrl(url string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return attachment.FromUrl(contractsai.AttachmentKindFile, url, resolveOptions(options))
}

func FromUpload(file contractsfilesystem.File, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return attachment.FromUpload(contractsai.AttachmentKindFile, file, resolveOptions(options))
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
