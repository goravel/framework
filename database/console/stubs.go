package console

type Stubs struct {
}

func (r Stubs) Model() string {
	return `package models

import (
	"github.com/goravel/framework/database/orm"
)

type DummyModel struct {
	orm.Model
}
`
}

func (r Stubs) Observer() string {
	return `package observers

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

func (r Stubs) Factory(name string) string {
	return fmt.Sprintf(`package factories

import (
	"github.com/goravel/framework/database/orm"
)

type %s struct {
	orm.Model
}
`, str.Camel2Case(name))
}
