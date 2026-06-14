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
	// Title provides a display name for attachments created from
	// sources without a natural filename (e.g. fromString, fromByte).
	// Providers such as Anthropic require a non-empty document title.
	Title string
}

type AttachmentOption func(options *AttachmentOptions)

// StorableFile describes file content that can be uploaded to an AI provider.
type StorableFile interface {
	FileName() string
	MimeType() string
	Content(ctx context.Context) ([]byte, error)
}

// FileResponse describes a provider-managed file handle.
//
// Upload operations may return a lightweight handle that only guarantees ID().
// In that case MimeType() returns an empty string and Content(ctx) returns nil,
// nil until the file is resolved through the provider.
type FileResponse interface {
	ID() string
	MimeType() string
	Content(ctx context.Context) ([]byte, error)
}

// Attachment is request-scoped content sent with a user prompt.
type Attachment interface {
	StorableFile
	Kind() AttachmentKind
	Put(ctx context.Context, options ...Option) (FileResponse, error)
}

// ProviderFile describes a provider-managed file handle that can be attached to
// prompts and resolved or deleted later by ID.
type ProviderFile interface {
	Attachment
	ID() string
	Get(ctx context.Context, options ...Option) (FileResponse, error)
	Delete(ctx context.Context, options ...Option) error
}
