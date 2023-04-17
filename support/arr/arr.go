package arr

import (
	"fmt"
	"github.com/goravel/framework/support/time"
	"math/rand"
	"reflect"
	"sort"
	"strings"
)

// Accessible Determine whether the given value is array accessible.
func Accessible[T any](value T) bool {
	return reflect.TypeOf(value).Kind() == reflect.Slice || reflect.TypeOf(value).Kind() == reflect.Array
}

// Add an element to an array using “dot” notation if it doesn't exist.
func Add[T any](arr []T, key int, value T) ([]T, error) {
	_, found := Get(arr, key, false)
	if !found {
		if err := Set(&arr, key, value); err != nil {
			return arr, err
		}
	}

	return arr, nil
}

// Collapse collapses an array of arrays into a single array.
func Collapse[T any](arr []T) []T {
	if len(arr) == 0 {
		return []T{}
	}
	var res []T

	for _, v := range arr {
		if Accessible(v) {
			res = append(res, Collapse(reflect.ValueOf(v).Interface().([]T))...)
		} else {
			res = append(res, v)
		}
	}
	return res
}

// CrossJoin returns all possible permutations of the given arrays.
func CrossJoin[T any](arr ...[]T) ([][]T, error) {
	if len(arr) == 0 {
		return nil, ErrArrayRequired
	}

	res := [][]T{{}}

	for _, v := range arr {
		if len(v) == 0 {
			return nil, ErrEmptySliceNotAllowed
		}

		var apd [][]T

		for _, product := range res {
			for _, item := range v {
				productCopy := make([]T, len(product))
				copy(productCopy, product)

				apd = append(apd, append(productCopy, item))
			}
		}

		res = apd
	}

	return res, nil
}

// Divide an array into two arrays. One with keys and the other with values.
func Divide[T any](arr []T) ([]int, []T, error) {
	if len(arr) == 0 {
		return nil, nil, ErrEmptySliceNotAllowed
	}

	keys := make([]int, len(arr))
	values := make([]T, len(arr))

	for i, v := range arr {
		keys[i] = i
		values[i] = v
	}

	return keys, values, nil
}

// Dot returns a flattened associative array with dot notation
func Dot[T any](m []T, prefix string) ([]T, error) {
	return nil, ErrNoImplementation
}

// Undot returns an expanded array from flattened dot notation array
func Undot[T any](m []T) ([]T, error) {
	return nil, ErrNoImplementation
}

// Except returns all the given array except for a specified array of keys.
func Except[T any](arr []T, keys []int) []T {
	excludedKeys := make(map[int]bool)

	for _, v := range keys {
		excludedKeys[v] = true
	}

	res := make([]T, 0, len(arr))

	for i, v := range arr {
		if excludedKeys[i] {
			continue
		}

		res = append(res, v)
	}

	return res
}

// Exists determines if the given key exists in the provided array.
func Exists[T any](arr []T, key int) bool {
	return key >= 0 && key < len(arr)
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
func Last[T any](arr []T, callback func(T) bool, def T) T {
	if callback == nil {
		if len(arr) == 0 {
			return def
		}

		return arr[len(arr)-1]
	}

	for i := len(arr) - 1; i >= 0; i-- {
		if callback(arr[i]) {
			return arr[i]
		}
	}

	return def
}

// Flatten flattens a multi-dimensional array into a single level.
// todo: generic
func Flatten(arr []any, depth int) []any {
	var res []any

	for _, v := range arr {
		if !Accessible(v) {
			res = append(res, v)
		} else {
			val := make([]any, 0)

			if depth == 0 {
				for _, v := range v.([]any) {
					val = append(val, v)
				}
			} else {
				val = Flatten(v.([]any), depth-1)
			}

			res = append(res, val...)
		}
	}

	return res
}

// Forget Remove one or many array items from a given array.
func Forget[T any](arr []T, keys any) ([]T, error) {
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

	for _, v := range keys.([]int) {
		if v >= 0 && v < len(arr) {
			copy(arr[v:], arr[v+1:])
			arr[len(arr)-1] = reflect.Zero(reflect.TypeOf(arr[0])).Interface().(T)
			arr = arr[:len(arr)-1]

			continue
		}
	}
	return arr, nil
}

// Get an item from an array using int key.
// todo: thread safe?
func Get[T any](arr []T, key int, def any) (T, bool) {
	if key < 0 || key > len(arr)-1 {
		return def.(T), false
	}

	return arr[key], true
}

// Has checks if an item or items exist in an array using "dot" notation.
func Has[T any](arr []T, keys any) bool {
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

	for _, v := range keys.([]int) {
		if v < 0 || v >= len(arr) {
			return false
		}
	}

	return true
}

// HasAny Determine if any of the keys exist in an array using int key
func HasAny[T any](arr []T, keys any) bool {
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

	for _, v := range keys.([]int) {
		if v >= 0 && v < len(arr) {
			return true
		}
	}

	return false
}

// IsAssoc Determines if an array is associative.
func IsAssoc[T any](arr T) bool {
	return false
}

// IsList Determines if an array is a list.
func IsList[T any](arr T) bool {
	return true
}

// Join concatenates elements of a slice into a string with a specified delimiter and final separator
func Join[T any](arr []T, delimiter string, finalSeparator ...string) string {
	l := len(arr)
	if l == 0 {
		return ""
	}
	if l == 1 {
		return fmt.Sprint(arr[0])
	}
	if l == 2 {
		return fmt.Sprintf("%v%s%v", arr[0], finalSeparator[0], arr[1])
	}
	if len(finalSeparator) == 0 {
		finalSeparator = []string{delimiter}
	}

	var builder strings.Builder
	for i, v := range arr {
		builder.WriteString(fmt.Sprint(v))
		if i == len(arr)-2 {
			builder.WriteString(finalSeparator[0])
		} else if i < len(arr)-1 {
			builder.WriteString(delimiter)
		}
	}
	return builder.String()
}

// KeyBy Key an associative array by a field or using a callback.
func KeyBy[T any, K comparable](array []T, key K) ([]T, error) {
	return nil, ErrNoImplementation
}

// PrependKeysWith Prepend the key names of an associative array.
func PrependKeysWith[T any](array []T, prependWith string) ([]T, error) {
	return nil, ErrNoImplementation
}

// Only returns a subset of the items from the given map with specified keys.
func Only[T any](arr []T, keys any) []T {
	var res []T
	switch v := keys.(type) {
	case int:
		if v < 0 || v > len(arr)-1 {
			return res
		}
		res = append(res, arr[v])

	case []int:
		for _, vv := range v {
			if vv < 0 || vv > len(arr)-1 {
				return res
			}
			res = append(res, arr[vv])
		}
	}
	return res
}

// Pluck an array of values from an array.
func Pluck[T any](array []T, value T, key ...int) ([]T, error) {
	return nil, ErrNoImplementation
}

// ExplodePluckParameters Explode the "value" and "key" arguments passed to "pluck".
func ExplodePluckParameters[T any](array []T, key ...int) ([]T, error) {
	return nil, ErrNoImplementation
}

// Map Run a map over each of the items in the array.
func Map[T, U any](arr []T, fn func(T, int) U) []U {
	res := make([]U, len(arr))
	for i, v := range arr {
		res[i] = fn(v, i)
	}
	return res
}

// Prepend the given value to the beginning of an array or associative array.
func Prepend[T any](arr []T, value T) []T {
	return append([]T{value}, arr...)
}

// Pull Get a value from the array, and remove it.
func Pull[T any](arr []T, key int, def T) ([]T, T, error) {
	v, _ := Get(arr, key, def)

	res, err := Forget(arr, key)
	return res, v, err
}

// Query Convert the array into a query string.
func Query[T any](m []T) (*string, error) {
	return nil, ErrNoImplementation
}

// Random returns one or a specified number of random values from a slice.
func Random[T any](arr []T, number *int) ([]T, error) {
	//rand.Seed(time.Now().UnixNano())

	requested := 1
	if number != nil {
		requested = *number
	}

	count := len(arr)

	if requested > count {
		return nil, ErrExceedMaxLength
	}

	if number == nil {
		return []T{arr[rand.Intn(count)]}, nil
	}

	if requested == 0 {
		return []T{}, nil
	}

	indices := rand.Perm(count)[:requested]

	res := make([]T, 0, requested)
	for _, index := range indices {
		res = append(res, arr[index])
	}

	return res, nil
}

// Set an array item to a given value using int key
// todo: thread safe?
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

// Shuffle the given array and return the result.
func Shuffle[T any](arr []T, seed *int64) []T {
	res := make([]T, len(arr))
	copy(res, arr)

	var r *rand.Rand
	if seed == nil {
		randSeed := time.Now().UnixNano()
		r = rand.New(rand.NewSource(randSeed))
	} else {
		r = rand.New(rand.NewSource(*seed))
	}
	r.Shuffle(len(res), func(i, j int) { res[i], res[j] = res[j], res[i] })
	return res
}

// Sort the nested array using the given callback.
// todo: generic
func Sort(arr []any, fn func(i, j int) bool) []any {
	if len(arr) == 0 {
		return arr
	}

	sort.Slice(arr, func(i, j int) bool {
		return fn(i, j)
	})

	for i, v := range arr {
		switch val := v.(type) {
		case []any:
			arr[i] = Sort(val, fn)
		default:
		}
	}

	return arr
}

// Sort the nested array in descending order using the given callback.
// todo: generic
func SortDesc(arr []any, fn func(i, j int) bool) []any {
	return Sort(arr, fn)
}

// SortRecursive Recursively sort an array by values.
// todo: generic
func SortRecursive(arr []any, descending bool) ([]any, error) {
	if len(arr) == 0 {
		return arr, nil
	}

	res := make([]any, len(arr))
	copy(res, arr)

	var err error
	for i, v := range res {
		if Accessible(v) {
			vv := reflect.ValueOf(v)
			subSlice := make([]any, vv.Len())
			for j := 0; j < vv.Len(); j++ {
				subSlice[j] = vv.Index(j).Interface()
			}
			res[i], err = SortRecursive(subSlice, descending)
			if err != nil {
				return arr, err
			}
		}
	}

	fn, err := generateLessFunc(arr)
	if err != nil {
		return res, nil
	}

	sort.Slice(res, func(i, j int) bool {
		if !descending {
			return fn(res[i], res[j])
		} else {
			return fn(res[j], res[i])
		}
	})

	return res, nil
}

// ToCssClasses Convert an array of strings to a string of CSS classes.
func ToCssClasses[T any](arr []T) string {
	var res []string

	for _, v := range arr {
		res = append(res, fmt.Sprint(v))
	}

	return strings.Join(res, " ")
}

// ToCssStyles Convert an array of strings to a string of CSS styles.
func ToCssStyles[T any](arr []T) string {
	var res []string

	for _, v := range arr {
		res = append(res, fmt.Sprint(v))
	}

	return strings.Join(res, "; ")
}

// Where Filter the array using the given callback.
func Where[T any](arr []T, fn func(T) bool) []T {
	res := make([]T, 0)

	for _, v := range arr {
		if fn(v) {
			res = append(res, v)
		}
	}

	return res
}

// WhereNotNull Filter items where the value is not null.
func WhereNotNull[T any](arr []T) []T {
	notNilFilter := func(item T) bool {
		rv := reflect.ValueOf(item)
		return rv.IsValid() && !rv.IsZero()
	}
	return Where(arr, notNilFilter)
}

// Wrap If the given value is not an array and not null, wrap it in one.
// todo: generic
func Wrap(value any) []any {
	if value == nil {
		return []interface{}{}
	}

	if Accessible(value) {
		v := reflect.ValueOf(value)
		slice := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			slice[i] = v.Index(i).Interface().(any)
		}
		return slice
	}

	return []any{value}
}

// generateLessFunc return a comparison func for sorting the elements based on their type
func generateLessFunc[T any](arr []T) (func(a, b T) bool, error) {
	if len(arr) == 0 {
		return nil, ErrEmptySliceNotAllowed
	}

	return func(a, b T) bool {
		if reflect.TypeOf(a) != reflect.TypeOf(b) {
			return false
		}

		switch reflect.TypeOf(a).Kind() {
		case reflect.Int:
			ai := reflect.ValueOf(a).Interface().(int)
			bi := reflect.ValueOf(b).Interface().(int)
			return ai < bi
		case reflect.Float64:
			af := reflect.ValueOf(a).Interface().(float64)
			bf := reflect.ValueOf(b).Interface().(float64)
			return af < bf
		case reflect.String:
			as := reflect.ValueOf(a).Interface().(string)
			bs := reflect.ValueOf(b).Interface().(string)
			return as < bs
		}
		return false
	}, nil
}
