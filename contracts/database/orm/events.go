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
	IsDirty(columns ...string) bool
	IsClean(columns ...string) bool
	Query() Query
	Context() context.Context
	SetAttribute(key string, value any)
	GetAttribute(key string) any
	GetOriginal(key string, def ...any) any
}

type DispatchesEvents interface {
	DispatchesEvents() map[EventType]func(Event) error
}

type Retrieved interface {
	Retrieved(Event) error
}

type Creating interface {
	Creating(Event) error
}

type Created interface {
	Created(Event) error
}

type Updating interface {
	Updating(Event) error
}

type Updated interface {
	Updated(Event) error
}

type Saving interface {
	Saving(Event) error
}

type Saved interface {
	Saved(Event) error
}

type Deleting interface {
	Deleting(Event) error
}

type Deleted interface {
	Deleted(Event) error
}

type ForceDeleting interface {
	ForceDeleting(Event) error
}

type ForceDeleted interface {
	ForceDeleted(Event) error
}
