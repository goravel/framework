package database

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

// Count Specify the number of models you wish to create / make.
func (f *FactoryImpl) Count(count int) ormcontract.Factory {
	return f.newInstance(map[string]any{"count": count})
}

// Create a model and persist it in the database.
func (f *FactoryImpl) Create(value any, attributes ...map[string]any) error {
	if err := f.Make(value, attributes...); err != nil {
		return err
	}

	return f.query.Create(value)
}

// CreateQuietly create a model and persist it in the database without firing any events.
func (f *FactoryImpl) CreateQuietly(value any, attributes ...map[string]any) error {
	if err := f.Make(value, attributes...); err != nil {
		return err
	}

	return f.query.WithoutEvents().Create(value)
}

// Make a model instance that's not persisted in the database.
func (f *FactoryImpl) Make(value any, attributes ...map[string]any) error {
	reflectValue := reflect.Indirect(reflect.ValueOf(value))
	switch reflectValue.Kind() {
	case reflect.Array, reflect.Slice:
		count := 1
		if f.count != nil {
			count = *f.count
		}
		for i := 0; i < count; i++ {
			elemValue := reflect.New(reflectValue.Type().Elem()).Interface()
			attributes, err := f.getRawAttributes(elemValue, attributes...)
			if err != nil {
				return err
			}
			if attributes == nil {
				return errors.New("failed to get raw attributes")
			}
			decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				Squash: true,
				Result: elemValue,
			})
			if err != nil {
				return err
			}
			if err := decoder.Decode(attributes); err != nil {
				return err
			}
			reflectValue = reflect.Append(reflectValue, reflect.ValueOf(elemValue).Elem())
		}

		reflect.ValueOf(value).Elem().Set(reflectValue)

		return nil
	default:
		attributes, err := f.getRawAttributes(value, attributes...)
		if err != nil {
			return err
		}
		if attributes == nil {
			return errors.New("failed to get raw attributes")
		}
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Squash: true,
			Result: value,
		})
		if err != nil {
			return err
		}

		return decoder.Decode(attributes)
	}
}

func (f *FactoryImpl) getRawAttributes(value any, attributes ...map[string]any) (map[string]any, error) {
	factoryModel, exist := value.(factory.Model)
	if !exist {
		return nil, fmt.Errorf("%s does not find factory method", reflect.TypeOf(value).String())
	}

	definition := factoryModel.Factory().Definition()
	if len(attributes) > 0 {
		for key, value := range attributes[0] {
			definition[key] = value
		}
	}

	return definition, nil
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
