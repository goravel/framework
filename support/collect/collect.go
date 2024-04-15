package collect

import (
	"github.com/samber/lo"
	"golang.org/x/exp/constraints"
)

// Map manipulates a slice and transforms it to a slice of another type.
func Map[T any, R any](collection []T, iteratee func(item T, index int) R) []R {
	return lo.Map(collection, iteratee)
}

// Unique returns a duplicate-free version of an array, in which only the first occurrence of each element is kept.
func Unique[T comparable](collection []T) []T {
	return lo.Uniq(collection)
}

// Filter iterates over elements of collection, returning an array of all elements predicate returns truthy for.
func Filter[V any](collection []V, predicate func(item V, index int) bool) []V {
	return lo.Filter(collection, predicate)
}

// Sum sums the values in a collection. If collection is empty 0 is returned.
func Sum[T constraints.Float | constraints.Integer | constraints.Complex](collection []T) T {
	return lo.Sum(collection)
}
