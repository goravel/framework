package gorm

import (
	"fmt"
	"reflect"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gookit/color"
	ormcontract "github.com/goravel/framework/contracts/database/orm"
)

type FactoryImpl struct {
	model  any             // model to generate
	count  int             // number of models to generate
	facker *gofakeit.Faker // faker instance
}

func NewFactoryImpl() *FactoryImpl {
	return &FactoryImpl{
		count:  1,
		facker: gofakeit.New(0),
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
	return nil
}

func (f *FactoryImpl) CreateManyQuietly() error {
	return nil
}

func (f *FactoryImpl) Create() error {
	color.Redf("Create %v", f.model)
	return nil
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
			factoryInstance, ok := factoryResult[0].Interface().(ormcontract.Factory)
			if ok {
				definitionMethod := reflect.ValueOf(factoryInstance).MethodByName("Definition")
				if definitionMethod.IsValid() {
					definitionResult := definitionMethod.Call(nil)
					if len(definitionResult) > 0 {
						definition := definitionResult[0].Interface()
						fmt.Printf("%#v\n", definition) // Print the definition in a human-readable format
						return definition
					}
				}
			}
		}
	}

	definition := f.Definition()
	return definition
}

func (f *FactoryImpl) Faker() *gofakeit.Faker { // Adjust this line to retrieve the Faker instance from your DI container
	// return gofakeit.New(0)
	return f.facker
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
	}

	if len(attributes) > 0 {
		attr := attributes[0]
		if count, ok := attr["count"].(int); ok {
			instance.count = count
		}
		if model, ok := attr["model"]; ok {
			instance.model = model
		}
	}

	return instance
}

func (f *FactoryImpl) Model(value any) ormcontract.Factory {
	return f.NewInstance(map[string]any{"model": value})
}

func (f *FactoryImpl) Definition() any {
	return nil
}
