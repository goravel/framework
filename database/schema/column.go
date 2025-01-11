package schema

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/support/convert"
)

type ColumnDefinition struct {
	allowed            []any
	autoIncrement      *bool
	change             bool
	comment            *string
	def                any
	length             *int
	name               *string
	nullable           *bool
	onUpdate           any
	places             *int
	precision          *int
	total              *int
	ttype              *string
	unsigned           *bool
	useCurrent         *bool
	useCurrentOnUpdate *bool
}

func (r *ColumnDefinition) AutoIncrement() schema.ColumnDefinition {
	r.autoIncrement = convert.Pointer(true)

	return r
}

func (r *ColumnDefinition) Change() schema.ColumnDefinition {
	r.change = true

	return r
}

func (r *ColumnDefinition) Comment(comment string) schema.ColumnDefinition {
	r.comment = &comment

	return r
}

func (r *ColumnDefinition) Default(def any) schema.ColumnDefinition {
	r.def = def

	return r
}

func (r *ColumnDefinition) GetAllowed() []any {
	return r.allowed
}

func (r *ColumnDefinition) GetAutoIncrement() bool {
	if r.autoIncrement != nil {
		return *r.autoIncrement
	}

	return false
}

func (r *ColumnDefinition) GetComment() string {
	if r.comment != nil {
		return *r.comment
	}

	return ""
}

func (r *ColumnDefinition) GetDefault() any {
	return r.def
}

func (r *ColumnDefinition) GetName() string {
	if r.name != nil {
		return *r.name
	}

	return ""
}

func (r *ColumnDefinition) GetLength() int {
	if r.length != nil {
		return *r.length
	}

	return 0
}

func (r *ColumnDefinition) GetNullable() bool {
	if r.nullable != nil {
		return *r.nullable
	}

	return false
}

func (r *ColumnDefinition) GetOnUpdate() any {
	return r.onUpdate
}

func (r *ColumnDefinition) GetPlaces() int {
	if r.places != nil {
		return *r.places
	}

	return 2
}

func (r *ColumnDefinition) GetPrecision() int {
	if r.precision != nil {
		return *r.precision
	}

	return 0
}

func (r *ColumnDefinition) GetTotal() int {
	if r.total != nil {
		return *r.total
	}

	return 8
}

func (r *ColumnDefinition) GetType() string {
	if r.ttype != nil {
		return *r.ttype
	}

	return ""
}

func (r *ColumnDefinition) GetUnsigned() bool {
	if r.unsigned != nil {
		return *r.unsigned
	}

	return false
}

func (r *ColumnDefinition) GetUseCurrent() bool {
	if r.useCurrent != nil {
		return *r.useCurrent
	}

	return false
}

func (r *ColumnDefinition) GetUseCurrentOnUpdate() bool {
	if r.useCurrentOnUpdate != nil {
		return *r.useCurrentOnUpdate
	}

	return false
}

func (r *ColumnDefinition) IsChange() bool {
	return r.change
}

func (r *ColumnDefinition) IsSetComment() bool {
	return r != nil && r.comment != nil
}

func (r *ColumnDefinition) Nullable() schema.ColumnDefinition {
	r.nullable = convert.Pointer(true)

	return r
}

func (r *ColumnDefinition) OnUpdate(value any) schema.ColumnDefinition {
	r.onUpdate = value

	return r
}

func (r *ColumnDefinition) Places(places int) schema.ColumnDefinition {
	r.places = convert.Pointer(places)

	return r
}

func (r *ColumnDefinition) Total(total int) schema.ColumnDefinition {
	r.total = convert.Pointer(total)

	return r
}

func (r *ColumnDefinition) Unsigned() schema.ColumnDefinition {
	r.unsigned = convert.Pointer(true)

	return r
}

func (r *ColumnDefinition) UseCurrent() schema.ColumnDefinition {
	r.useCurrent = convert.Pointer(true)

	return r
}

func (r *ColumnDefinition) UseCurrentOnUpdate() schema.ColumnDefinition {
	r.useCurrentOnUpdate = convert.Pointer(true)

	return r
}
