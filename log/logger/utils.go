package logger

import (
	"reflect"
	"unsafe"
)

// getContextValues gets all key-value pairs from the context
// TODO: This method has a high performance overhead; is it necessary to retain it?
func getContextValues(ctx any, values map[any]any) {
	contextValues := reflect.Indirect(reflect.ValueOf(ctx))
	contextKeys := reflect.TypeOf(ctx)
	if contextKeys.Kind() == reflect.Ptr {
		contextKeys = contextKeys.Elem()
	}

	if contextKeys.Kind() != reflect.Struct {
		return
	}

	value := struct {
		Key any
		Val any
	}{}

	for i := 0; i < contextValues.NumField(); i++ {
		reflectValue := contextValues.Field(i)
		if !reflectValue.CanAddr() {
			continue
		}

		reflectValue = reflect.NewAt(reflectValue.Type(), unsafe.Pointer(reflectValue.UnsafeAddr())).Elem()
		reflectField := contextKeys.Field(i)

		switch reflectField.Name {
		case "Context":
			getContextValues(reflectValue.Interface(), values)
		case "key":
			value.Key = reflectValue.Interface()
		case "val":
			value.Val = reflectValue.Interface()
		}
	}

	if value.Key != nil {
		values[value.Key] = value.Val
	}
}
