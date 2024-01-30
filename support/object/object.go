package object

import (
	"github.com/gookit/goutil/maputil"
)

// Add an element to a map using “dot” notation if it doesn't exist.
func Add(obj *map[string]any, k string, v any) error {
	if val := Get(*obj, k); val != nil {
		return nil
	}

	return Set(obj, k, v)
}

// Dot flattens a map using dot notation.
func Dot(obj map[string]any) map[string]any {
	return maputil.Flatten(obj)
}

// Exists checks if the given key exists in the provided map (only top level).
func Exists[K comparable, V any](obj map[K]V, key K) bool {
	_, ok := obj[key]
	return ok
}

// Forget removes a given key or keys from the provided map.
func Forget(obj map[string]any, keys ...string) {
	// need to finish
}

// Get an item from an object using "dot" notation.
func Get(obj map[string]any, key string, defaults ...any) any {
	val, ok := maputil.GetByPath(key, obj)

	if !ok && len(defaults) > 0 {
		return defaults[0]
	}

	return val
}

// Has checks if the given key or keys exist in the provided map.
func Has(obj map[string]any, keys ...string) bool {
	if len(keys) == 0 || len(obj) == 0 {
		return false
	}

	for _, key := range keys {
		_, ok := maputil.GetByPath(key, obj)
		if !ok {
			return false
		}
	}

	return true
}

// HasAny checks if the given key or keys exist in the provided map.
func HasAny(obj map[string]any, keys ...string) bool {
	for _, key := range keys {
		if Has(obj, key) {
			return true
		}
	}

	return false
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
func Set(obj *map[string]any, k string, v any) error {
	return maputil.SetByPath(obj, k, v)
}
