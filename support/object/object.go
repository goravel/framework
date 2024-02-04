package object

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/gookit/goutil/maputil"
	"github.com/gookit/goutil/reflects"
)

const (
	Wildcard = "*"
	PathSep  = "."
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
	for _, key := range keys {
		if _, ok := obj[key]; ok {
			delete(obj, key)
			continue
		}
		deleteByPath(obj, obj, strings.Split(key, PathSep))
	}
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
	if val, ok := obj[key]; ok {
		delete(obj, key)
		return val
	}

	value, ok := deleteByPath(obj, obj, strings.Split(key, PathSep))

	if !ok && len(def) > 0 {
		return def[0]
	}

	return value
}

// Set an element to a map using “dot” notation.
func Set(obj *map[string]any, k string, v any) error {
	return maputil.SetByPath(obj, k, v)
}

// deleteByPath delete value by key path from a map(map[string]any).eg "top" "top.sub"
//
// Example:
//
//	mp := map[string]any{
//		"top": map[string]any{
//			"sub": "value",
//		},
//	}
//	val, ok := deleteByPath(mp, []string{"top", "sub"}) // return "value", true
//
// Inspired by GetByPathKeys function from https://github.com/gookit/goutil/maputil package
func deleteByPath(parent, child any, keys []string) (any, bool) {
	var (
		prevLevel, currLevel any
		ok                   bool
		prevKey              string
	)

	kl := len(keys)

	prevLevel = parent
	currLevel = child

	for i, k := range keys {
		switch tData := currLevel.(type) {
		case map[string]string:
			if _, ok = tData[k]; !ok {
				return nil, false
			}
			prevLevel = currLevel
			currLevel = tData[k]
			prevKey = k
			if kl == i+1 {
				delete(tData, k)
				return currLevel, true
			}
		case map[string]any:
			if _, ok = tData[k]; !ok {
				return nil, false
			}
			prevLevel = currLevel
			currLevel = tData[k]
			prevKey = k
			if kl == i+1 {
				delete(tData, k)
				return currLevel, true
			}
		case map[any]any:
			// Check if the key exists in the map
			if val, ok := tData[k]; ok {
				prevLevel = currLevel
				currLevel = val
				prevKey = k

				// If it's the last key, delete it and return the current level
				if kl == i+1 {
					delete(tData, k)
					return currLevel, true
				}
			} else {
				// Try converting the key to an integer
				if idx, err := strconv.Atoi(k); err == nil {
					prevLevel = currLevel
					currLevel = tData[idx]
					prevKey = k

					// If it's the last key, delete it and return the current level
					if kl == i+1 {
						delete(tData, idx)
						return currLevel, true
					}
				}
			}
		case map[string]int:
			if _, ok = tData[k]; !ok {
				return nil, false
			}
			prevLevel = currLevel
			currLevel = tData[k]
			prevKey = k
			if kl == i+1 {
				delete(tData, k)
				return currLevel, true
			}
		default:
			rv := reflect.ValueOf(tData)
			// check is slice
			if rv.Kind() == reflect.Slice {
				if k == Wildcard {
					if kl == i+1 { // * is last key
						rv = rv.Slice(0, 0)
						reflect.ValueOf(prevLevel).SetMapIndex(reflect.ValueOf(prevKey), rv)
						return currLevel, true
					}

					for si := 0; si < rv.Len(); si++ {
						el := reflects.Indirect(rv.Index(si))
						if el.Kind() != reflect.Map {
							return nil, false
						}

						// el is map value.
						if _, ok := deleteByPath(prevLevel, el.Interface(), keys[i+1:]); ok {
							if reflects.IsEmpty(el) {
								if rv.Len() > 1 {
									rv = reflect.AppendSlice(rv.Slice(0, si), rv.Slice(si+1, rv.Len()))
									si--
								} else {
									rv = reflect.MakeSlice(rv.Type(), 0, 0)
									break
								}
							}
						}
					}
					currLevel = rv.Interface()
					reflect.ValueOf(prevLevel).SetMapIndex(reflect.ValueOf(prevKey), rv)
					if rv.Len() > 0 {
						return currLevel, true
					}

					return nil, false
				}

				ii, err := strconv.Atoi(k)
				if err != nil || ii >= rv.Len() {
					return nil, false
				}

				currLevel = rv.Index(ii).Interface()

				if kl == i+1 {
					rv = reflect.AppendSlice(rv.Slice(0, ii), rv.Slice(ii+1, rv.Len()))
					reflect.ValueOf(prevLevel).SetMapIndex(reflect.ValueOf(prevKey), rv)
					return currLevel, true
				}
				continue
			}
		}
	}

	return currLevel, false
}
