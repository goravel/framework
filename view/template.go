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

	values := r.view.GetShared()
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
			if field.IsExported() {
				values[field.Name] = value.Field(i).Interface()
			}
		}
	default:
		return values, errors.ViewInvalidData.Args(view, data[0])
	}

	return values, nil
}
