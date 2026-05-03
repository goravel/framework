package document

import (
	"io"

	frameworkai "github.com/goravel/framework/ai"
	contractsai "github.com/goravel/framework/contracts/ai"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
)

func WithMimeType(mimeType string) contractsai.AttachmentOption {
	return frameworkai.WithMimeType(mimeType)
}

func WithDisk(disk string) contractsai.AttachmentOption {
	return frameworkai.WithDisk(disk)
}

func FromByte(content []byte, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.DocumentFromByte(content, options...)
}

func FromString(content string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.DocumentFromString(content, options...)
}

func FromBase64(content string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.DocumentFromBase64(content, options...)
}

func FromReader(reader io.Reader, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.DocumentFromReader(reader, options...)
}

func FromPath(path string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.DocumentFromPath(path, options...)
}

func FromStorage(path string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.DocumentFromStorage(path, options...)
}

func FromURL(rawURL string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.DocumentFromURL(rawURL, options...)
}

func FromUpload(file contractsfilesystem.File, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.DocumentFromUpload(file, options...)
}

func FromID(id string) contractsai.ProviderFile {
	return frameworkai.DocumentFromID(id)
}
