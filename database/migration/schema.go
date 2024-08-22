package migration

import (
	"github.com/goravel/framework/contracts/database/schema"
)

type Schema struct{}

func NewSchema() *Schema {
	return &Schema{}
}

func (r *Schema) Create(table string, callback func(table schema.Blueprint)) error {
	//TODO implement me
	panic("implement me")
}

func (r *Schema) Connection() schema.Schema {
	//TODO implement me
	panic("implement me")
}

func (r *Schema) Register(migrations []schema.Migration) {
	//TODO implement me
	panic("implement me")
}

func (r *Schema) Table(table string, callback func(table schema.Blueprint)) error {
	//TODO implement me
	panic("implement me")
}
