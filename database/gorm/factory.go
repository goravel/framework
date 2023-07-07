package gorm

import (
	"reflect"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/database/factory"
	ormcontract "github.com/goravel/framework/contracts/database/orm"
)

type FactoryImpl struct {
	model any               // model to generate
	count int               // number of models to generate
	faker *gofakeit.Faker   // faker instance
	query ormcontract.Query // query instance
}

func NewFactoryImpl(query ormcontract.Query) *FactoryImpl {
	return &FactoryImpl{
		faker: gofakeit.New(0),
		query: query,
	}
}

func (f *FactoryImpl) New(attributes ...map[string]any) ormcontract.Factory {
	return f.NewInstance(attributes...).Configure()
}

func (f *FactoryImpl) Times(count int) ormcontract.Factory {
	return f.New().Count(count)
}

func (f *FactoryImpl) Configure() ormcontract.Factory {
	return f
}

func (f *FactoryImpl) Raw() any {
	return nil
}

func (f *FactoryImpl) CreateOne() error {
	return nil
}

func (f *FactoryImpl) CreateOneQuietly() error {
	return nil
}

func (f *FactoryImpl) CreateMany() error {
	for i := 0; i < f.count; i++ {
		color.Cyanln(f.GetRawAttributes())
		//err := f.query.Create(f.GenerateOne())
		//if err != nil {
		//	return err
		//}
	}
	return nil
}

func (f *FactoryImpl) CreateManyQuietly() error {
	return nil
}

func (f *FactoryImpl) Create() error {
	return f.query.Model(f.model).Create(f.GetRawAttributes())
}

func (f *FactoryImpl) CreateQuietly() error {
	return nil
}

func (f *FactoryImpl) Store() error {
	return nil
}

func (f *FactoryImpl) MakeOne() ormcontract.Factory {
	return nil
}

func (f *FactoryImpl) Make() ormcontract.Factory {
	return nil
}

func (f *FactoryImpl) MakeInstance() ormcontract.Factory {
	return nil
}

func (f *FactoryImpl) GetExpandedAttributes() map[string]any {
	return nil
}

func (f *FactoryImpl) GetRawAttributes() any {
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
						definition := definitionResult[0].Interface() // Print the definition in a human-readable format
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

func (f *FactoryImpl) ExpandAttributes(definition map[string]interface{}) map[string]interface{} {
	expandedAttributes := make(map[string]interface{})

	for key, attribute := range definition {
		switch attr := attribute.(type) {
		case func(map[string]interface{}) interface{}:
			// Evaluate the callable attribute
			expandedAttribute := attr(definition)
			expandedAttributes[key] = expandedAttribute
		default:
			expandedAttributes[key] = attr
		}
	}

	return expandedAttributes
}

func (f *FactoryImpl) Set() ormcontract.Factory {
	return f
}

func (f *FactoryImpl) Count(count int) ormcontract.Factory {
	return f.NewInstance(map[string]any{"count": count})
}

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
			instance.count = count
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

func (f *FactoryImpl) Model(value any) ormcontract.Factory {
	return f.NewInstance(map[string]any{"model": value})
}
