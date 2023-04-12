package arr

import (
	"reflect"
	"sort"
)

// Accessible Determine whether the given value is array accessible.
func Accessible(value interface{}) bool {
	_, isArray := value.([]interface{})
	return isArray
}

// Add an element to an array using “dot” notation if it doesn't exist.
func Add(arr []any, key int, value any) ([]any, error) {
	if Get(arr, key, nil) == nil {
		if err := Set(&arr, key, value); err != nil {
			return arr, err
		}
	}

	return arr, nil
}

// Collapse an array of arrays into a single array.
func Collapse(arr []interface{}) []interface{} {
	if len(arr) == 0 {
		return []interface{}{}
	}
	var res []interface{}

	for _, values := range arr {
		switch v := values.(type) {
		case []interface{}:
			res = append(res, Collapse(v)...)
		default:
			res = append(res, v)
		}
	}
	return res
}

// CrossJoin returns all possible permutations of the given arrays.
func CrossJoin(arr ...[]any) ([][]any, error) {
	if len(arr) == 0 {
		return nil, ErrArrayRequired
	}

	res := [][]any{{}}

	for _, array := range arr {
		if len(array) == 0 {
			return nil, ErrEmptyArrayNotAllowed
		}

		apd := [][]any{}

		for _, product := range res {
			for _, item := range array {
				productCopy := make([]any, len(product))
				copy(productCopy, product)

				apd = append(apd, append(productCopy, item))
			}
		}

		res = apd
	}

	return res, nil
}

// Divide an array into two arrays. One with keys and the other with values.
func Divide(arr []any) ([]any, []any, error) {
	if len(arr) == 0 {
		return nil, nil, ErrEmptyArrayNotAllowed
	}

	keys := make([]any, len(arr))
	values := make([]any, len(arr))

	for i, v := range arr {
		keys[i] = i
		values[i] = v
	}

	return keys, values, nil
}

// Dot returns a flattened associative array with dot notation
func Dot(m []interface{}, prefix string) ([]interface{}, error) {
	return nil, ErrNoImplementation
}

// Undot returns an expanded array from flattened dot notation array
func Undot(m []interface{}) ([]interface{}, error) {
	return nil, ErrNoImplementation
}

// Except returns all of the given array except for a specified array of keys.
func Except[T any](arr []T, keys []int) []T {
	excludedKeys := make(map[int]bool)

	for _, key := range keys {
		excludedKeys[key] = true
	}

	res := make([]T, 0, len(arr))

	for key, item := range arr {
		if excludedKeys[key] {
			continue
		}

		res = append(res, item)
	}

	return res
}

// Exists determines if the given key exists in the provided array.
func Exists(arr []interface{}, key int) bool {
	if key < 0 {
		return false
	}

	if key > len(arr)-1 {
		return false
	}

	return true
}

// First returns the first element in an array that passes a given truth test.
func First[T any](arr []T, callback func(T, int) bool, def T) T {
	if callback == nil {
		if len(arr) == 0 {
			return def
		}
		return arr[0]
	}

	for i, v := range arr {
		if callback(v, i) {
			return v
		}
	}

	return def
}

// Last returns the last element in an array passing a given truth test.
func Last[T any](arr []T, callback func(T) bool, defaultValue T) T {
	if callback == nil {
		if len(arr) == 0 {
			return defaultValue
		}

		return arr[len(arr)-1]
	}

	for i := len(arr) - 1; i >= 0; i-- {
		if callback(arr[i]) {
			return arr[i]
		}
	}

	return defaultValue
}

// Flatten flattens a multi-dimensional array into a single level.
func Flatten(arr []interface{}, depth int) []interface{} {
	var result []interface{}

	for _, item := range arr {
		if !isArray(item) {
			result = append(result, item)
		} else {
			values := make([]interface{}, 0)

			if depth == 1 {
				for _, v := range item.([]interface{}) {
					values = append(values, v)
				}
			} else {
				values = Flatten(item.([]interface{}), depth-1)
			}

			result = append(result, values...)
		}
	}

	return result
}

// Forget Remove one or many array items from a given array.
func Forget[T any](arr []T, keys interface{}) ([]T, error) {
	if len(arr) == 0 || keys == nil {
		return arr, nil
	}

	switch v := keys.(type) {
	case int:
		keys = []int{v}
	case []int:
		sort.Sort(sort.Reverse(sort.IntSlice(v)))
	default:
		return arr, ErrInvalidKeys
	}

	for _, key := range keys.([]int) {
		if key >= 0 && key < len(arr) {
			copy(arr[key:], arr[key+1:])
			arr[len(arr)-1] = reflect.Zero(reflect.TypeOf(arr[0])).Interface().(T)
			arr = arr[:len(arr)-1]

			continue
		}
	}
	return arr, nil
}

// Get an item from an array using int key.
func Get[T any](arr []T, key int, def T) T {
	if key < 0 {
		return def
	}

	if key > len(arr)-1 {
		return def
	}

	return arr[key]
}

// Has checks if an item or items exist in an array using "dot" notation.
func Has[T any](arr []T, keys interface{}) bool {
	if len(arr) == 0 || keys == nil {
		return false
	}

	switch v := keys.(type) {
	case int:
		keys = []int{v}
	case []int:
	default:
		return false
	}

	for _, key := range keys.([]int) {
		if key >= 0 && key < len(arr) {
			return true
		}
	}

	return false
}

// todo: hasAny($array, $keys)
// todo: isAssoc(array $array)
// todo: isList($array)
// todo: join($array, $glue, $finalGlue = '')
// todo: keyBy($array, $keyBy)
// todo: prependKeysWith($array, $prependWith)
// todo: only($array, $keys)
// todo: pluck($array, $value, $key = null)
// todo: explodePluckParameters($value, $key)

// Map Run a map over each of the items in the array.
func Map[T, U any](arr []T, fn func(T, int) U) []U {
	res := make([]U, len(arr))
	for i, v := range arr {
		res[i] = fn(v, i)
	}
	return res
}

// todo: prepend($array, $value, $key = null)
// todo: pull(&$array, $key, $default = null)
// todo: query($array)
// todo: random($array, $number = null, $preserveKeys = false)

// Set an array item to a given value using int key
func Set[T any](arr *[]T, key int, value T) error {
	if key < 0 {
		return ErrInvalidKey
	}

	if key >= len(*arr) {
		newSlice := make([]T, key+1)
		copy(newSlice, *arr)
		*arr = newSlice
	}

	(*arr)[key] = value
	return nil
}

// todo: shuffle($array, $seed = null)
// todo: sort($array, $callback = null)
// todo: sortDesc($array, $callback = null)
// todo: sortRecursive($array, $options = SORT_REGULAR, $descending = false)
// todo: toCssClasses($array)
// todo: toCssStyles($array)
// todo: where($array, callable $callback)
// todo: whereNotNull($array)
// todo: wrap($value)

// IsArray determines whether the given value is an array.
func isArray(arr interface{}) bool {
	_, ok := arr.([]interface{})
	return ok
}
