package orm

import (
	"github.com/brianvoe/gofakeit/v6"
)

//go:generate mockery --name=Factory
type Factory interface {
	New(attributes ...map[string]any) Factory
	Times(count int) Factory
	Count(count int) Factory
	Configure() Factory
	Raw() any
	CreateOne() error
	CreateOneQuietly() error
	CreateMany() error
	CreateManyQuietly() error
	Create() error
	CreateQuietly() error
	Store() error
	MakeOne() Factory
	Make() Factory
	MakeInstance() Factory
	GetExpandedAttributes() map[string]any
	GetRawAttributes() any
	Faker() *gofakeit.Faker
	ExpandAttributes(definition map[string]interface{}) map[string]interface{}
	Set() Factory
	Model(value any) Factory
	NewInstance(attributes ...map[string]any) Factory
}
