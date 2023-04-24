package orm

type Observer interface {
	Retrieved(Event) error
	Creating(Event) error
	Created(Event) error
	Updating(Event) error
	Updated(Event) error
	Saving(Event) error
	Saved(Event) error
	Deleting(Event) error
	Deleted(Event) error
	ForceDeleting(Event) error
	ForceDeleted(Event) error
}
