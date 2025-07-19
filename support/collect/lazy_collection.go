package collect

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type LazyCollection[T any] struct {
	generator func() <-chan T
	pipeline  []func(<-chan T) <-chan T
}

type LazyIterator[T any] interface {
	Next() (T, bool)
	Reset()
}

type lazyIterator[T any] struct {
	ch   <-chan T
	done bool
}

func (it *lazyIterator[T]) Next() (T, bool) {
	if it.done {
		var zero T
		return zero, false
	}

	value, ok := <-it.ch
	if !ok {
		it.done = true
		var zero T
		return zero, false
	}

	return value, true
}

func (it *lazyIterator[T]) Reset() {
	it.done = false
}

// PUBLIC METHODS (Alphabetical Order)

func (lc *LazyCollection[T]) All() []T {
	var result []T
	ch := lc.execute()
	for item := range ch {
		result = append(result, item)
	}
	return result
}

func (lc *LazyCollection[T]) Average(keyFunc func(T) float64) float64 {
	count := 0
	sum := 0.0
	ch := lc.execute()

	for item := range ch {
		sum += keyFunc(item)
		count++
	}

	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func (lc *LazyCollection[T]) Avg(keyFunc func(T) float64) float64 {
	return lc.Average(keyFunc)
}

func (lc *LazyCollection[T]) Chunk(size int) [][]T {
	if size <= 0 {
		return [][]T{}
	}

	var chunks [][]T
	var currentChunk []T
	ch := lc.execute()

	for item := range ch {
		currentChunk = append(currentChunk, item)
		if len(currentChunk) == size {
			chunks = append(chunks, currentChunk)
			currentChunk = []T{}
		}
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

func (lc *LazyCollection[T]) Collect() *Collection[T] {
	return &Collection[T]{items: lc.All()}
}

func (lc *LazyCollection[T]) Contains(value T) bool {
	ch := lc.execute()
	for item := range ch {
		if reflect.DeepEqual(item, value) {
			return true
		}
	}
	return false
}

func (lc *LazyCollection[T]) Count() int {
	count := 0
	ch := lc.execute()
	for range ch {
		count++
	}
	return count
}

func (lc *LazyCollection[T]) CountBy(keyFunc func(T) string) map[string]int {
	counts := make(map[string]int)
	ch := lc.execute()

	for item := range ch {
		key := keyFunc(item)
		counts[key]++
	}

	return counts
}

func (lc *LazyCollection[T]) Debug() *LazyCollection[T] {
	newPipeline := make([]func(<-chan T) <-chan T, len(lc.pipeline))
	copy(newPipeline, lc.pipeline)

	newPipeline = append(newPipeline, func(input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			for item := range input {
				fmt.Printf("Debug: %+v\n", item)
				output <- item
			}
		}()
		return output
	})

	return &LazyCollection[T]{
		generator: lc.generator,
		pipeline:  newPipeline,
	}
}

func (lc *LazyCollection[T]) Drop(n int) *LazyCollection[T] {
	return lc.Skip(n)
}

func (lc *LazyCollection[T]) DropWhile(predicate func(T) bool) *LazyCollection[T] {
	newPipeline := make([]func(<-chan T) <-chan T, len(lc.pipeline))
	copy(newPipeline, lc.pipeline)

	newPipeline = append(newPipeline, func(input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			dropping := true
			for item := range input {
				if dropping && predicate(item) {
					continue
				}
				dropping = false
				output <- item
			}
		}()
		return output
	})

	return &LazyCollection[T]{
		generator: lc.generator,
		pipeline:  newPipeline,
	}
}

func (lc *LazyCollection[T]) Each(fn func(T, int)) *LazyCollection[T] {
	newPipeline := make([]func(<-chan T) <-chan T, len(lc.pipeline))
	copy(newPipeline, lc.pipeline)

	newPipeline = append(newPipeline, func(input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			index := 0
			for item := range input {
				fn(item, index)
				output <- item
				index++
			}
		}()
		return output
	})

	return &LazyCollection[T]{
		generator: lc.generator,
		pipeline:  newPipeline,
	}
}

func (lc *LazyCollection[T]) Every(predicate func(T) bool) bool {
	ch := lc.execute()
	for item := range ch {
		if !predicate(item) {
			return false
		}
	}
	return true
}

func (lc *LazyCollection[T]) Filter(predicate func(T, int) bool) *LazyCollection[T] {
	newPipeline := make([]func(<-chan T) <-chan T, len(lc.pipeline))
	copy(newPipeline, lc.pipeline)

	newPipeline = append(newPipeline, func(input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			index := 0
			for item := range input {
				if predicate(item, index) {
					output <- item
				}
				index++
			}
		}()
		return output
	})

	return &LazyCollection[T]{
		generator: lc.generator,
		pipeline:  newPipeline,
	}
}

func (lc *LazyCollection[T]) First() *T {
	ch := lc.execute()
	for item := range ch {
		return &item
	}
	return nil
}

func (lc *LazyCollection[T]) FirstOrFail() (*T, error) {
	first := lc.First()
	if first == nil {
		return nil, fmt.Errorf("collection is empty")
	}
	return first, nil
}

func (lc *LazyCollection[T]) FirstWhere(predicate func(T) bool) *T {
	ch := lc.execute()
	for item := range ch {
		if predicate(item) {
			return &item
		}
	}
	return nil
}

func (lc *LazyCollection[T]) FlatMap(fn func(T) []T) *LazyCollection[T] {
	newPipeline := make([]func(<-chan T) <-chan T, len(lc.pipeline))
	copy(newPipeline, lc.pipeline)

	newPipeline = append(newPipeline, func(input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			for item := range input {
				results := fn(item)
				for _, result := range results {
					output <- result
				}
			}
		}()
		return output
	})

	return &LazyCollection[T]{
		generator: lc.generator,
		pipeline:  newPipeline,
	}
}

func (lc *LazyCollection[T]) ForEach(fn func(T)) {
	ch := lc.execute()
	for item := range ch {
		fn(item)
	}
}

func (lc *LazyCollection[T]) GroupBy(keyFunc func(T) string) map[string]*Collection[T] {
	groups := make(map[string]*Collection[T])
	ch := lc.execute()

	for item := range ch {
		key := keyFunc(item)
		if _, exists := groups[key]; !exists {
			groups[key] = &Collection[T]{items: []T{}}
		}
		groups[key].items = append(groups[key].items, item)
	}

	return groups
}

func (lc *LazyCollection[T]) IsEmpty() bool {
	ch := lc.execute()
	for range ch {
		return false
	}
	return true
}

func (lc *LazyCollection[T]) IsNotEmpty() bool {
	return !lc.IsEmpty()
}

func (lc *LazyCollection[T]) Iterator() LazyIterator[T] {
	return &lazyIterator[T]{
		ch:   lc.execute(),
		done: false,
	}
}

func (lc *LazyCollection[T]) Join(separator string) string {
	var parts []string
	ch := lc.execute()
	for item := range ch {
		parts = append(parts, fmt.Sprintf("%v", item))
	}
	return strings.Join(parts, separator)
}

func (lc *LazyCollection[T]) Last() *T {
	var last *T
	ch := lc.execute()
	for item := range ch {
		last = &item
	}
	return last
}

func (lc *LazyCollection[T]) Map(fn func(T, int) interface{}) *LazyCollection[interface{}] {
	return &LazyCollection[interface{}]{
		generator: func() <-chan interface{} {
			ch := make(chan interface{})
			go func() {
				defer close(ch)
				input := lc.execute()
				index := 0
				for item := range input {
					ch <- fn(item, index)
					index++
				}
			}()
			return ch
		},
		pipeline: []func(<-chan interface{}) <-chan interface{}{},
	}
}

func LazyCollect[T any](items []T) *LazyCollection[T] {
	return &LazyCollection[T]{
		generator: func() <-chan T {
			ch := make(chan T)
			go func() {
				defer close(ch)
				for _, item := range items {
					ch <- item
				}
			}()
			return ch
		},
		pipeline: []func(<-chan T) <-chan T{},
	}
}

func LazyFromChannel[T any](ch <-chan T) *LazyCollection[T] {
	return &LazyCollection[T]{
		generator: func() <-chan T {
			return ch
		},
		pipeline: []func(<-chan T) <-chan T{},
	}
}

func LazyFromFunc[T any](fn func() <-chan T) *LazyCollection[T] {
	return &LazyCollection[T]{
		generator: fn,
		pipeline:  []func(<-chan T) <-chan T{},
	}
}

func LazyGenerate[T any](fn func(int) T, count int) *LazyCollection[T] {
	return &LazyCollection[T]{
		generator: func() <-chan T {
			ch := make(chan T)
			go func() {
				defer close(ch)
				for i := 0; i < count; i++ {
					ch <- fn(i)
				}
			}()
			return ch
		},
		pipeline: []func(<-chan T) <-chan T{},
	}
}

func LazyMap[T, R any](lc *LazyCollection[T], fn func(T, int) R) *LazyCollection[R] {
	return &LazyCollection[R]{
		generator: func() <-chan R {
			ch := make(chan R)
			go func() {
				defer close(ch)
				input := lc.execute()
				index := 0
				for item := range input {
					ch <- fn(item, index)
					index++
				}
			}()
			return ch
		},
		pipeline: []func(<-chan R) <-chan R{},
	}
}

func LazyNew[T any](items ...T) *LazyCollection[T] {
	return LazyCollect(items)
}

func LazyRange(start, end int) *LazyCollection[int] {
	return &LazyCollection[int]{
		generator: func() <-chan int {
			ch := make(chan int)
			go func() {
				defer close(ch)
				for i := start; i < end; i++ {
					ch <- i
				}
			}()
			return ch
		},
		pipeline: []func(<-chan int) <-chan int{},
	}
}

func LazyReduce[T, R any](lc *LazyCollection[T], fn func(R, T, int) R, initial R) R {
	result := initial
	ch := lc.execute()
	index := 0
	for item := range ch {
		result = fn(result, item, index)
		index++
	}
	return result
}

func LazyRepeat[T any](value T, count int) *LazyCollection[T] {
	return &LazyCollection[T]{
		generator: func() <-chan T {
			ch := make(chan T)
			go func() {
				defer close(ch)
				for i := 0; i < count; i++ {
					ch <- value
				}
			}()
			return ch
		},
		pipeline: []func(<-chan T) <-chan T{},
	}
}

func (lc *LazyCollection[T]) Max(keyFunc func(T) float64) float64 {
	ch := lc.execute()
	var max float64
	first := true

	for item := range ch {
		val := keyFunc(item)
		if first || val > max {
			max = val
			first = false
		}
	}

	return max
}

func (lc *LazyCollection[T]) Min(keyFunc func(T) float64) float64 {
	ch := lc.execute()
	var min float64
	first := true

	for item := range ch {
		val := keyFunc(item)
		if first || val < min {
			min = val
			first = false
		}
	}

	return min
}

func (lc *LazyCollection[T]) Partition(predicate func(T) bool) (*Collection[T], *Collection[T]) {
	var truthy, falsy []T
	ch := lc.execute()

	for item := range ch {
		if predicate(item) {
			truthy = append(truthy, item)
		} else {
			falsy = append(falsy, item)
		}
	}

	return &Collection[T]{items: truthy}, &Collection[T]{items: falsy}
}

func (lc *LazyCollection[T]) Pluck(field string) *LazyCollection[interface{}] {
	return &LazyCollection[interface{}]{
		generator: func() <-chan interface{} {
			ch := make(chan interface{})
			go func() {
				defer close(ch)
				input := lc.execute()
				for item := range input {
					v := reflect.ValueOf(item)
					if v.Kind() == reflect.Ptr {
						v = v.Elem()
					}
					if v.Kind() == reflect.Struct {
						fieldValue := v.FieldByName(field)
						if fieldValue.IsValid() {
							ch <- fieldValue.Interface()
						}
					}
				}
			}()
			return ch
		},
		pipeline: []func(<-chan interface{}) <-chan interface{}{},
	}
}

func (lc *LazyCollection[T]) Reject(predicate func(T, int) bool) *LazyCollection[T] {
	return lc.Filter(func(item T, index int) bool {
		return !predicate(item, index)
	})
}

func (lc *LazyCollection[T]) Reverse() *LazyCollection[T] {
	return &LazyCollection[T]{
		generator: func() <-chan T {
			ch := make(chan T)
			go func() {
				defer close(ch)
				items := lc.All()
				for i := len(items) - 1; i >= 0; i-- {
					ch <- items[i]
				}
			}()
			return ch
		},
		pipeline: []func(<-chan T) <-chan T{},
	}
}

func (lc *LazyCollection[T]) Skip(n int) *LazyCollection[T] {
	newPipeline := make([]func(<-chan T) <-chan T, len(lc.pipeline))
	copy(newPipeline, lc.pipeline)

	newPipeline = append(newPipeline, func(input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			skipped := 0
			for item := range input {
				if skipped < n {
					skipped++
					continue
				}
				output <- item
			}
		}()
		return output
	})

	return &LazyCollection[T]{
		generator: lc.generator,
		pipeline:  newPipeline,
	}
}

func (lc *LazyCollection[T]) SkipWhile(predicate func(T) bool) *LazyCollection[T] {
	return lc.DropWhile(predicate)
}

func (lc *LazyCollection[T]) Some(predicate func(T) bool) bool {
	ch := lc.execute()
	for item := range ch {
		if predicate(item) {
			return true
		}
	}
	return false
}

func (lc *LazyCollection[T]) Sort(less func(T, T) bool) *LazyCollection[T] {
	return &LazyCollection[T]{
		generator: func() <-chan T {
			ch := make(chan T)
			go func() {
				defer close(ch)
				items := lc.All()
				sort.Slice(items, func(i, j int) bool {
					return less(items[i], items[j])
				})
				for _, item := range items {
					ch <- item
				}
			}()
			return ch
		},
		pipeline: []func(<-chan T) <-chan T{},
	}
}

func (lc *LazyCollection[T]) SortBy(keyFunc func(T) string) *LazyCollection[T] {
	return lc.Sort(func(a, b T) bool {
		return keyFunc(a) < keyFunc(b)
	})
}

func (lc *LazyCollection[T]) Sum(keyFunc func(T) float64) float64 {
	total := 0.0
	ch := lc.execute()
	for item := range ch {
		total += keyFunc(item)
	}
	return total
}

func (lc *LazyCollection[T]) Take(n int) *LazyCollection[T] {
	newPipeline := make([]func(<-chan T) <-chan T, len(lc.pipeline))
	copy(newPipeline, lc.pipeline)

	newPipeline = append(newPipeline, func(input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			taken := 0
			for item := range input {
				if taken >= n {
					break
				}
				output <- item
				taken++
			}
		}()
		return output
	})

	return &LazyCollection[T]{
		generator: lc.generator,
		pipeline:  newPipeline,
	}
}

func (lc *LazyCollection[T]) TakeWhile(predicate func(T) bool) *LazyCollection[T] {
	newPipeline := make([]func(<-chan T) <-chan T, len(lc.pipeline))
	copy(newPipeline, lc.pipeline)

	newPipeline = append(newPipeline, func(input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			for item := range input {
				if !predicate(item) {
					break
				}
				output <- item
			}
		}()
		return output
	})

	return &LazyCollection[T]{
		generator: lc.generator,
		pipeline:  newPipeline,
	}
}

func (lc *LazyCollection[T]) Tap(fn func(*LazyCollection[T])) *LazyCollection[T] {
	fn(lc)
	return lc
}

func (lc *LazyCollection[T]) ToArray() []T {
	return lc.All()
}

func (lc *LazyCollection[T]) ToJSON() (string, error) {
	data, err := json.Marshal(lc.All())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (lc *LazyCollection[T]) Unique() *LazyCollection[T] {
	newPipeline := make([]func(<-chan T) <-chan T, len(lc.pipeline))
	copy(newPipeline, lc.pipeline)

	newPipeline = append(newPipeline, func(input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			seen := make(map[string]bool)
			for item := range input {
				key := fmt.Sprintf("%v", item)
				if !seen[key] {
					seen[key] = true
					output <- item
				}
			}
		}()
		return output
	})

	return &LazyCollection[T]{
		generator: lc.generator,
		pipeline:  newPipeline,
	}
}

func (lc *LazyCollection[T]) UniqueBy(keyFunc func(T) string) *LazyCollection[T] {
	newPipeline := make([]func(<-chan T) <-chan T, len(lc.pipeline))
	copy(newPipeline, lc.pipeline)

	newPipeline = append(newPipeline, func(input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			seen := make(map[string]bool)
			for item := range input {
				key := keyFunc(item)
				if !seen[key] {
					seen[key] = true
					output <- item
				}
			}
		}()
		return output
	})

	return &LazyCollection[T]{
		generator: lc.generator,
		pipeline:  newPipeline,
	}
}

func (lc *LazyCollection[T]) When(condition bool, fn func(*LazyCollection[T]) *LazyCollection[T]) *LazyCollection[T] {
	if condition {
		return fn(lc)
	}
	return lc
}

func (lc *LazyCollection[T]) Where(params ...interface{}) *LazyCollection[T] {
	switch len(params) {
	case 1:
		// where(callback)
		if callback, ok := params[0].(func(T) bool); ok {
			return lc.Filter(func(item T, _ int) bool {
				return callback(item)
			})
		}
		return lc
	case 2:
		// where(field, value) - assumes '=' operator
		if field, ok := params[0].(string); ok {
			return lc.Filter(func(item T, _ int) bool {
				return compareFieldValue(item, field, "=", params[1])
			})
		}
		return lc
	case 3:
		// where(field, operator, value)
		if field, ok := params[0].(string); ok {
			if operator, ok := params[1].(string); ok {
				return lc.Filter(func(item T, _ int) bool {
					return compareFieldValue(item, field, operator, params[2])
				})
			}
		}
		return lc
	default:
		return lc
	}
}

func (lc *LazyCollection[T]) WhereIn(field string, values []interface{}) *LazyCollection[T] {
	valueMap := make(map[string]bool)
	for _, v := range values {
		valueMap[fmt.Sprintf("%v", v)] = true
	}

	return lc.Filter(func(item T, _ int) bool {
		fieldValue := getFieldValue(item, field)
		if fieldValue == nil {
			return false
		}
		return valueMap[fmt.Sprintf("%v", *fieldValue)]
	})
}

func (lc *LazyCollection[T]) WhereNotIn(field string, values []interface{}) *LazyCollection[T] {
	valueMap := make(map[string]bool)
	for _, v := range values {
		valueMap[fmt.Sprintf("%v", v)] = true
	}

	return lc.Filter(func(item T, _ int) bool {
		fieldValue := getFieldValue(item, field)
		if fieldValue == nil {
			return true
		}
		return !valueMap[fmt.Sprintf("%v", *fieldValue)]
	})
}

func (lc *LazyCollection[T]) Zip(other *LazyCollection[T]) [][]T {
	var result [][]T
	ch1 := lc.execute()
	ch2 := other.execute()

	for {
		select {
		case item1, ok1 := <-ch1:
			if !ok1 {
				return result
			}
			select {
			case item2, ok2 := <-ch2:
				if !ok2 {
					return result
				}
				result = append(result, []T{item1, item2})
			}
		}
	}
}

// INTERNAL HELPER FUNCTIONS (Alphabetical Order)

func (lc *LazyCollection[T]) execute() <-chan T {
	ch := lc.generator()

	for _, stage := range lc.pipeline {
		ch = stage(ch)
	}

	return ch
}
