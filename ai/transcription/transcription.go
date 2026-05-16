package transcription

import (
	frameworkai "github.com/goravel/framework/ai"
	contractsai "github.com/goravel/framework/contracts/ai"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
)

func WithMimeType(mimeType string) contractsai.AttachmentOption {
	return frameworkai.WithMimeType(mimeType)
}

func WithDisk(disk string) contractsai.AttachmentOption {
	return frameworkai.WithDisk(disk)
}

func Of(file contractsai.StorableFile) contractsai.TranscriptionRequest {
	return facades.AI().Transcription(file)
}

func FromPath(path string, options ...contractsai.AttachmentOption) contractsai.TranscriptionRequest {
	return Of(frameworkai.DocumentFromPath(path, options...))
}

func FromStorage(path string, options ...contractsai.AttachmentOption) contractsai.TranscriptionRequest {
	return Of(frameworkai.DocumentFromStorage(path, options...))
}

func FromUpload(file contractsfilesystem.File, options ...contractsai.AttachmentOption) contractsai.TranscriptionRequest {
	return Of(frameworkai.DocumentFromUpload(file, options...))
}
