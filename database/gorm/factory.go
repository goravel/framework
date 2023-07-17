package gorm

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"

	"github.com/goravel/framework/contracts/database/factory"
	ormcontract "github.com/goravel/framework/contracts/database/orm"
)

type FactoryImpl struct {
	count *int              // number of models to generate
	query ormcontract.Query // query instance
}

func NewFactoryImpl(query ormcontract.Query) *FactoryImpl {
	return &FactoryImpl{
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
			attributes, err := f.getRawAttributes(elemValue)
			if err != nil {
				return err
			}
			if attributes == nil {
				return errors.New("failed to get raw attributes")
			}
			if err := mapstructure.Decode(attributes, elemValue); err != nil {
				return err
			}
			reflectValue = reflect.Append(reflectValue, reflect.ValueOf(elemValue).Elem())
		}
		reflect.ValueOf(value).Elem().Set(reflectValue)
		return nil
	default:
		attributes, err := f.getRawAttributes(value)
		if err != nil {
			return err
		}
		if attributes == nil {
			return errors.New("failed to get raw attributes")
		}
		if err := mapstructure.Decode(attributes, value); err != nil {
			return err
		}
		return nil
	}
}

func (f *FactoryImpl) getRawAttributes(value any) (any, error) {
	modelFactoryMethod := reflect.ValueOf(value).MethodByName("Factory")
	if !modelFactoryMethod.IsValid() {
		return nil, errors.New("factory method not found")
	}
	if !modelFactoryMethod.IsValid() {
		return nil, errors.New("factory method not found for value type " + reflect.TypeOf(value).String())
	}
	factoryResult := modelFactoryMethod.Call(nil)
	if len(factoryResult) == 0 {
		return nil, errors.New("factory method returned nothing")
	}
	factoryInstance, ok := factoryResult[0].Interface().(factory.Factory)
	if !ok {
		expectedType := reflect.TypeOf((*factory.Factory)(nil)).Elem()
		return nil, fmt.Errorf("factory method does not return a factory instance (expected %v)", expectedType)
	}
	definitionMethod := reflect.ValueOf(factoryInstance).MethodByName("Definition")
	if !definitionMethod.IsValid() {
		return nil, errors.New("definition method not found in factory instance")
	}
	definitionResult := definitionMethod.Call(nil)
	if len(definitionResult) == 0 {
		return nil, errors.New("definition method returned nothing")
	}

	return definitionResult[0].Interface(), nil
}

// newInstance create a new factory instance.
func (f *FactoryImpl) newInstance(attributes ...map[string]any) ormcontract.Factory {
	instance := &FactoryImpl{
		count: f.count,
		query: f.query,
	}

	if len(attributes) > 0 {
		attr := attributes[0]
		if count, ok := attr["count"].(int); ok {
			instance.count = &count
		}
	}

	return instance
}
