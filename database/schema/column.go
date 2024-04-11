package schema

import (
	schemacontract "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/support/convert"
)

type ColumnDefinition struct {
	allowed       []string
	autoIncrement *bool
	change        *bool
	comment       *string
	def           *string
	length        *int
	name          *string
	nullable      *bool
	places        *int
	precision     *int
	total         *int
	ttype         *string
	unsigned      *bool
}

func (r *ColumnDefinition) Change() {
	r.change = convert.Pointer(true)
}

func (r *ColumnDefinition) Comment(comment string) schemacontract.ColumnDefinition {
	r.comment = &comment

	return r
}

func (r *ColumnDefinition) GetAllowed() []string {
	return r.allowed
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

func (r *ColumnDefinition) GetComment() (comment string) {
	if r.comment != nil {
		return *r.comment
	}

	return
}

func (r *ColumnDefinition) GetLength() (length int) {
	if r.length != nil {
		return *r.length
	}

	return
}

func (r *ColumnDefinition) GetName() (name string) {
	if r.name != nil {
		return *r.name
	}

	return
}

func (r *ColumnDefinition) GetNullable() bool {
	if r.nullable != nil {
		return *r.nullable
	}

	return false
}

func (r *ColumnDefinition) GetPlaces() (places int) {
	if r.places != nil {
		return *r.places
	}

	return 2
}

func (r *ColumnDefinition) GetPrecision() (precision int) {
	if r.precision != nil {
		return *r.precision
	}

	return
}

func (r *ColumnDefinition) GetTotal() (total int) {
	if r.total != nil {
		return *r.total
	}

	return 8
}

func (r *ColumnDefinition) GetType() (ttype string) {
	if r.ttype != nil {
		return *r.ttype
	}

	return
}

func (r *ColumnDefinition) GetUnsigned() (unsigned bool) {
	if r.unsigned != nil {
		return *r.unsigned
	}

	return
}

func (r *ColumnDefinition) Nullable() schemacontract.ColumnDefinition {
	r.nullable = convert.Pointer(true)

	return r
}
