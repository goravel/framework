package migration

import (
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/support/convert"
)

type ColumnDefinition struct {
	autoIncrement *bool
	change        *bool
	comment       *string
	def           any
	length        *int
	name          *string
	nullable      *bool
	ttype         *string
	unsigned      *bool
}

func (r *ColumnDefinition) AutoIncrement() migration.ColumnDefinition {
	r.autoIncrement = convert.Pointer(true)

	return r
}

func (r *ColumnDefinition) GetAutoIncrement() (autoIncrement bool) {
	if r.autoIncrement != nil {
		return *r.autoIncrement
	}

	return
}

func (r *ColumnDefinition) GetChange() (change bool) {
	if r.change != nil {
		return *r.change
	}

	return
}

func (r *ColumnDefinition) GetDefault() any {
	return r.def
}

func (r *ColumnDefinition) GetName() (name string) {
	if r.name != nil {
		return *r.name
	}

	return
}

func (r *ColumnDefinition) GetLength() (length int) {
	if r.length != nil {
		return *r.length
	}

	return
}

func (r *ColumnDefinition) GetNullable() bool {
	if r.nullable != nil {
		return *r.nullable
	}

	return false
}

func (r *ColumnDefinition) GetType() (ttype string) {
	if r.ttype != nil {
		return *r.ttype
	}

	return
}

func (r *ColumnDefinition) Unsigned() migration.ColumnDefinition {
	r.unsigned = convert.Pointer(true)

	return r
}
