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

func TestShuffle(t *testing.T) {
	randomOrder := Shuffle([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	assert.NotEqual(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, randomOrder)
}

func TestGroupBy(t *testing.T) {
	groups := GroupBy([]int{0, 1, 2, 3, 4, 5}, func(i int) int {
		return i % 3
	})
	assert.Equal(t, map[int][]int{0: []int{0, 3}, 1: []int{1, 4}, 2: []int{2, 5}}, groups)
}

func TestCount(t *testing.T) {
	count := Count([]int{1, 5, 1})
	assert.Equal(t, 3, count)
}

func TestCountBy(t *testing.T) {
	count := CountBy([]int{1, 5, 1}, func(i int) bool {
		return i < 4
	})
	assert.Equal(t, 2, count)
}

func TestEach(t *testing.T) {
	Each([]string{"hello", "world"}, func(x string, i int) {
		if i == 0 {
			assert.Equal(t, "hello", x)
		} else {
			assert.Equal(t, "world", x)
		}
	})
	Each([]int{0, 1, 2, 3}, func(x int, i int) {
		assert.Equal(t, i, x)
	})
}

func TestMin(t *testing.T) {
	min1 := Min([]int{1, 2, 3})
	min2 := Min([]int{})
	assert.Equal(t, 1, min1)
	assert.Equal(t, 0, min2)
}
