package orm

import (
	"context"
)

type EventType string

const EventRetrieved EventType = "retrieved"
const EventCreating EventType = "creating"
const EventCreated EventType = "created"
const EventUpdating EventType = "updating"
const EventUpdated EventType = "Updated"
const EventSaving EventType = "saving"
const EventSaved EventType = "saved"
const EventDeleting EventType = "deleting"
const EventDeleted EventType = "deleted"
const EventForceDeleting EventType = "force_deleting"
const EventForceDeleted EventType = "force_deleted"

type Event interface {
	// Context returns the event context.
	Context() context.Context
	// GetAttribute returns the attribute value for the given key.
	GetAttribute(key string) any
	// GetOriginal returns the original attribute value for the given key.
	GetOriginal(key string, def ...any) any
	// IsDirty returns true if the given column is dirty.
	IsDirty(columns ...string) bool
	// IsClean returns true if the given column is clean.
	IsClean(columns ...string) bool
	// Query returns the query instance.
	Query() Query
	// SetAttribute sets the attribute value for the given key.
	SetAttribute(key string, value any)
}

type DispatchesEvents interface {
	// DispatchesEvents returns the event handlers.
	DispatchesEvents() map[EventType]func(Event) error
}
