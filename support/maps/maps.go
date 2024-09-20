package maps

import (
	"reflect"
)

// Add an element to a map if it doesn't exist.
func Add[K comparable, V any](mp map[K]V, k K, v V) {
	if Exists(mp, k) {
		return
	}

	Set(mp, k, v)
}

// Exists checks if the given key exists in the provided map.
func Exists[K comparable, V any](mp map[K]V, key K) bool {
	_, ok := mp[key]
	return ok
}

// Forget removes a given key or keys from the provided map.
func Forget[K comparable, V any](mp map[K]V, keys ...K) {
	for _, key := range keys {
		if _, ok := mp[key]; ok {
			delete(mp, key)
			continue
		}
	}
}

func FromStruct(data any) map[string]any {
	res := make(map[string]any)
	dataType := reflect.TypeOf(data)
	dataValue := reflect.ValueOf(data)

	if dataType.Kind() == reflect.Pointer {
		dataType = dataType.Elem()
		dataValue = dataValue.Elem()
	}

	if dataType.Kind() != reflect.Struct {
		return res
	}

	for i := 0; i < dataType.NumField(); i++ {
		fieldType := dataType.Field(i)
		fieldValue := dataValue.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		if fieldValue.Kind() == reflect.Pointer {
			if fieldValue.IsNil() {
				res[fieldType.Name] = nil
				continue
			}

			fieldValue = fieldValue.Elem()
		}

		if fieldValue.Kind() == reflect.Struct {
			subStructMap := FromStruct(fieldValue.Interface())
			if fieldType.Anonymous {
				for key, value := range subStructMap {
					res[key] = value
				}
			} else {
				res[fieldType.Name] = subStructMap
			}
		} else {
			res[fieldType.Name] = fieldValue.Interface()
		}
	}

	return res
}

// Get an element from a map
func Get[K comparable, V any](mp map[K]V, key K, defaults ...V) V {
	val, ok := mp[key]

	if !ok && len(defaults) > 0 {
		return defaults[0]
	}

	return val
}

func GetDeep(mp any, keys []any, defaults ...any) any {
	if len(keys) == 0 {
		return *new(any)
	}

	defaultValue := *new(any)

	if len(defaults) > 0 {
		defaultValue = defaults[0]
	}

	mpValue := reflect.ValueOf(mp)
	if mpValue.Kind() != reflect.Map {
		return defaultValue
	}

	keyValue := reflect.ValueOf(keys[0])
	mpValue = mpValue.MapIndex(keyValue)

	if !mpValue.IsValid() {
		return defaultValue
	}

	if len(keys) == 1 {
		return mpValue.Interface()
	}

	return GetDeep(mpValue.Interface(), keys[1:], defaults...)
}

// Has checks if the given key or keys exist in the provided map.
func Has[K comparable, V any](mp map[K]V, keys ...K) bool {
	if len(keys) == 0 || len(mp) == 0 {
		return false
	}

	for _, key := range keys {
		if !Exists(mp, key) {
			return false
		}
	}

	return true
}

// HasAny checks if the given key or keys exist in the provided map.
func HasAny[K comparable, V any](mp map[K]V, keys ...K) bool {
	for _, key := range keys {
		if Has(mp, key) {
			return true
		}
	}

	return false
}

// Only returns the items in the map with the specified keys.
func Only[K comparable, V any](mp map[K]V, keys ...K) map[K]V {
	result := make(map[K]V)
	for _, key := range keys {
		if Exists(mp, key) {
			result[key] = mp[key]
		}
	}

	return result
}

// Pull returns a new map with the specified keys removed.
func Pull[K comparable, V any](mp map[K]V, key K, def ...V) V {
	if val, ok := mp[key]; ok {
		delete(mp, key)
		return val
	}

	if len(def) > 0 {
		return def[0]
	}

	return *new(V)
}

// Set an element to a map.
func Set[K comparable, V any](mp map[K]V, k K, v V) {
	mp[k] = v
}

// Where filters the items in a map using the given callback.
func Where[K comparable, V any](mp map[K]V, callback func(K, V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range mp {
		if callback(k, v) {
			result[k] = v
		}
	}

	return result
}
