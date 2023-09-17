package orm

type Observer interface {
	// Retrieved called when the model is retrieved from the database.
	Retrieved(Event) error
	// Creating called when the model is being created.
	Creating(Event) error
	// Created called when the model has been created.
	Created(Event) error
	// Updating called when the model is being updated.
	Updating(Event) error
	// Updated called when the model has been updated.
	Updated(Event) error
	// Saving called when the model is being saved.
	Saving(Event) error
	// Saved called when the model has been saved.
	Saved(Event) error
	// Deleting called when the model is being deleted.
	Deleting(Event) error
	// Deleted called when the model has been deleted.
	Deleted(Event) error
	// ForceDeleting called when the model is being force deleted.
	ForceDeleting(Event) error
	// ForceDeleted called when the model has been force deleted.
	ForceDeleted(Event) error
}
