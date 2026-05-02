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

// StorableFile describes file content that can be uploaded to an AI provider.
type StorableFile interface {
	FileName() string
	MimeType() string
	Content(ctx context.Context) ([]byte, error)
}

// StoredFileResponse describes a provider-managed file that can be referenced later.
type StoredFileResponse interface {
	ID() string
}

// Attachment is request-scoped content sent with a user prompt.
type Attachment interface {
	StorableFile
	Kind() AttachmentKind
}

// UploadableAttachment is an attachment that can be uploaded to a provider and reused by ID.
type UploadableAttachment interface {
	Attachment
	Put(ctx context.Context, options ...Option) (StoredFileResponse, error)
}
