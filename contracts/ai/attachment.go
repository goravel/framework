package ai

import "context"

type AttachmentKind string

const (
	AttachmentKindImage AttachmentKind = "image"
	AttachmentKindFile  AttachmentKind = "file"
)

type AttachmentOptions struct {
	MimeType string
	// Only used by FromStorage.
	Disk string
}

type AttachmentOption func(options *AttachmentOptions)

type StorableFile interface {
	FileName() string
	MimeType() string
	Content(ctx context.Context) ([]byte, error)
}

type StoredFileResponse interface {
	ID() string
}

// Attachment is request-scoped content sent with a user prompt.
type Attachment interface {
	StorableFile
	Kind() AttachmentKind
	Put(options ...Option) (StoredFileResponse, error)
}
