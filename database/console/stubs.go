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
	
func (s *DummySeeder) Run() error {
	// Run executes the seeder logic.
	// To use the DummySeeder, register it in a service provider by calling facades.Seeder().Register and passing an instance of the DummySeeder as a pointer.
	// Example:
	//     facades.Seeder().Register([]seeder.Seeder{
	//         ...
	//         &seeders.DummySeeder{},
	//         ...
	//     })
	//
	// After registering, run the seeder by invoking the seed command in the console.
	// Example:
    //     go run . artisan db:seed --seeder DummySeeder
	//
	// The Run method performs the actual seeding operations.
	// To register other seeders to run, use the CallOnce method and provide the seeder instances.
	// Example:
	//     // Register multiple seeders
	//     s.CallOnce([]seeder.Seeder{&seeders.&OtherSeeder{},&seeders.&OtherSeeder2{}})
	//
	// Make sure to adjust the import statements and package paths based on your project structure.

	return nil
}
`
}
