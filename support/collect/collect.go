package collect

import (
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
)

func Map[T any, R any](collection []T, iteratee func(item T, index int) R) []R {
	return lop.Map(collection, iteratee)
}

func Unique[T comparable](collection []T) []T {
	return lo.Uniq(collection)
}
