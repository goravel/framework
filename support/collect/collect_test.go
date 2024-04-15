package collect

import (
	"testing"

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
	uniqValues := Unique([]int{1, 2, 2, 1})
	assert.Equal(t, []int{1, 2}, uniqValues)
}

func TestFilter(t *testing.T) {
	even := Filter([]int{1, 2, 3, 4}, func(x int, index int) bool {
		return x%2 == 0
	})
	assert.Equal(t, []int{2, 4}, even)
}

func TestSum(t *testing.T) {
	list := []int{1, 2, 3, 4, 5}
	sum := Sum(list)
	assert.Equal(t, 15, sum)
}

func TestMax(t *testing.T) {
	max1 := Max([]int{1, 2, 3})
	max2 := Max([]int{})
	assert.Equal(t, 3, max1)
	assert.Equal(t, 0, max2)
}

func TestSplit(t *testing.T) {
	result := Split([]int{0, 1, 2, 3, 4, 5}, 2)
	result1 := Split([]int{0, 1, 2, 3, 4, 5, 6}, 2)
	result2 := Split([]int{}, 2)
	result3 := Split([]int{0}, 2)

	assert.Equal(t, [][]int{{0, 1}, {2, 3}, {4, 5}}, result)
	assert.Equal(t, [][]int{{0, 1}, {2, 3}, {4, 5}, {6}}, result1)
	assert.Equal(t, [][]int{}, result2)
	assert.Equal(t, [][]int{{0}}, result3)
}

func TestReverse(t *testing.T) {
	reverseOrder := Reverse([]int{0, 1, 2, 3, 4, 5})
	assert.Equal(t, []int{5, 4, 3, 2, 1, 0}, reverseOrder)
}
