package collect

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Collection[T any] struct {
	items []T
}

func (c *Collection[T]) After(value T) *T {
	for i, item := range c.items {
		if reflect.DeepEqual(item, value) && i+1 < len(c.items) {
			return &c.items[i+1]
		}
	}
	return nil
}

func (c *Collection[T]) All() []T {
	return c.items
}

func (c *Collection[T]) Average(keyFunc func(T) float64) float64 {
	if len(c.items) == 0 {
		return 0
	}
	return c.Sum(keyFunc) / float64(len(c.items))
}

func (c *Collection[T]) Avg(keyFunc func(T) float64) float64 {
	return c.Average(keyFunc)
}

func (c *Collection[T]) Before(value T) *T {
	for i, item := range c.items {
		if reflect.DeepEqual(item, value) && i > 0 {
			return &c.items[i-1]
		}
	}
	return nil
}

func (c *Collection[T]) Chunk(size int) [][]T {
	if size <= 0 {
		return [][]T{}
	}

	var chunks [][]T
	for i := 0; i < len(c.items); i += size {
		end := i + size
		if end > len(c.items) {
			end = len(c.items)
		}
		chunks = append(chunks, c.items[i:end])
	}
	return chunks
}

func (c *Collection[T]) ChunkWhile(predicate func(T, int, []T) bool) [][]T {
	if len(c.items) == 0 {
		return [][]T{}
	}

	var chunks [][]T
	var currentChunk []T

	for i, item := range c.items {
		if len(currentChunk) == 0 || predicate(item, i, currentChunk) {
			currentChunk = append(currentChunk, item)
		} else {
			chunks = append(chunks, currentChunk)
			currentChunk = []T{item}
		}
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

func (c *Collection[T]) Clone() *Collection[T] {
	cloned := make([]T, len(c.items))
	copy(cloned, c.items)
	return &Collection[T]{items: cloned}
}

func (c *Collection[T]) Collapse() *Collection[T] {
	var flattened []T
	for _, item := range c.items {
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Slice {
			for i := 0; i < v.Len(); i++ {
				if elem, ok := v.Index(i).Interface().(T); ok {
					flattened = append(flattened, elem)
				}
			}
		} else {
			flattened = append(flattened, item)
		}
	}
	return &Collection[T]{items: flattened}
}

func Collect[T any](items []T) *Collection[T] {
	return &Collection[T]{items: items}
}

func (c *Collection[T]) Combine(keys []string) map[string]T {
	result := make(map[string]T)
	for i, key := range keys {
		if i < len(c.items) {
			result[key] = c.items[i]
		}
	}
	return result
}

func (c *Collection[T]) Concat(other *Collection[T]) *Collection[T] {
	merged := make([]T, len(c.items)+len(other.items))
	copy(merged, c.items)
	copy(merged[len(c.items):], other.items)
	return &Collection[T]{items: merged}
}

func (c *Collection[T]) Contains(value T) bool {
	for _, item := range c.items {
		if reflect.DeepEqual(item, value) {
			return true
		}
	}
	return false
}

func (c *Collection[T]) ContainsStrict(value T) bool {
	for _, item := range c.items {
		if reflect.DeepEqual(item, value) && reflect.TypeOf(item) == reflect.TypeOf(value) {
			return true
		}
	}
	return false
}

func (c *Collection[T]) Count() int {
	return len(c.items)
}

func (c *Collection[T]) CountBy(keyFunc func(T) string) map[string]int {
	counts := make(map[string]int)
	for _, item := range c.items {
		key := keyFunc(item)
		counts[key]++
	}
	return counts
}

func (c *Collection[T]) CrossJoin(other *Collection[T]) [][]T {
	var result [][]T
	for _, item1 := range c.items {
		for _, item2 := range other.items {
			result = append(result, []T{item1, item2})
		}
	}
	return result
}

func (c *Collection[T]) Debug() *Collection[T] {
	fmt.Printf("Collection contents: %+v\n", c.items)
	return c
}

func (c *Collection[T]) Diff(other *Collection[T]) *Collection[T] {
	otherMap := make(map[string]bool)
	for _, item := range other.items {
		otherMap[fmt.Sprintf("%v", item)] = true
	}

	var diff []T
	for _, item := range c.items {
		if !otherMap[fmt.Sprintf("%v", item)] {
			diff = append(diff, item)
		}
	}

	return &Collection[T]{items: diff}
}

func (c *Collection[T]) DiffAssoc(other *Collection[T]) *Collection[T] {
	var diff []T
	for i, item := range c.items {
		if i >= len(other.items) || !reflect.DeepEqual(item, other.items[i]) {
			diff = append(diff, item)
		}
	}
	return &Collection[T]{items: diff}
}

func (c *Collection[T]) DiffKeys(other *Collection[T]) *Collection[T] {
	var diff []T
	for i, item := range c.items {
		if i >= len(other.items) {
			diff = append(diff, item)
		}
	}
	return &Collection[T]{items: diff}
}

func (c *Collection[T]) Doesnt(value T) bool {
	return !c.Contains(value)
}

func (c *Collection[T]) Dot() map[string]interface{} {
	result := make(map[string]interface{})
	for i, item := range c.items {
		result[strconv.Itoa(i)] = item
	}
	return result
}

func (c *Collection[T]) Drop(n int) *Collection[T] {
	if n >= len(c.items) {
		return &Collection[T]{items: []T{}}
	}
	return &Collection[T]{items: c.items[n:]}
}

func (c *Collection[T]) DropUntil(predicate func(T) bool) *Collection[T] {
	for i, item := range c.items {
		if predicate(item) {
			return &Collection[T]{items: c.items[i:]}
		}
	}
	return &Collection[T]{items: []T{}}
}

func (c *Collection[T]) DropWhile(predicate func(T) bool) *Collection[T] {
	for i, item := range c.items {
		if !predicate(item) {
			return &Collection[T]{items: c.items[i:]}
		}
	}
	return &Collection[T]{items: []T{}}
}

func (c *Collection[T]) Dump() *Collection[T] {
	fmt.Printf("Collection: %+v\n", c.items)
	return c
}

func (c *Collection[T]) Duplicates() *Collection[T] {
	seen := make(map[string]bool)
	var duplicates []T

	for _, item := range c.items {
		key := fmt.Sprintf("%v", item)
		if seen[key] {
			duplicates = append(duplicates, item)
		} else {
			seen[key] = true
		}
	}
	return &Collection[T]{items: duplicates}
}

func (c *Collection[T]) Each(fn func(T, int)) *Collection[T] {
	for i, item := range c.items {
		fn(item, i)
	}
	return c
}

func (c *Collection[T]) EachSpread(fn func(...T)) *Collection[T] {
	for _, item := range c.items {
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Slice {
			args := make([]T, v.Len())
			for i := 0; i < v.Len(); i++ {
				if elem, ok := v.Index(i).Interface().(T); ok {
					args[i] = elem
				}
			}
			fn(args...)
		} else {
			fn(item)
		}
	}
	return c
}

func (c *Collection[T]) EachUntil(fn func(T, int) bool) *Collection[T] {
	for i, item := range c.items {
		if !fn(item, i) {
			break
		}
	}
	return c
}

func (c *Collection[T]) Every(predicate func(T) bool) bool {
	for _, item := range c.items {
		if !predicate(item) {
			return false
		}
	}
	return true
}

func (c *Collection[T]) Except(indices ...int) *Collection[T] {
	excludeMap := make(map[int]bool)
	for _, index := range indices {
		excludeMap[index] = true
	}

	var result []T
	for i, item := range c.items {
		if !excludeMap[i] {
			result = append(result, item)
		}
	}
	return &Collection[T]{items: result}
}

func (c *Collection[T]) Filter(predicate func(T, int) bool) *Collection[T] {
	var filtered []T
	for i, item := range c.items {
		if predicate(item, i) {
			filtered = append(filtered, item)
		}
	}
	return &Collection[T]{items: filtered}
}

func (c *Collection[T]) First() *T {
	if len(c.items) == 0 {
		return nil
	}
	return &c.items[0]
}

func (c *Collection[T]) FirstOrFail() (*T, error) {
	if len(c.items) == 0 {
		return nil, fmt.Errorf("collection is empty")
	}
	return &c.items[0], nil
}

func (c *Collection[T]) FirstWhere(predicate func(T) bool) *T {
	for _, item := range c.items {
		if predicate(item) {
			return &item
		}
	}
	return nil
}

func (c *Collection[T]) FlatMap(fn func(T) []T) *Collection[T] {
	var result []T
	for _, item := range c.items {
		result = append(result, fn(item)...)
	}
	return &Collection[T]{items: result}
}

func (c *Collection[T]) Flatten() *Collection[T] {
	return c.Collapse()
}

func (c *Collection[T]) Flip() map[string]string {
	result := make(map[string]string)
	for i, item := range c.items {
		result[fmt.Sprintf("%v", item)] = strconv.Itoa(i)
	}
	return result
}

func (c *Collection[T]) ForPage(page, perPage int) *Collection[T] {
	start := (page - 1) * perPage
	if start >= len(c.items) {
		return &Collection[T]{items: []T{}}
	}

	end := start + perPage
	if end > len(c.items) {
		end = len(c.items)
	}

	return &Collection[T]{items: c.items[start:end]}
}

func (c *Collection[T]) Forget(indices ...int) *Collection[T] {
	return c.Except(indices...)
}

func (c *Collection[T]) Get(index int) *T {
	if index < 0 || index >= len(c.items) {
		return nil
	}
	return &c.items[index]
}

func (c *Collection[T]) GroupBy(keyFunc func(T) string) map[string]*Collection[T] {
	groups := make(map[string]*Collection[T])

	for _, item := range c.items {
		key := keyFunc(item)
		if _, exists := groups[key]; !exists {
			groups[key] = &Collection[T]{items: []T{}}
		}
		groups[key].items = append(groups[key].items, item)
	}

	return groups
}

func (c *Collection[T]) Has(index int) bool {
	return index >= 0 && index < len(c.items)
}

func (c *Collection[T]) HasAny(indices ...int) bool {
	for _, index := range indices {
		if c.Has(index) {
			return true
		}
	}
	return false
}

func (c *Collection[T]) Implode(separator string) string {
	var parts []string
	for _, item := range c.items {
		parts = append(parts, fmt.Sprintf("%v", item))
	}
	return strings.Join(parts, separator)
}

func (c *Collection[T]) Intersect(other *Collection[T]) *Collection[T] {
	otherMap := make(map[string]bool)
	for _, item := range other.items {
		otherMap[fmt.Sprintf("%v", item)] = true
	}

	var intersection []T
	for _, item := range c.items {
		if otherMap[fmt.Sprintf("%v", item)] {
			intersection = append(intersection, item)
		}
	}

	return &Collection[T]{items: intersection}
}

func (c *Collection[T]) IntersectByKeys(other *Collection[T]) *Collection[T] {
	var intersection []T
	for i, item := range c.items {
		if i < len(other.items) {
			intersection = append(intersection, item)
		}
	}
	return &Collection[T]{items: intersection}
}

func (c *Collection[T]) IsEmpty() bool {
	return len(c.items) == 0
}

func (c *Collection[T]) IsNotEmpty() bool {
	return !c.IsEmpty()
}

func (c *Collection[T]) Join(separator string) string {
	return c.Implode(separator)
}

func (c *Collection[T]) KeyBy(keyFunc func(T) string) map[string]T {
	result := make(map[string]T)
	for _, item := range c.items {
		key := keyFunc(item)
		result[key] = item
	}
	return result
}

func (c *Collection[T]) Keys() []int {
	keys := make([]int, len(c.items))
	for i := range c.items {
		keys[i] = i
	}
	return keys
}

func (c *Collection[T]) Last() *T {
	if len(c.items) == 0 {
		return nil
	}
	return &c.items[len(c.items)-1]
}

func (c *Collection[T]) LastOrFail() (*T, error) {
	if len(c.items) == 0 {
		return nil, fmt.Errorf("collection is empty")
	}
	return &c.items[len(c.items)-1], nil
}

func (c *Collection[T]) Make(items ...T) *Collection[T] {
	return &Collection[T]{items: items}
}

func (c *Collection[T]) Map(fn func(T, int) interface{}) *Collection[interface{}] {
	mapped := make([]interface{}, len(c.items))
	for i, item := range c.items {
		mapped[i] = fn(item, i)
	}
	return &Collection[interface{}]{items: mapped}
}

// Reduce reduces the collection to a single value using the given reducer function
func (c *Collection[T]) Reduce(fn func(acc interface{}, item T, index int) interface{}, initial interface{}) interface{} {
	acc := initial
	for i, item := range c.items {
		acc = fn(acc, item, i)
	}
	return acc
}

// MapInto maps each element of the collection using reflection to cast to target type
// and returns a new collection of the target type
func (c *Collection[T]) MapInto(target interface{}) *Collection[interface{}] {
	targetType := reflect.TypeOf(target)
	mapped := make([]interface{}, len(c.items))
	
	for i, item := range c.items {
		itemValue := reflect.ValueOf(item)
		
		// Try to convert the item to the target type
		if itemValue.Type().ConvertibleTo(targetType) {
			converted := itemValue.Convert(targetType)
			mapped[i] = converted.Interface()
		} else {
			// If not convertible, keep original value
			mapped[i] = item
		}
	}
	
	return &Collection[interface{}]{items: mapped}
}

func (c *Collection[T]) MapSpread(fn func(...T) T) *Collection[T] {
	var result []T
	for _, item := range c.items {
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Slice {
			args := make([]T, v.Len())
			for i := 0; i < v.Len(); i++ {
				if elem, ok := v.Index(i).Interface().(T); ok {
					args[i] = elem
				}
			}
			result = append(result, fn(args...))
		} else {
			result = append(result, fn(item))
		}
	}
	return &Collection[T]{items: result}
}

func (c *Collection[T]) MapToDictionary(keyFunc func(T) string) map[string][]T {
	result := make(map[string][]T)
	for _, item := range c.items {
		key := keyFunc(item)
		result[key] = append(result[key], item)
	}
	return result
}

func (c *Collection[T]) MapToGroups(keyFunc func(T) string) map[string]*Collection[T] {
	return c.GroupBy(keyFunc)
}

func (c *Collection[T]) MapWithKeys(fn func(T) (string, T)) map[string]T {
	result := make(map[string]T)
	for _, item := range c.items {
		key, value := fn(item)
		result[key] = value
	}
	return result
}

func (c *Collection[T]) Max(keyFunc func(T) float64) float64 {
	if len(c.items) == 0 {
		return 0
	}

	max := keyFunc(c.items[0])
	for _, item := range c.items[1:] {
		if val := keyFunc(item); val > max {
			max = val
		}
	}
	return max
}

func (c *Collection[T]) Median(keyFunc func(T) float64) float64 {
	if len(c.items) == 0 {
		return 0
	}

	values := make([]float64, len(c.items))
	for i, item := range c.items {
		values[i] = keyFunc(item)
	}

	sort.Float64s(values)
	n := len(values)

	if n%2 == 0 {
		return (values[n/2-1] + values[n/2]) / 2
	}
	return values[n/2]
}

func (c *Collection[T]) Merge(other *Collection[T]) *Collection[T] {
	return c.Concat(other)
}

func (c *Collection[T]) MergeRecursive(other *Collection[T]) *Collection[T] {
	return c.Concat(other)
}

func (c *Collection[T]) Min(keyFunc func(T) float64) float64 {
	if len(c.items) == 0 {
		return 0
	}

	min := keyFunc(c.items[0])
	for _, item := range c.items[1:] {
		if val := keyFunc(item); val < min {
			min = val
		}
	}
	return min
}

func (c *Collection[T]) Mode(keyFunc func(T) string) []string {
	counts := c.CountBy(keyFunc)
	maxCount := 0
	for _, count := range counts {
		if count > maxCount {
			maxCount = count
		}
	}

	var modes []string
	for key, count := range counts {
		if count == maxCount {
			modes = append(modes, key)
		}
	}
	return modes
}

func New[T any](items ...T) *Collection[T] {
	return &Collection[T]{items: items}
}

func (c *Collection[T]) Nth(n int) *Collection[T] {
	if n <= 0 {
		return &Collection[T]{items: []T{}}
	}

	var result []T
	for i := 0; i < len(c.items); i += n {
		result = append(result, c.items[i])
	}
	return &Collection[T]{items: result}
}

func (c *Collection[T]) Only(indices ...int) *Collection[T] {
	var result []T
	for _, index := range indices {
		if index >= 0 && index < len(c.items) {
			result = append(result, c.items[index])
		}
	}
	return &Collection[T]{items: result}
}

func (c *Collection[T]) Pad(size int, value T) *Collection[T] {
	if size <= len(c.items) {
		return c.Clone()
	}

	padded := make([]T, size)
	copy(padded, c.items)

	for i := len(c.items); i < size; i++ {
		padded[i] = value
	}

	return &Collection[T]{items: padded}
}

func (c *Collection[T]) Partition(predicate func(T) bool) (*Collection[T], *Collection[T]) {
	var truthy, falsy []T

	for _, item := range c.items {
		if predicate(item) {
			truthy = append(truthy, item)
		} else {
			falsy = append(falsy, item)
		}
	}

	return &Collection[T]{items: truthy}, &Collection[T]{items: falsy}
}

func (c *Collection[T]) Pipe(fn func(*Collection[T]) interface{}) interface{} {
	return fn(c)
}

func (c *Collection[T]) Pluck(field string) *Collection[interface{}] {
	var result []interface{}

	for _, item := range c.items {
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			fieldValue := v.FieldByName(field)
			if fieldValue.IsValid() {
				result = append(result, fieldValue.Interface())
			}
		}
	}

	return &Collection[interface{}]{items: result}
}

func (c *Collection[T]) Pop() *T {
	if len(c.items) == 0 {
		return nil
	}
	last := c.items[len(c.items)-1]
	c.items = c.items[:len(c.items)-1]
	return &last
}

func (c *Collection[T]) Prepend(items ...T) *Collection[T] {
	c.items = append(items, c.items...)
	return c
}

func (c *Collection[T]) Pull(index int) *T {
	if index < 0 || index >= len(c.items) {
		return nil
	}

	item := c.items[index]
	c.items = append(c.items[:index], c.items[index+1:]...)
	return &item
}

func (c *Collection[T]) Push(items ...T) *Collection[T] {
	c.items = append(c.items, items...)
	return c
}

func (c *Collection[T]) Put(index int, value T) *Collection[T] {
	if index < 0 || index >= len(c.items) {
		return c
	}
	c.items[index] = value
	return c
}

func (c *Collection[T]) Random() *T {
	if len(c.items) == 0 {
		return nil
	}
	rand.Seed(time.Now().UnixNano())
	return &c.items[rand.Intn(len(c.items))]
}

func (c *Collection[T]) Reject(predicate func(T, int) bool) *Collection[T] {
	return c.Filter(func(item T, index int) bool {
		return !predicate(item, index)
	})
}

func (c *Collection[T]) Replace(replacements map[int]T) *Collection[T] {
	result := c.Clone()
	for index, value := range replacements {
		if index >= 0 && index < len(result.items) {
			result.items[index] = value
		}
	}
	return result
}

func (c *Collection[T]) ReplaceRecursive(replacements map[int]T) *Collection[T] {
	return c.Replace(replacements)
}

func (c *Collection[T]) Reverse() *Collection[T] {
	reversed := make([]T, len(c.items))
	for i, item := range c.items {
		reversed[len(c.items)-1-i] = item
	}
	return &Collection[T]{items: reversed}
}

func (c *Collection[T]) Search(value T) int {
	for i, item := range c.items {
		if reflect.DeepEqual(item, value) {
			return i
		}
	}
	return -1
}

func (c *Collection[T]) SearchBy(predicate func(T) bool) int {
	for i, item := range c.items {
		if predicate(item) {
			return i
		}
	}
	return -1
}

func (c *Collection[T]) Shift() *T {
	if len(c.items) == 0 {
		return nil
	}
	first := c.items[0]
	c.items = c.items[1:]
	return &first
}

func (c *Collection[T]) Shuffle() *Collection[T] {
	shuffled := make([]T, len(c.items))
	copy(shuffled, c.items)

	rand.Seed(time.Now().UnixNano())
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return &Collection[T]{items: shuffled}
}

func (c *Collection[T]) Skip(n int) *Collection[T] {
	return c.Drop(n)
}

func (c *Collection[T]) SkipUntil(predicate func(T) bool) *Collection[T] {
	return c.DropUntil(predicate)
}

func (c *Collection[T]) SkipWhile(predicate func(T) bool) *Collection[T] {
	return c.DropWhile(predicate)
}

func (c *Collection[T]) Slice(start, length int) *Collection[T] {
	if start < 0 {
		start = len(c.items) + start
	}
	if start < 0 {
		start = 0
	}
	if start >= len(c.items) {
		return &Collection[T]{items: []T{}}
	}

	end := start + length
	if end > len(c.items) {
		end = len(c.items)
	}

	return &Collection[T]{items: c.items[start:end]}
}

func (c *Collection[T]) Some(predicate func(T) bool) bool {
	for _, item := range c.items {
		if predicate(item) {
			return true
		}
	}
	return false
}

func (c *Collection[T]) Sort(less func(T, T) bool) *Collection[T] {
	sorted := make([]T, len(c.items))
	copy(sorted, c.items)

	sort.Slice(sorted, func(i, j int) bool {
		return less(sorted[i], sorted[j])
	})

	return &Collection[T]{items: sorted}
}

func (c *Collection[T]) SortBy(keyFunc func(T) string) *Collection[T] {
	return c.Sort(func(a, b T) bool {
		return keyFunc(a) < keyFunc(b)
	})
}

func (c *Collection[T]) SortByDesc(keyFunc func(T) string) *Collection[T] {
	return c.Sort(func(a, b T) bool {
		return keyFunc(a) > keyFunc(b)
	})
}

func (c *Collection[T]) SortDesc(less func(T, T) bool) *Collection[T] {
	return c.Sort(func(a, b T) bool {
		return less(b, a)
	})
}

func (c *Collection[T]) SortKeys() *Collection[T] {
	return c.Clone()
}

func (c *Collection[T]) SortKeysDesc() *Collection[T] {
	return c.Clone()
}

func (c *Collection[T]) Splice(start, deleteCount int, replacement ...T) *Collection[T] {
	if start < 0 {
		start = len(c.items) + start
	}
	if start < 0 {
		start = 0
	}
	if start > len(c.items) {
		start = len(c.items)
	}

	end := start + deleteCount
	if end > len(c.items) {
		end = len(c.items)
	}

	result := make([]T, start)
	copy(result, c.items[:start])
	result = append(result, replacement...)
	result = append(result, c.items[end:]...)

	return &Collection[T]{items: result}
}

func (c *Collection[T]) Split(groups int) [][]T {
	if groups <= 0 {
		return [][]T{}
	}

	size := int(math.Ceil(float64(len(c.items)) / float64(groups)))
	return c.Chunk(size)
}

func (c *Collection[T]) Sum(keyFunc func(T) float64) float64 {
	total := 0.0
	for _, item := range c.items {
		total += keyFunc(item)
	}
	return total
}

func (c *Collection[T]) Take(n int) *Collection[T] {
	if n < 0 {
		return c.Slice(n, -n)
	}
	return c.Slice(0, n)
}

func (c *Collection[T]) TakeUntil(predicate func(T) bool) *Collection[T] {
	for i, item := range c.items {
		if predicate(item) {
			return &Collection[T]{items: c.items[:i]}
		}
	}
	return c.Clone()
}

func (c *Collection[T]) TakeWhile(predicate func(T) bool) *Collection[T] {
	for i, item := range c.items {
		if !predicate(item) {
			return &Collection[T]{items: c.items[:i]}
		}
	}
	return c.Clone()
}

func (c *Collection[T]) Tap(fn func(*Collection[T])) *Collection[T] {
	fn(c)
	return c
}

func (c *Collection[T]) Times(n int, fn func(int) T) *Collection[T] {
	items := make([]T, n)
	for i := 0; i < n; i++ {
		items[i] = fn(i)
	}
	return &Collection[T]{items: items}
}

func (c *Collection[T]) ToArray() []T {
	return c.items
}

func (c *Collection[T]) ToJSON() (string, error) {
	data, err := json.Marshal(c.items)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Collection[T]) Transform(fn func(T, int) T) *Collection[T] {
	for i, item := range c.items {
		c.items[i] = fn(item, i)
	}
	return c
}

func (c *Collection[T]) Union(other *Collection[T]) *Collection[T] {
	existing := make(map[string]bool)
	for _, item := range c.items {
		existing[fmt.Sprintf("%v", item)] = true
	}

	var result []T
	result = append(result, c.items...)

	for _, item := range other.items {
		key := fmt.Sprintf("%v", item)
		if !existing[key] {
			result = append(result, item)
			existing[key] = true
		}
	}

	return &Collection[T]{items: result}
}

func (c *Collection[T]) Unique() *Collection[T] {
	seen := make(map[string]bool)
	var unique []T

	for _, item := range c.items {
		key := fmt.Sprintf("%v", item)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, item)
		}
	}
	return &Collection[T]{items: unique}
}

func (c *Collection[T]) UniqueBy(keyFunc func(T) string) *Collection[T] {
	seen := make(map[string]bool)
	var unique []T

	for _, item := range c.items {
		key := keyFunc(item)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, item)
		}
	}
	return &Collection[T]{items: unique}
}

func (c *Collection[T]) Unless(condition bool, fn func(*Collection[T]) *Collection[T]) *Collection[T] {
	return c.When(!condition, fn)
}

func (c *Collection[T]) UnlessEmpty(fn func(*Collection[T]) *Collection[T]) *Collection[T] {
	return c.Unless(c.IsEmpty(), fn)
}

func (c *Collection[T]) UnlessNotEmpty(fn func(*Collection[T]) *Collection[T]) *Collection[T] {
	return c.Unless(c.IsNotEmpty(), fn)
}

func (c *Collection[T]) Unshift(items ...T) *Collection[T] {
	c.items = append(items, c.items...)
	return c
}

func (c *Collection[T]) Values() *Collection[T] {
	return c.Clone()
}

func (c *Collection[T]) When(condition bool, fn func(*Collection[T]) *Collection[T]) *Collection[T] {
	if condition {
		return fn(c)
	}
	return c
}

func (c *Collection[T]) WhenEmpty(fn func(*Collection[T]) *Collection[T]) *Collection[T] {
	return c.When(c.IsEmpty(), fn)
}

func (c *Collection[T]) WhenNotEmpty(fn func(*Collection[T]) *Collection[T]) *Collection[T] {
	return c.When(c.IsNotEmpty(), fn)
}

func (c *Collection[T]) Where(params ...interface{}) *Collection[T] {
	switch len(params) {
	case 1:
		// where(callback)
		if callback, ok := params[0].(func(T) bool); ok {
			return c.Filter(func(item T, _ int) bool {
				return callback(item)
			})
		}
		return c
	case 2:
		// where(field, value) - assumes '=' operator
		if field, ok := params[0].(string); ok {
			return c.Filter(func(item T, _ int) bool {
				return compareFieldValue(item, field, "=", params[1])
			})
		}
		return c
	case 3:
		// where(field, operator, value)
		if field, ok := params[0].(string); ok {
			if operator, ok := params[1].(string); ok {
				return c.Filter(func(item T, _ int) bool {
					return compareFieldValue(item, field, operator, params[2])
				})
			}
		}
		return c
	default:
		return c
	}
}

func (c *Collection[T]) WhereIn(field string, values []interface{}) *Collection[T] {
	valueMap := make(map[string]bool)
	for _, v := range values {
		valueMap[fmt.Sprintf("%v", v)] = true
	}

	return c.Filter(func(item T, _ int) bool {
		fieldValue := getFieldValue(item, field)
		if fieldValue == nil {
			return false
		}
		return valueMap[fmt.Sprintf("%v", *fieldValue)]
	})
}

func (c *Collection[T]) WhereNotIn(field string, values []interface{}) *Collection[T] {
	valueMap := make(map[string]bool)
	for _, v := range values {
		valueMap[fmt.Sprintf("%v", v)] = true
	}

	return c.Filter(func(item T, _ int) bool {
		fieldValue := getFieldValue(item, field)
		if fieldValue == nil {
			return true
		}
		return !valueMap[fmt.Sprintf("%v", *fieldValue)]
	})
}

func (c *Collection[T]) WhereNotNull(field string) *Collection[T] {
	return c.Filter(func(item T, _ int) bool {
		fieldValue := getFieldValue(item, field)
		return fieldValue != nil
	})
}

func (c *Collection[T]) WhereNull(field string) *Collection[T] {
	return c.Filter(func(item T, _ int) bool {
		fieldValue := getFieldValue(item, field)
		return fieldValue == nil
	})
}

func (c *Collection[T]) Wrap(wrapper interface{}) interface{} {
	return wrapper
}

func (c *Collection[T]) Zip(other *Collection[T]) [][]T {
	maxLen := len(c.items)
	if len(other.items) > maxLen {
		maxLen = len(other.items)
	}

	var result [][]T
	for i := 0; i < maxLen; i++ {
		var pair []T
		if i < len(c.items) {
			pair = append(pair, c.items[i])
		}
		if i < len(other.items) {
			pair = append(pair, other.items[i])
		}
		result = append(result, pair)
	}

	return result
}

func compareFieldValue(item interface{}, field string, operator string, value interface{}) bool {
	fieldValue := getFieldValue(item, field)

	// Handle null comparisons
	if value == nil {
		switch operator {
		case "=", "==":
			return fieldValue == nil
		case "!=":
			return fieldValue != nil
		default:
			return false
		}
	}

	// If field is null but value isn't
	if fieldValue == nil {
		switch operator {
		case "=", "==":
			return false
		case "!=":
			return true
		default:
			return false
		}
	}

	switch operator {
	case "=", "==":
		return reflect.DeepEqual(*fieldValue, value)
	case "!=":
		return !reflect.DeepEqual(*fieldValue, value)
	case ">":
		return compareValues(*fieldValue, value) > 0
	case ">=":
		return compareValues(*fieldValue, value) >= 0
	case "<":
		return compareValues(*fieldValue, value) < 0
	case "<=":
		return compareValues(*fieldValue, value) <= 0
	case "like":
		return strings.Contains(strings.ToLower(fmt.Sprintf("%v", *fieldValue)), strings.ToLower(fmt.Sprintf("%v", value)))
	case "not like":
		return !strings.Contains(strings.ToLower(fmt.Sprintf("%v", *fieldValue)), strings.ToLower(fmt.Sprintf("%v", value)))
	default:
		return false
	}
}

func compareValues(a, b interface{}) int {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	if aFloat, err := strconv.ParseFloat(aStr, 64); err == nil {
		if bFloat, err := strconv.ParseFloat(bStr, 64); err == nil {
			if aFloat < bFloat {
				return -1
			} else if aFloat > bFloat {
				return 1
			}
			return 0
		}
	}

	if aStr < bStr {
		return -1
	} else if aStr > bStr {
		return 1
	}
	return 0
}

func getFieldValue(item interface{}, field string) *interface{} {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	fieldValue := v.FieldByName(field)
	if !fieldValue.IsValid() {
		return nil
	}

	// Handle nil pointer fields
	if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
		return nil
	}

	value := fieldValue.Interface()
	return &value
}
