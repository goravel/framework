package gorm

import (
	"reflect"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/goravel/framework/contracts/database/factory"
	ormcontract "github.com/goravel/framework/contracts/database/orm"
)

type FactoryImpl struct {
	model any               // model to generate
	count *int              // number of models to generate
	faker *gofakeit.Faker   // faker instance
	query ormcontract.Query // query instance
}

func NewFactoryImpl(query ormcontract.Query) *FactoryImpl {
	return &FactoryImpl{
		faker: gofakeit.New(0),
		query: query,
	}
}

// Count Specify the number of models you wish to create / make.
func (f *FactoryImpl) Count(count int) ormcontract.Factory {
	return f.NewInstance(map[string]any{"count": count})
}

// Raw Get a raw attribute array for the model's fields.
func (f *FactoryImpl) Raw() any {
	if f.count == nil {
		return f.getRawAttributes()
	}
	result := make([]map[string]any, *f.count)
	for i := 0; i < *f.count; i++ {
		item := f.getRawAttributes()
		if itemMap, ok := item.(map[string]any); ok {
			result[i] = itemMap
		}
	}
	return result
}

// Create a model and persist it in the database.
func (f *FactoryImpl) Create() error {
	return f.query.Model(f.model).Create(f.Raw())
}

// CreateQuietly create a model and persist it in the database without firing any events.
func (f *FactoryImpl) CreateQuietly() error {
	return f.query.Model(f.model).WithoutEvents().Create(f.Raw())
}

// Make a model instance that's not persisted in the database.
func (f *FactoryImpl) Make() ormcontract.Factory {
	return nil
}

func (f *FactoryImpl) getRawAttributes() any {
	modelFactoryMethod := reflect.ValueOf(f.model).MethodByName("Factory")

	if modelFactoryMethod.IsValid() {
		factoryResult := modelFactoryMethod.Call(nil)
		if len(factoryResult) > 0 {
			factoryInstance, ok := factoryResult[0].Interface().(factory.Factory)
			if ok {
				definitionMethod := reflect.ValueOf(factoryInstance).MethodByName("Definition")
				if definitionMethod.IsValid() {
					definitionResult := definitionMethod.Call(nil)
					if len(definitionResult) > 0 {
						definition := definitionResult[0].Interface()
						return definition
					}
				}
			}
		}
	}
	return nil
}

func (f *FactoryImpl) Faker() *gofakeit.Faker {
	return f.faker
}

// NewInstance create a new factory instance.
func (f *FactoryImpl) NewInstance(attributes ...map[string]any) ormcontract.Factory {
	instance := &FactoryImpl{
		count: f.count,
		model: f.model,
		query: f.query,
		faker: f.faker,
	}

	if len(attributes) > 0 {
		attr := attributes[0]
		if count, ok := attr["count"].(int); ok {
			instance.count = &count
		}
		if model, ok := attr["model"]; ok {
			instance.model = model
		}
		if faker, ok := attr["faker"]; ok {
			instance.faker = faker.(*gofakeit.Faker)
		}
		if query, ok := attr["query"]; ok {
			instance.query = query.(ormcontract.Query)
		}
	}

	return instance
}

// Model Set the model's attributes.
func (f *FactoryImpl) Model(value any) ormcontract.Factory {
	return f.NewInstance(map[string]any{"model": value})
}
