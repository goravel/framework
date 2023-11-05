package validation

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/gookit/validate"
	"github.com/spf13/cast"

	httpvalidate "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/support/str"
)

const (
	errBindPointerOnly = "bind: must pass a pointer, not a value"
	errBindStructOnly  = "bind: must pass a pointer to a struct"
	errUnsupportedType = "bind: unsupported data source type"
	errCastValueField  = "bind: cannot cast value to field"
	errCastValueMap    = "bind: cannot cast value to map"
)

func init() {
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = true
	})
}

type Validator struct {
	instance *validate.Validation
	data     validate.DataFace
}

func NewValidator(instance *validate.Validation, data validate.DataFace) *Validator {
	instance.Validate()

	return &Validator{instance: instance, data: data}
}

func (v *Validator) Bind(ptr any) error {
	val := reflect.ValueOf(ptr)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New(errBindPointerOnly)
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New(errBindStructOnly)
	}

	dataSrc := v.data.Src()
	tagToFieldMap := v.createTagToFieldMap(val)

	switch data := dataSrc.(type) {
	case url.Values:
		formData, ok := v.data.(*validate.FormData)
		if ok {
			return v.bindFromURLValues(data, formData, tagToFieldMap)
		}

		return v.bindFromURLValues(data, nil, tagToFieldMap)
	case map[string]any:
		return v.bindFromMap(data, tagToFieldMap)
	default:
		val := reflect.Indirect(reflect.ValueOf(dataSrc))
		if val.Kind() == reflect.Struct {
			return v.bindFromStruct(val, tagToFieldMap)
		}

		return fmt.Errorf("%s: %s", errUnsupportedType, reflect.TypeOf(dataSrc).String())
	}
}

func (v *Validator) Errors() httpvalidate.Errors {
	if v.instance.Errors == nil || len(v.instance.Errors) == 0 {
		return nil
	}

	return NewErrors(v.instance.Errors)
}

func (v *Validator) Fails() bool {
	return v.instance.IsFail()
}

func (v *Validator) getFieldKey(structField reflect.StructField) string {
	if formTag := structField.Tag.Get("form"); formTag != "" {
		return formTag
	}
	if jsonTag := structField.Tag.Get("json"); jsonTag != "" {
		return jsonTag
	}

	return structField.Name
}

func (v *Validator) setFieldValue(field reflect.Value, value any) error {
	castedValue, err := v.castValueToType(field, value)
	if err != nil {
		return fmt.Errorf("%s %s: %w", errCastValueField, field.Type().String(), err)
	}

	field.Set(castedValue)
	return nil
}

func (v *Validator) createTagToFieldMap(val reflect.Value) map[string]reflect.Value {
	tagToFieldMap := make(map[string]reflect.Value)
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.CanSet() {
			continue
		}
		structField := typ.Field(i)
		tag := v.getFieldKey(structField)
		tagToFieldMap[tag] = field
		tagToFieldMap[strings.ToLower(tag)] = field
		tagToFieldMap[str.Camel2Case(tag)] = field
	}

	return tagToFieldMap
}

func (v *Validator) bindFromURLValues(values url.Values, formData *validate.FormData, tagToFieldMap map[string]reflect.Value) error {
	for tag, field := range tagToFieldMap {
		if value, ok := values[tag]; ok && len(value) > 0 {
			if err := v.setFieldValue(field, value[0]); err != nil {
				return err
			}
		}
		if formData != nil {
			if value, ok := formData.Get(tag); ok {
				if err := v.setFieldValue(field, value); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (v *Validator) bindFromMap(dataMap map[string]any, tagToFieldMap map[string]reflect.Value) error {
	for tag, field := range tagToFieldMap {
		if value, ok := dataMap[tag]; ok {
			if err := v.setFieldValue(field, value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Validator) bindFromStruct(dataSrc reflect.Value, tagToFieldMap map[string]reflect.Value) error {
	for tag, field := range tagToFieldMap {
		if value := dataSrc.FieldByName(tag); value.IsValid() && value.CanInterface() {
			if err := v.setFieldValue(field, value.Interface()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Validator) castValueToType(field reflect.Value, value any) (reflect.Value, error) {
	var castedValue any
	var err error

	switch field.Kind() {
	case reflect.String:
		castedValue = cast.ToString(value)
	case reflect.Int:
		castedValue, err = cast.ToIntE(value)
	case reflect.Int8:
		castedValue, err = cast.ToInt8E(value)
	case reflect.Int16:
		castedValue, err = cast.ToInt16E(value)
	case reflect.Int32:
		castedValue, err = cast.ToInt32E(value)
	case reflect.Int64:
		castedValue, err = cast.ToInt64E(value)
	case reflect.Uint:
		castedValue, err = cast.ToUintE(value)
	case reflect.Uint8:
		castedValue, err = cast.ToUint8E(value)
	case reflect.Uint16:
		castedValue, err = cast.ToUint16E(value)
	case reflect.Uint32:
		castedValue, err = cast.ToUint32E(value)
	case reflect.Uint64:
		castedValue, err = cast.ToUint64E(value)
	case reflect.Bool:
		castedValue, err = cast.ToBoolE(value)
	case reflect.Float32:
		castedValue, err = cast.ToFloat32E(value)
	case reflect.Float64:
		castedValue, err = cast.ToFloat64E(value)
	case reflect.Slice:
		castedValue, err = cast.ToSliceE(value)
	case reflect.Map:
		castedValue, err = cast.ToStringMapE(value)
	case reflect.Array:
		castedValue, err = cast.ToSliceE(value)
	case reflect.Struct:
		structType := field.Type()
		newStruct := reflect.New(structType).Elem()
		err := v.populateStruct(newStruct, value)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("%s: %w", errCastValueField, err)
		}
		castedValue = newStruct.Interface()
	default:
		castedValue = value
	}

	if err != nil {
		return reflect.Value{}, fmt.Errorf("%s: %w", errCastValueField, err)
	}

	return reflect.ValueOf(castedValue), nil
}

func (v *Validator) populateStruct(structVal reflect.Value, valueMap any) error {
	valueMapCasted, ok := valueMap.(map[string]any)
	if !ok {
		return fmt.Errorf("%s: %s", errCastValueMap, reflect.TypeOf(valueMap).String())
	}

	for i := 0; i < structVal.NumField(); i++ {
		structField := structVal.Field(i)
		structTypeField := structVal.Type().Field(i)
		if !structField.CanSet() {
			continue
		}

		if value, exists := valueMapCasted[v.getFieldKey(structTypeField)]; exists {
			// Handle the case where the struct field itself is a struct
			if structField.Kind() == reflect.Struct {
				// Recursively populate the nested struct
				err := v.populateStruct(structField, value)
				if err != nil {
					return err
				}
			} else {
				// Attempt to set the value directly
				valueToSet, err := v.castValueToType(structField, value)
				if err != nil {
					return fmt.Errorf("%s %s: %w", errCastValueField, structField.Type().String(), err)
				}
				structField.Set(valueToSet)
			}
		}
	}

	return nil
}
