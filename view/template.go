package view

import (
	"reflect"

	contractsview "github.com/goravel/framework/contracts/view"
	"github.com/goravel/framework/errors"
)

var _ contractsview.Template = (*Template)(nil)

// Template is a view bound to its data. Data that cannot be rendered, and a First that matched
// nothing, are reported by Render rather than when the template is built.
type Template struct {
	view *View
	name string
	data map[string]any
	err  error
}

// NewTemplate binds a view to its data. The data may be a map with string keys or a struct,
// and is copied, so later changes to the caller's value are not picked up.
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
	data := make(map[string]any, len(r.data))
	for key, value := range r.data {
		data[key] = value
	}

	return data
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
	if len(data) == 0 || data[0] == nil {
		return make(map[string]any), nil
	}

	// The map the templates already want is copied directly: reflect boxes every value it
	// reads back, which for this type costs an allocation per entry and buys nothing.
	switch typed := data[0].(type) {
	case map[string]any:
		values := make(map[string]any, len(typed))
		for key, value := range typed {
			values[key] = value
		}

		return values, nil
	case *map[string]any:
		if typed == nil {
			return make(map[string]any), nil
		}

		values := make(map[string]any, len(*typed))
		for key, value := range *typed {
			values[key] = value
		}

		return values, nil
	}

	value := reflect.ValueOf(data[0])
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return make(map[string]any), nil
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return make(map[string]any), errors.ViewInvalidData.Args(view, data[0])
		}
		values := make(map[string]any, value.Len())
		iter := value.MapRange()
		for iter.Next() {
			values[iter.Key().String()] = iter.Value().Interface()
		}

		return values, nil
	case reflect.Struct:
		values := make(map[string]any, value.NumField())
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

		return values, nil
	default:
		return make(map[string]any), errors.ViewInvalidData.Args(view, data[0])
	}
}
