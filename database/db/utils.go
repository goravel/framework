package db

import (
	"reflect"
	"strings"

	"github.com/goravel/framework/errors"
)

func convertToSliceMap(data any) ([]map[string]any, error) {
	if data == nil {
		return nil, nil
	}

	if maps, ok := data.([]map[string]any); ok {
		return maps, nil
	}

	val := reflect.ValueOf(data)
	typ := val.Type()

	// Handle pointer
	if typ.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, nil
		}
		val = val.Elem()
		typ = val.Type()
	}

	// Handle slice
	if typ.Kind() == reflect.Slice {
		length := val.Len()
		if length == 0 {
			return []map[string]any{}, nil
		}

		result := make([]map[string]any, length)
		for i := 0; i < length; i++ {
			elem := val.Index(i)
			m, err := convertToMap(elem.Interface())
			if err != nil {
				return nil, err
			}
			if m != nil {
				result[i] = m
			}
		}
		return result, nil
	}

	// Handle single value (struct or map)
	m, err := convertToMap(data)
	if err != nil {
		return nil, err
	}
	if m != nil {
		return []map[string]any{m}, nil
	}
	return nil, nil
}

func convertToMap(data any) (map[string]any, error) {
	if data == nil {
		return nil, nil
	}

	if m, ok := data.(map[string]any); ok {
		return m, nil
	}

	val := reflect.ValueOf(data)
	typ := val.Type()

	// Handle pointer
	if typ.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, nil
		}
		val = val.Elem()
		typ = val.Type()
	}

	if typ.Kind() != reflect.Struct {
		return nil, errors.DatabaseUnsupportedType.Args(typ.String(), "struct, []struct, map[string]any, []map[string]any").SetModule("DB")
	}

	// Handle struct
	result := make(map[string]any)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		// Handle embedded struct
		if field.Anonymous {
			fieldValue := val.Field(i)
			if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
				fieldValue = fieldValue.Elem()
			}
			if fieldValue.Kind() == reflect.Struct {
				embedded, err := convertToMap(fieldValue.Interface())
				if err != nil {
					return nil, err
				}
				for k, v := range embedded {
					result[k] = v
				}
			}
			continue
		}

		// Get field name from db tag or use field name
		tag := field.Tag.Get("db")
		if tag == "" || tag == "-" {
			continue
		}
		var fieldName string
		if comma := strings.Index(tag, ","); comma != -1 {
			fieldName = tag[:comma]
		} else {
			fieldName = tag
		}

		fieldValue := val.Field(i)
		if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
			fieldValue = fieldValue.Elem()
		}
		result[fieldName] = fieldValue.Interface()
	}
	return result, nil
}
