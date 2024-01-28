package object

import (
	"reflect"

	"github.com/gookit/goutil/maputil"
)

// Accessible determines whether the given value is array-accessible.
func Accessible[V any](obj V) bool {
	kind := reflect.ValueOf(obj).Kind()

	return kind == reflect.Array || kind == reflect.Slice || kind == reflect.Map
}

// Add an element to a map using “dot” notation if it doesn't exist.
func Add(obj map[string]any, k string, v any) (map[string]any, error) {
	if val := Get(obj, k); val != nil {
		return obj, nil
	}

	return Set(obj, k, v)
}

// Dot flattens a map using dot notation.
func Dot(obj map[string]any) map[string]any {
	return maputil.Flatten(obj)
}

// Exists checks if the given key exists in the provided map.
func Exists[K comparable, V any](obj map[K]V, key K) bool {
	return Has(obj, key)
}

// Forget removes a given key or keys from the provided map.
func Forget(obj map[string]any, keys ...string) {
	// need to finish
}

// Get an item from an object using "dot" notation.
func Get(obj map[string]any, key string, defaults ...any) any {
	val := maputil.DeepGet(obj, key)

	if val == nil && len(defaults) > 0 {
		return defaults[0]
	}

	return val
}

// Has checks if the given key or keys exist in the provided map.
func Has(obj any, keys ...any) bool {
	ok, _ := maputil.HasAllKeys(obj, keys...)
	return ok
}

// HasAny checks if the given key or keys exist in the provided map.
func HasAny(obj any, keys ...any) bool {
	ok, _ := maputil.HasOneKey(obj, keys...)
	return ok
}

// Only returns the items in the map with the specified keys.
func Only[K comparable, V any](obj map[K]V, keys ...K) map[K]V {
	result := make(map[K]V)
	for _, key := range keys {
		if Exists(obj, key) {
			result[key] = obj[key]
		}
	}

	return result
}

// Pull returns a new map with the specified keys removed.
func Pull(obj map[string]any, key string, def ...any) any {
	value := Get(obj, key, def)
	Forget(obj, key)
	return value
}

// Set an element to a map using “dot” notation.
func Set(obj map[string]any, k string, v any) (map[string]any, error) {
	if err := maputil.SetByPath(&obj, k, v); err != nil {
		return nil, err
	}

	return obj, nil
}
