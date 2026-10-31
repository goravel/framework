package config

import "reflect"

// deepCopyMap returns a copy of source with every nested map and slice copied as well, so
// maps and slices reached through the result can not mutate the configuration.
func deepCopyMap(source map[string]any) map[string]any {
	copied := make(map[string]any, len(source))
	for key, value := range source {
		copied[key] = deepCopyValue(value)
	}

	return copied
}

// deepCopyValue copies maps and slices, and returns every other value as is.
func deepCopyValue(value any) any {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return nil
	}

	switch rv.Kind() {
	case reflect.Map:
		if rv.IsNil() {
			return value
		}

		copied := reflect.MakeMapWithSize(rv.Type(), rv.Len())
		iterator := rv.MapRange()
		for iterator.Next() {
			copied.SetMapIndex(iterator.Key(), copyValue(iterator.Value()))
		}

		return copied.Interface()
	case reflect.Slice:
		if rv.IsNil() {
			return value
		}

		copied := reflect.MakeSlice(rv.Type(), rv.Len(), rv.Len())
		for i := range rv.Len() {
			copied.Index(i).Set(copyValue(rv.Index(i)))
		}

		return copied.Interface()
	default:
		return value
	}
}

// copyValue copies v and re-boxes the result as a value assignable to v's type,
// using the zero value when the copy is a nil interface.
func copyValue(v reflect.Value) reflect.Value {
	copied := deepCopyValue(v.Interface())
	if copied == nil {
		return reflect.Zero(v.Type())
	}

	return reflect.ValueOf(copied)
}
