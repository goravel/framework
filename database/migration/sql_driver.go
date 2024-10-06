package migration

import (
	"github.com/goravel/framework/support/file"
)

// TODO Remove in v1.16
type SqlDriver struct {
	creator *SqlCreator
}

func NewSqlDriver(driver, charset string) *SqlDriver {
	return &SqlDriver{
		creator: NewSqlCreator(driver, charset),
	}
}

func (r *SqlDriver) Create(name string) error {
	table, create := TableGuesser{}.Guess(name)

	upStub, downStub := r.creator.GetStub(table, create)

	// Create the up.sql file.
	if err := file.Create(r.creator.GetPath(name, "up"), r.creator.PopulateStub(upStub, table)); err != nil {
		return err
	}

	// Create the down.sql file.
	if err := file.Create(r.creator.GetPath(name, "down"), r.creator.PopulateStub(downStub, table)); err != nil {
		return err
	}

	return nil
}

func (r *SqlDriver) Run(paths []string) error {
	return nil
}
