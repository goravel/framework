package collection

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"strconv"
)

func TestMap(t *testing.T) {
	results1 := Map([]int64{1, 2, 3, 4}, func(x int64, _ int) string {
		return strconv.FormatInt(x, 10)
	})
	results2 := Map([]int64{1, 2, 3, 4}, func(x int64, _ int) int64 {
		return x + 1
	})
	assert.Equal(t, []string{"1", "2", "3", "4"}, results1)
	assert.Equal(t, []int64{2, 3, 4, 5}, results2)
}

func TestUnique(t *testing.T) {
	uniqValues := lo.Uniq([]int{1, 2, 2, 1})
	assert.Equal(t, []int{1, 2}, uniqValues)
}
