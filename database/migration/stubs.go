package migration

type Stubs struct {
}

func (receiver Stubs) Empty() string {
	return `package migrations

type DummyMigration struct{}

// Signature The unique signature for the migration.
func (r *DummyMigration) Signature() string {
	return "DummySignature"
}

// Up Run the migrations.
func (r *DummyMigration) Up() error {
	return nil
}

// Down Reverse the migrations.
func (r *DummyMigration) Down() error {
	return nil
}
`
}

func (receiver Stubs) Create() string {
	return `package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type DummyMigration struct{}

// Signature The unique signature for the migration.
func (r *DummyMigration) Signature() string {
	return "DummySignature"
}

// Up Run the migrations.
func (r *DummyMigration) Up() error {
	if !facades.Schema().HasTable("DummyTable") {
		return facades.Schema().Create("DummyTable", func(table schema.Blueprint) {
			table.ID()
			table.TimestampsTz()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *DummyMigration) Down() error {
 	return facades.Schema().DropIfExists("DummyTable")
}
`
}

func (receiver Stubs) Update() string {
	return `package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type DummyMigration struct{}

// Signature The unique signature for the migration.
func (r *DummyMigration) Signature() string {
	return "DummySignature"
}

// Up Run the migrations.
func (r *DummyMigration) Up() error {
	return facades.Schema().Table("DummyTable", func(table schema.Blueprint) {

	})
}

// Down Reverse the migrations.
func (r *DummyMigration) Down() error {
	return nil
}
`
}
