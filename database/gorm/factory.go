package gorm

import (
	"github.com/mitchellh/mapstructure"
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

// Times Specify the number of models you wish to create / make.
func (f *FactoryImpl) Times(count int) ormcontract.Factory {
	return f.newInstance(map[string]any{"count": count})
}

// Create a model and persist it in the database.
func (f *FactoryImpl) Create(value any) error {
	if err := f.Make(value); err != nil {
		return err
	}
	return f.query.Create(value)
}

// CreateQuietly create a model and persist it in the database without firing any events.
func (f *FactoryImpl) CreateQuietly(value any) error {
	if err := f.Make(value); err != nil {
		return err
	}
	return f.query.WithoutEvents().Create(value)
}

// Make a model instance that's not persisted in the database.
func (f *FactoryImpl) Make(value any) error {
	reflectValue := reflect.Indirect(reflect.ValueOf(value))
	switch reflectValue.Kind() {
	case reflect.Array, reflect.Slice:
		count := 1
		if f.count != nil {
			count = *f.count
		}
		for i := 0; i < count; i++ {
			elemValue := reflect.New(reflectValue.Type().Elem()).Interface()
			attributes := f.getRawAttributes(elemValue)
			if err := mapstructure.Decode(attributes, elemValue); err != nil {
				return err
			}
			reflectValue = reflect.Append(reflectValue, reflect.ValueOf(elemValue).Elem())
		}
		reflect.ValueOf(value).Elem().Set(reflectValue)
		return nil
	default:
		attributes := f.getRawAttributes(value)
		if err := mapstructure.Decode(attributes, value); err != nil {
			return err
		}
		return nil
	}
}

func (f *FactoryImpl) getRawAttributes(value any) any {
	modelFactoryMethod := reflect.ValueOf(value).MethodByName("Factory")
	if modelFactoryMethod.IsValid() {
		factoryResult := modelFactoryMethod.Call(nil)
		if len(factoryResult) > 0 {
			factoryInstance, ok := factoryResult[0].Interface().(factory.Factory)
			if ok {
				definitionMethod := reflect.ValueOf(factoryInstance).MethodByName("Definition")
				if definitionMethod.IsValid() {
					definitionResult := definitionMethod.Call(nil)
					if len(definitionResult) > 0 {
						return definitionResult[0].Interface()
					}
				}
			}
		}
	}
	return nil
}

// newInstance create a new factory instance.
func (f *FactoryImpl) newInstance(attributes ...map[string]any) ormcontract.Factory {
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

// TODO: Method below this will be removed in final release

// Model Set the model's attributes.
func (f *FactoryImpl) Model(value any) ormcontract.Factory {
	return f.newInstance(map[string]any{"model": value})
}

func (f *FactoryImpl) Faker() *gofakeit.Faker {
	return f.faker
}

// Raw Get a raw attribute array for the model's fields.
func (f *FactoryImpl) Raw() any {
	//if f.count == nil {
	//	return f.getRawAttributes()
	//}
	//result := make([]map[string]any, *f.count)
	//for i := 0; i < *f.count; i++ {
	//	item := f.getRawAttributes()
	//	if itemMap, ok := item.(map[string]any); ok {
	//		result[i] = itemMap
	//	}
	//}
	//return result
	return nil
}
