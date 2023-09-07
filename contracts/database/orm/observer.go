package orm

type Observer interface {
	// Retrieved Called when the model is retrieved from the database.
	Retrieved(Event) error
	// Creating Called when the model is being created.
	Creating(Event) error
	// Created Called when the model has been created.
	Created(Event) error
	// Updating Called when the model is being updated.
	Updating(Event) error
	// Updated Called when the model has been updated.
	Updated(Event) error
	// Saving Called when the model is being saved.
	Saving(Event) error
	// Saved Called when the model has been saved.
	Saved(Event) error
	// Deleting Called when the model is being deleted.
	Deleting(Event) error
	// Deleted Called when the model has been deleted.
	Deleted(Event) error
	// ForceDeleting Called when the model is being force deleted.
	ForceDeleting(Event) error
	// ForceDeleted Called when the model has been force deleted.
	ForceDeleted(Event) error
}
