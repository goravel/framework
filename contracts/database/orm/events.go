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
	Context() context.Context
	GetAttribute(key string) any
	GetOriginal(key string, def ...any) any
	IsDirty(columns ...string) bool
	IsClean(columns ...string) bool
	Query() Query
	SetAttribute(key string, value any)
}

type DispatchesEvents interface {
	DispatchesEvents() map[EventType]func(Event) error
}
