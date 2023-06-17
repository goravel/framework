package console

type Stubs struct {
}

func (r Stubs) Model() string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/database/orm"
)

type DummyModel struct {
	orm.Model
}
`
}

func (r Stubs) Observer() string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/contracts/database/orm"
)


type DummyObserver struct{}

func (u *DummyObserver) Retrieved(event orm.Event) error {
	return nil
}

func (u *DummyObserver) Creating(event orm.Event) error {
	return nil
}

func (u *DummyObserver) Created(event orm.Event) error {
	return nil
}

func (u *DummyObserver) Updating(event orm.Event) error {
	return nil
}

func (u *DummyObserver) Updated(event orm.Event) error {
	return nil
}

func (u *DummyObserver) Saving(event orm.Event) error {
	return nil
}

func (u *DummyObserver) Saved(event orm.Event) error {
	return nil
}

func (u *DummyObserver) Deleting(event orm.Event) error {
	return nil
}

func (u *DummyObserver) Deleted(event orm.Event) error {
	return nil
}

func (u *DummyObserver) ForceDeleting(event orm.Event) error {
	return nil
}

func (u *DummyObserver) ForceDeleted(event orm.Event) error {
	return nil
}
`
}

func (r Stubs) Seeder() string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/database"
)
	
type DummySeeder struct {
	database.Seeder
}

func (s *DummySeeder) Signature() string {
	// Signature The name and signature of the seeder.
	
	return "DummySeeder"
}

func (s *DummySeeder) Run() error {
	// Run executes the seeder logic.
	
	return nil
}
`
}
