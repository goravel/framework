package image

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

func WithTitle(title string) contractsai.AttachmentOption {
	return frameworkai.WithTitle(title)
}

func FromByte(content []byte, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.ImageFromByte(content, options...)
}

func FromBase64(content string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.ImageFromBase64(content, options...)
}

func FromReader(reader io.Reader, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.ImageFromReader(reader, options...)
}

func FromPath(path string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.ImageFromPath(path, options...)
}

func FromStorage(path string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.ImageFromStorage(path, options...)
}

func FromURL(rawURL string, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.ImageFromURL(rawURL, options...)
}

func FromUpload(file contractsfilesystem.File, options ...contractsai.AttachmentOption) contractsai.Attachment {
	return frameworkai.ImageFromUpload(file, options...)
}

func FromID(id string) contractsai.ProviderFile {
	return frameworkai.ImageFromID(id)
}
