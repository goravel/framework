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
	errCastSliceElem   = "bind: cannot cast slice element value to field"
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

	tagToFieldMap := createTagToFieldMap(val)

	switch data := dataSrc.(type) {
	case url.Values:
		formData, ok := v.data.(*validate.FormData)
		if ok {
			return bindFromURLValues(data, formData, tagToFieldMap)
		}
		return bindFromURLValues(data, nil, tagToFieldMap)
	case map[string]any:
		return bindFromMap(data, tagToFieldMap)
	default:
		val := reflect.Indirect(reflect.ValueOf(dataSrc))
		if val.Kind() == reflect.Struct {
			return bindFromStruct(val, tagToFieldMap)
		} else {
			return fmt.Errorf("%s: %s", errUnsupportedType, reflect.TypeOf(dataSrc).String())
		}
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

func getFieldKey(structField reflect.StructField) string {
	if formTag := structField.Tag.Get("form"); formTag != "" {
		return formTag
	}
	if jsonTag := structField.Tag.Get("json"); jsonTag != "" {
		return jsonTag
	}
	return structField.Name
}

func setFieldValue(field reflect.Value, value any) error {
	castedValue, err := castValueToType(field, value)
	if err != nil {
		return fmt.Errorf("%s %s: %w", errCastValueField, field.Type().String(), err)
	}

	field.Set(castedValue)
	return nil
}

func createTagToFieldMap(val reflect.Value) map[string]reflect.Value {
	tagToFieldMap := make(map[string]reflect.Value)
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.CanSet() {
			continue
		}
		structField := typ.Field(i)
		tag := getFieldKey(structField)
		tagToFieldMap[tag] = field
		tagToFieldMap[strings.ToLower(tag)] = field
		tagToFieldMap[str.Camel2Case(tag)] = field
	}
	return tagToFieldMap
}

func bindFromURLValues(values url.Values, formData *validate.FormData, tagToFieldMap map[string]reflect.Value) error {
	for tag, field := range tagToFieldMap {
		if value, ok := values[tag]; ok && len(value) > 0 {
			if err := setFieldValue(field, value[0]); err != nil {
				return err
			}
		}
		if formData != nil {
			if value, ok := formData.Get(tag); ok {
				if err := setFieldValue(field, value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func bindFromMap(dataMap map[string]any, tagToFieldMap map[string]reflect.Value) error {
	for tag, field := range tagToFieldMap {
		if value, ok := dataMap[tag]; ok {
			if err := setFieldValue(field, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func bindFromStruct(dataSrc reflect.Value, tagToFieldMap map[string]reflect.Value) error {
	for tag, field := range tagToFieldMap {
		if value := dataSrc.FieldByName(tag); value.IsValid() && value.CanInterface() {
			if err := setFieldValue(field, value.Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

func castValueToType(field reflect.Value, value any) (reflect.Value, error) {
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
	default:
		castedValue = value
	}

	if err != nil {
		return reflect.Value{}, fmt.Errorf("%s: %w", errCastValueField, err)
	}

	return reflect.ValueOf(castedValue), nil
}
