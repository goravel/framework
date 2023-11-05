package validation

import (
	"errors"
	"net/url"
	"reflect"
	"strings"

	"github.com/gookit/validate"
	"github.com/goravel/framework/support/str"

	httpvalidate "github.com/goravel/framework/contracts/validation"
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
		return errors.New("bind: must pass a pointer, not a value")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("bind: must pass a pointer to a struct")
	}

	dataSrc := v.data.Src()

	// getFieldKey returns the key to use for the given field
	getFieldKey := func(structField reflect.StructField) string {
		formTag := structField.Tag.Get("form")
		if len(formTag) == 0 {
			formTag = structField.Tag.Get("json")
		}
		if len(formTag) == 0 {
			formTag = structField.Name
		}
		return formTag
	}

	// setFieldValue sets the value of the given field
	setFieldValue := func(field reflect.Value, structField reflect.StructField, value any) error {
		fieldValue := reflect.ValueOf(value)
		if fieldValue.Kind() == field.Kind() {
			field.Set(fieldValue.Convert(field.Type()))
		} else {
			return errors.New("bind: cannot assign value to field " + structField.Name)
		}
		return nil
	}

	// Check for url.Values
	if values, ok := dataSrc.(url.Values); ok {
		data := make(map[string]any)
		for key, value := range values {
			if len(value) > 0 {
				data[key] = value[0]
			}
		}

		formData, ok := v.data.(*validate.FormData)
		if ok {
			for key, value := range formData.Files {
				data[key] = value
			}
		}

		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if !field.CanSet() {
				continue
			}

			structField := val.Type().Field(i)
			formTag := getFieldKey(structField)

			if value, ok := data[formTag]; ok {
				if err := setFieldValue(field, structField, value); err != nil {
					return err
				}
				continue
			}
			// Try lower case
			if value, ok := data[strings.ToLower(formTag)]; ok {
				if err := setFieldValue(field, structField, value); err != nil {
					return err
				}
				continue
			}
			// Try snake case
			if value, ok := data[str.Camel2Case(formTag)]; ok {
				if err := setFieldValue(field, structField, value); err != nil {
					return err
				}
				continue
			}
		}
		return nil
	}

	// Check for map[string]any
	if dataSrcMap, ok := dataSrc.(map[string]any); ok {
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if !field.CanSet() {
				continue
			}

			structField := val.Type().Field(i)
			formTag := getFieldKey(structField)

			if value, ok := dataSrcMap[formTag]; ok {
				if err := setFieldValue(field, structField, value); err != nil {
					return err
				}
				continue
			}
			// Try lower case
			if value, ok := dataSrcMap[strings.ToLower(formTag)]; ok {
				if err := setFieldValue(field, structField, value); err != nil {
					return err
				}
				continue
			}
			// Try snake case
			if value, ok := dataSrcMap[str.Camel2Case(formTag)]; ok {
				if err := setFieldValue(field, structField, value); err != nil {
					return err
				}
				continue
			}
		}
		return nil
	}

	// Check for custom struct
	dataSrcVal := reflect.ValueOf(dataSrc)
	// struct may be a pointer
	if dataSrcVal.Kind() == reflect.Ptr {
		dataSrcVal = dataSrcVal.Elem()
	}
	if dataSrcVal.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if !field.CanSet() {
				continue
			}

			structField := val.Type().Field(i)
			formTag := getFieldKey(structField)

			if value := dataSrcVal.FieldByName(formTag); value.IsValid() && value.CanInterface() {
				if err := setFieldValue(field, structField, value.Interface()); err != nil {
					return err
				}
				continue
			}
			// Try lower case
			if value := dataSrcVal.FieldByName(strings.ToLower(formTag)); value.IsValid() && value.CanInterface() {
				if err := setFieldValue(field, structField, value.Interface()); err != nil {
					return err
				}
				continue
			}
			// Try snake case
			if value := dataSrcVal.FieldByName(str.Camel2Case(formTag)); value.IsValid() && value.CanInterface() {
				if err := setFieldValue(field, structField, value.Interface()); err != nil {
					return err
				}
				continue
			}
		}
		return nil
	}

	return errors.New("bind: unsupported data source type " + reflect.TypeOf(dataSrc).String())
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
