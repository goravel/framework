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

// Attachment is request-scoped content sent with a user prompt.
type Attachment interface {
	Kind() AttachmentKind
	FileName() string
	MimeType() string
	Content(ctx context.Context) ([]byte, error)
}
