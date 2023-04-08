package orm

type Retrieved interface {
	Retrieved(Query) error
}

type Creating interface {
	Creating(Query) error
}

type Created interface {
	Created(Query) error
}

type Updating interface {
	Updating(Query) error
}

type Updated interface {
	Updated(Query) error
}

type Saving interface {
	Saving(Query) error
}

type Saved interface {
	Saved(Query) error
}

type Deleting interface {
	Deleting(Query) error
}

type Deleted interface {
	Deleted(Query) error
}

type ForceDeleting interface {
	ForceDeleting(Query) error
}

type ForceDeleted interface {
	ForceDeleted(Query) error
}
