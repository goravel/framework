package ai

import "context"

type AttachmentKind string

const (
	AttachmentKindImage AttachmentKind = "image"
	AttachmentKindFile  AttachmentKind = "file"
)

// Attachment is request-scoped content sent with a user prompt.
type Attachment interface {
	Kind() AttachmentKind
	Filename() string
	MimeType() string
	Content(ctx context.Context) ([]byte, error)
}
