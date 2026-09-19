package view

import (
	"reflect"

	contractsview "github.com/goravel/framework/contracts/view"
	"github.com/goravel/framework/errors"
)

type Template struct {
	view *View
	name string
	data map[string]any
	err  error
}

func NewTemplate(view *View, name string, data ...any) *Template {
	values, err := toMap(name, data...)

	return &Template{
		view: view,
		name: name,
		data: values,
		err:  err,
	}
}

func (r *Template) Data() map[string]any {
	return r.data
}

func (r *Template) Name() string {
	return r.name
}

func (r *Template) Render() (string, error) {
	if r.err != nil {
		return "", r.err
	}

	values := make(map[string]any, len(r.data)+8)
	r.view.shared.Range(func(key, value any) bool {
		values[key.(string)] = value
		return true
	})
	for key, value := range r.data {
		values[key] = value
	}

	return r.view.render(r.name, values)
}

func (r *Template) With(key string, value any) contractsview.Template {
	r.data[key] = value

	return r
}

// toMap copies the given map or struct into a new map, leaving the caller's data untouched.
func toMap(view string, data ...any) (map[string]any, error) {
	values := make(map[string]any)
	if len(data) == 0 || data[0] == nil {
		return values, nil
	}

	value := reflect.ValueOf(data[0])
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return values, nil
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return values, errors.ViewInvalidData.Args(view, data[0])
		}
		iter := value.MapRange()
		for iter.Next() {
			values[iter.Key().String()] = iter.Value().Interface()
		}
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Type().Field(i)
			if !field.IsExported() {
				continue
			}
			// Pointer fields are dereferenced the same way the route drivers do it, so a
			// nil pointer renders as an empty value instead of "<nil>".
			fieldValue := value.Field(i)
			if fieldValue.Kind() == reflect.Pointer {
				if fieldValue.IsNil() {
					values[field.Name] = nil
				} else {
					values[field.Name] = fieldValue.Elem().Interface()
				}
			} else {
				values[field.Name] = fieldValue.Interface()
			}
		}
	default:
		return values, errors.ViewInvalidData.Args(view, data[0])
	}

	return values, nil
}
