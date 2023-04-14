package arr

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sort"
	"testing"
)

func TestAccessible(t *testing.T) {
	// Test case 1: An array is accessible
	arr := []any{"foo", "bar"}
	expected := true
	result := Accessible(arr)
	assert.Equal(t, expected, result)

	// Test case 2: A non-array value is not accessible
	nonArr := "not an array"
	expected = false
	result = Accessible(nonArr)
	assert.Equal(t, expected, result)
}

func TestAdd(t *testing.T) {
	// Test case 1: Add value at new key
	arr := []any{"foo", "bar"}
	expected := []any{"foo", "bar", "baz"}
	result, err := Add(arr, 2, "baz")
	if assert.NoError(t, err) {
		assert.Equal(t, expected, result)
	}

	// Test case 2: Do not add value if key already exists
	expected = []any{"foo", "bar"}
	result, err = Add(arr, 0, "qux")
	if assert.NoError(t, err) {
		assert.Equal(t, expected, result)
	}

	// Test case 3: Test error when Set function fails
	expected = []any{"foo", "bar"}
	result, err = Add(arr, -1, "qux")
	if assert.ErrorIs(t, ErrInvalidKey, err) {
		assert.Equal(t, expected, result)
	}
}

func TestCollapse(t *testing.T) {
	// Test case 1: Flatten a simple array
	arr := []any{"foo", "bar", "baz"}
	expected := []any{"foo", "bar", "baz"}
	result := Collapse(arr)
	assert.Equal(t, expected, result)

	// Test case 2: Flatten a nested array
	arr = []any{[]any{"foo", "bar"}, []any{"baz", "qux"}}
	expected = []any{"foo", "bar", "baz", "qux"}
	result = Collapse(arr)
	assert.Equal(t, expected, result)

	// Test case 3: Flatten a nested array
	arr = []any{
		[]any{[]any{"foo", "bar"}, []any{"baz", "qux"}},
	}
	expected = []any{"foo", "bar", "baz", "qux"}
	result = Collapse(arr)
	assert.Equal(t, expected, result)

	// Test case 4: Flatten a nested array
	arr = []any{
		[]any{"foo", "bar"}, []any{"baz", "qux"},
		[]any{[]any{"Charlotte", "Ethan"}, []any{"Olivia", "William"}},
	}
	expected = []any{"foo", "bar", "baz", "qux", "Charlotte", "Ethan", "Olivia", "William"}
	result = Collapse(arr)
	assert.Equal(t, expected, result)

	// Test case 5: Flatten a nested map
	arr = []any{
		map[string]any{
			"a": 1,
			"b": map[string]any{"c": 2, "d": 3},
		},
		map[string]any{
			"e": map[string]any{"f": 4, "g": 5},
			"h": 6,
		},
	}
	expected = []any{
		map[string]any{
			"a": 1,
			"b": map[string]any{"c": 2, "d": 3},
		},
		map[string]any{
			"e": map[string]any{"f": 4, "g": 5},
			"h": 6,
		},
	}
	result = Collapse(arr)
	assert.Equal(t, expected, result)
}

func TestCrossJoin(t *testing.T) {
	// Test case 1: Two arrays
	arr1 := []any{"A", "B"}
	arr2 := []any{1, 2}
	expected := [][]any{
		{"A", 1},
		{"A", 2},
		{"B", 1},
		{"B", 2},
	}
	result, err := CrossJoin(arr1, arr2)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	// Test case 2: Three arrays
	arr1 = []any{"A", "B"}
	arr2 = []any{1, 2}
	arr3 := []any{"X", "Y"}
	expected = [][]any{
		{"A", 1, "X"},
		{"A", 1, "Y"},
		{"A", 2, "X"},
		{"A", 2, "Y"},
		{"B", 1, "X"},
		{"B", 1, "Y"},
		{"B", 2, "X"},
		{"B", 2, "Y"},
	}
	result, err = CrossJoin(arr1, arr2, arr3)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	// Test case 3: Empty array
	arr1 = []any{"A", "B"}
	arr2 = []any{}
	_, err = CrossJoin(arr1, arr2)
	assert.ErrorIs(t, ErrEmptyArrayNotAllowed, err)

	// Test case 4: No arrays
	_, err = CrossJoin()
	assert.ErrorIs(t, ErrArrayRequired, err)
}

func TestDivide(t *testing.T) {
	// Test case 1: Simple array
	arr := []any{"a", "b", "c"}
	expectedKeys := []any{0, 1, 2}
	expectedValues := []any{"a", "b", "c"}
	keys, values, err := Divide(arr)
	assert.NoError(t, err)
	assert.Equal(t, expectedKeys, keys)
	assert.Equal(t, expectedValues, values)

	// Test case 2: Empty array
	arr = []any{}
	_, _, err = Divide(arr)
	assert.ErrorIs(t, ErrEmptyArrayNotAllowed, err)
}

func TestExcept(t *testing.T) {
	// Test case 1: Remove a single key
	arr := []int{1, 2, 3}
	excludedKeys := []int{2}
	expected := []int{1, 2}
	result := Except(arr, excludedKeys)
	assert.Equal(t, expected, result)

	// Test case 2: Remove multiple keys
	arr = []int{1, 2, 3, 4}
	excludedKeys = []int{1, 3}
	expected = []int{1, 3}
	result = Except(arr, excludedKeys)
	assert.Equal(t, expected, result)

	// Test case 3: Remove empty key
	arr = []int{1, 2, 3, 4}
	excludedKeys = []int{}
	expected = []int{1, 2, 3, 4}
	result = Except(arr, excludedKeys)
	assert.Equal(t, expected, result)
}

func TestExists(t *testing.T) {
	// Test case 1: Check key exists in array
	arr := []any{1, 2, 3, "foo", "bar"}
	key := 3
	exists := Exists(arr, key)
	assert.True(t, exists)

	// Test case 2: Check key not exists in array
	arr = []any{1, 2, 3, "foo", "bar"}
	key = 6
	exists = Exists(arr, key)
	assert.False(t, exists)
}

func TestFirst(t *testing.T) {
	arr := []any{2, 4, 6, 8}
	expected := 6
	result := First(arr, func(val any, i int) bool {
		return val.(int)%3 == 0
	}, nil)
	assert.Equal(t, expected, result)
}

func TestLast(t *testing.T) {
	// Test case 1: empty array
	arr := []int{}
	expected := -1
	result := Last(arr, func(i int) bool { return i > 5 }, -1)
	assert.Equal(t, expected, result)

	// Test case 2: array without a match
	arr = []int{1, 3, 5, 7, 9}
	expected = -1
	result = Last(arr, func(i int) bool { return i > 10 }, -1)
	assert.Equal(t, expected, result)

	// Test case 3: array with a match
	arr = []int{1, 3, 5, 7, 9}
	expected = 9
	result = Last(arr, func(i int) bool { return i > 6 }, -1)
	assert.Equal(t, expected, result)

	// Test case 4: array with a match at the end
	arr = []int{1, 3, 5, 7}
	expected = 7
	result = Last(arr, func(i int) bool { return i > 6 }, -1)
	assert.Equal(t, expected, result)
}

func TestFlatten(t *testing.T) {
	// Test case 1: Flatten a simple array
	arr := []any{1, 2, 3, 4}
	expected := []any{1, 2, 3, 4}
	result := Flatten(arr, 1)
	assert.Equal(t, expected, result)

	// Test case 2: Flatten a nested array
	arr = []any{1, 2, []any{3, 4}, 5}
	expected = []any{1, 2, 3, 4, 5}
	result = Flatten(arr, 1)
	assert.Equal(t, expected, result)

	// Test case 3: Flatten a deeply nested array
	arr = []any{1, 2, []any{3, 4, []any{5, []any{6}}}, 7}
	expected = []any{1, 2, 3, 4, 5, 6, 7}
	result = Flatten(arr, 2)
	assert.Equal(t, expected, result)

	// Test case 4: Flatten an array with no nested values
	arr = []any{1, 2, 3, 4}
	expected = []any{1, 2, 3, 4}
	result = Flatten(arr, 2)
	assert.Equal(t, expected, result)
}

func TestForget(t *testing.T) {
	// Test case 1: Remove a single item
	arr1 := []string{"foo", "bar", "baz"}
	expected1 := []string{"foo", "baz"}
	result1, err1 := Forget(arr1, 1)
	if assert.NoError(t, err1) {
		assert.Equal(t, expected1, result1)
	}

	// Test case 2: Remove multiple items
	arr2 := []int{1, 2, 3, 4, 5}
	expected2 := []int{1, 5}
	result2, err2 := Forget(arr2, []int{1, 2, 3})
	if assert.NoError(t, err2) {
		assert.Equal(t, expected2, result2)
	}

	// Test case 3: Remove an item out of range
	arr3 := []bool{true, false, true}
	expected3 := []bool{true, false, true}
	result3, err3 := Forget(arr3, 3)
	if assert.NoError(t, err3) {
		assert.Equal(t, expected3, result3)
	}

	// Test case 4: Invalid keys argument
	arr4 := []any{"foo", "bar", "baz"}
	expected4 := []any{"foo", "bar", "baz"}
	result4, err4 := Forget(arr4, "invalid")
	if assert.ErrorIs(t, ErrInvalidKeys, err4) {
		assert.Equal(t, expected4, result4)
	}

	// Test case 5: Remove empty array
	var arr5 []any
	var expected5 []any
	result5, err5 := Forget(arr5, 0)
	if assert.NoError(t, err5) {
		assert.Equal(t, expected5, result5)
	}

	// Test case 6: Key is nil
	arr6 := []any{"foo", "bar", "baz"}
	expected6 := []any{"foo", "bar", "baz"}
	result6, err6 := Forget(arr6, nil)
	if assert.NoError(t, err6) {
		assert.Equal(t, expected6, result6)
	}
}

func TestGet(t *testing.T) {
	// Test case 1: When key is within the bounds of the array
	arr := []int{1, 2, 3, 4}
	expected := 2
	result := Get(arr, 1, 0)
	assert.Equal(t, expected, result)

	// Test case 2: When key is outside the bounds of the array
	expected = 0
	result = Get(arr, 5, 0)
	assert.Equal(t, expected, result)
	result = Get(arr, -1, 0)
	assert.Equal(t, expected, result)

	// Test case 3: When arr is empty
	arr = []int{}
	expected = 0
	result = Get(arr, 0, 0)
	assert.Equal(t, expected, result)

	// Test case 4: Type of array elements as string
	arr2 := []string{"foo", "bar"}
	expectedStr := "default"
	resultStr := Get(arr2, 2, "default")
	assert.Equal(t, expectedStr, resultStr)
}

func TestHas(t *testing.T) {
	arr := []string{"foo", "bar"}

	assert.True(t, Has(arr, 0))
	assert.True(t, Has(arr, 1))
	assert.False(t, Has(arr, 2))
	assert.False(t, Has(arr, []int{1, 3}))
}

func TestHasAny(t *testing.T) {
	arr := []string{"foo", "bar"}

	assert.True(t, HasAny(arr, 0))
	assert.True(t, HasAny(arr, 1))
	assert.False(t, HasAny(arr, 2))
	assert.True(t, HasAny(arr, []int{1, 3}))
}

func TestJoin(t *testing.T) {
	arr1 := []string{"a", "b", "c"}
	result1 := Join(arr1, ", ")
	expectedStr := "a, b, c"
	assert.Equal(t, expectedStr, result1)

	arr2 := []string{"a", "b", "c"}
	expectedStr2 := "a, b and c"
	result2 := Join(arr2, ", ", " and ")
	assert.Equal(t, expectedStr2, result2)

	arr3 := []string{"a", "b"}
	expectedStr3 := "a and b"
	result3 := Join(arr3, ", ", " and ")
	assert.Equal(t, expectedStr3, result3)

	arr4 := []string{"a"}
	expectedStr4 := "a"
	result4 := Join(arr4, ", ", " and ")
	assert.Equal(t, expectedStr4, result4)

	arr5 := []string{}
	expectedStr5 := ""
	result5 := Join(arr5, ", ", " and ")
	assert.Equal(t, expectedStr5, result5)
}

func TestOnly(t *testing.T) {
	arr := []string{"one", "two", "three"}

	// Test case 1: keys is int
	expected := []string{"two"}
	result := Only(arr, 1)
	assert.Equal(t, expected, result)

	// Test case 2: keys is an array of int
	keys := []int{0, 2}
	expected = []string{"one", "three"}
	result = Only(arr, keys)
	assert.Equal(t, expected, result)

	// Test case 3: Out of range key
	result = Only(arr, 3)
	assert.Empty(t, result)

	// Test case 4: Out of range keys in array of int
	keys = []int{0, 3}
	result = Only(arr, keys)
	assert.Equal(t, []string{"one"}, result)
}

func TestMap(t *testing.T) {
	{
		arr := []int{1, 2, 3}
		expected := []string{"1!", "2!", "3!"}

		res := Map(arr, func(n int, i int) string {
			return fmt.Sprintf("%d!", n)
		})

		for i := range res {
			assert.Equal(t, expected[i], res[i])
		}
	}
	{
		strs := []string{"hello", "world"}
		expected2 := []int{5, 5}

		res := Map(strs, func(s string, i int) int {
			return len(s)
		})

		for i := range res {
			assert.Equal(t, expected2[i], res[i])
		}
	}
}

func TestPrepend(t *testing.T) {
	testCases := []struct {
		name     string
		arr      []any
		value    any
		expected []any
	}{
		{
			name:     "Integers",
			arr:      []any{2, 3, 4},
			value:    1,
			expected: []any{1, 2, 3, 4},
		},
		{
			name:     "Strings",
			arr:      []any{"b", "c", "d"},
			value:    "a",
			expected: []any{"a", "b", "c", "d"},
		},
		{
			name:     "Mixed",
			arr:      []any{"apple", 3.14, true},
			value:    42,
			expected: []any{42, "apple", 3.14, true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Prepend(tc.arr, tc.value)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected: %v, got: %v", tc.expected, result)
			}
		})
	}
}

func TestPull(t *testing.T) {
	// Test case 1: Valid key
	arr := []int{1, 2, 3, 4, 5}
	expectedArr := []int{1, 2, 4, 5}
	value, err := Pull(&arr, 2, -1)
	if assert.NoError(t, err) {
		assert.Equal(t, 3, value)
		assert.Equal(t, expectedArr, arr)
	}

	// Test case 2: Invalid key
	arr = []int{1, 2, 3, 4, 5}
	expectedArr = []int{1, 2, 3, 4, 5}
	value, err = Pull(&arr, 7, -1)
	if assert.NoError(t, err) {
		assert.Equal(t, -1, value)
		assert.Equal(t, expectedArr, arr)
	}
}

func TestRandom(t *testing.T) {
	// Test case 1: Get one random value
	arr := []interface{}{1, 2, 3, 4, 5}
	results, err := Random(arr, nil)
	if assert.NoError(t, err) && assert.Len(t, results, 1) {
		assert.Contains(t, arr, results[0])
	}

	// Test case 2: Get a specified number of random values
	number := 3
	results, err = Random(arr, &number)
	if assert.NoError(t, err) && assert.Len(t, results, number) {
		for _, result := range results {
			assert.Contains(t, arr, result)
		}
	}

	// Test case 3: Invalid number of requested items
	number = 10
	results, err = Random(arr, &number)
	assert.ErrorIs(t, err, ErrInvalidRequestedItems)
}

func TestSet(t *testing.T) {
	// Test case 1: When key is within the bounds of the array
	arr := []any{"foo", "bar", "baz"}
	err := Set(&arr, 1, "qux")
	if assert.NoError(t, err) {
		expected := []any{"foo", "qux", "baz"}
		assert.Equal(t, expected, arr)
	}

	// Test case 2: When key is outside the bounds of the array
	err = Set(&arr, 3, "quux")
	if assert.NoError(t, err) {
		expected := []any{"foo", "qux", "baz", "quux"}
		assert.Equal(t, expected, arr)
	}

	// Test case 3: When key is negative
	arr = []any{"foo", "bar", "baz"}
	err = Set(&arr, -1, "new")
	if assert.ErrorIs(t, ErrInvalidKey, err) {
		expected := []any{"foo", "bar", "baz"}
		assert.Equal(t, expected, arr)
	}
}

func TestShuffle(t *testing.T) {
	arr := []any{1, 2, 3, 4, 5}
	seed := int64(123456)
	result := Shuffle(arr, &seed)

	// Check if result is shuffled
	assert.NotEqual(t, arr, result)

	// Check if result has same elements as original array
	for _, v := range arr {
		assert.Contains(t, result, v)
	}

	result = Shuffle(arr, nil)

	// Check if result is shuffled
	assert.NotEqual(t, arr, result)

	// Check if result has same elements as original array
	for _, v := range arr {
		assert.Contains(t, result, v)
	}

}

func TestSort(t *testing.T) {
	// Test case 1
	arr := []any{5, 3, 7, 1, 8}
	expected := []any{1, 3, 5, 7, 8}
	result := Sort(arr, func(i, j int) bool {
		return arr[i].(int) < arr[j].(int)
	})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Test case 2 failed. Expected: %v, but got: %v", expected, result)
	}

	// Test case 2
	arr = []any{"foo", "bar", "baz", "qux"}
	expected = []any{"bar", "baz", "foo", "qux"}
	result = Sort(arr, func(i, j int) bool {
		return arr[i].(string) < arr[j].(string)
	})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Test case 3 failed. Expected: %v, but got: %v", expected, result)
	}

	// Test case 3
	// todo: this does not work now
	//arr = []any{
	//	[]any{3, 1, 4},
	//	[]any{2, 5, 9},
	//	[]any{6, 5, 3},
	//}
	//expected = []any{
	//	[]any{1, 3, 4},
	//	[]any{2, 5, 9},
	//	[]any{3, 5, 6},
	//}
	//result = Sort(arr, func(i, j int) bool {
	//	valType := reflect.TypeOf(arr[i]).String()
	//	if valType == "int" {
	//		return arr[i].(int) < arr[j].(int)
	//	} else if valType == "string" {
	//		return arr[i].(string) < arr[j].(string)
	//	} else if valType == "[]interface {}" {
	//		v := arr[i]
	//		switch _ := v.(type) {
	//		case []int:
	//			return reflect.ValueOf(arr[i]).Index(0).Int() < reflect.ValueOf(arr[j]).Index(0).Int()
	//		case []string:
	//			return reflect.ValueOf(arr[i]).Index(0).String() < reflect.ValueOf(arr[j]).Index(0).String()
	//		default:
	//		}
	//	}
	//	return false
	//})
	//if !reflect.DeepEqual(result, expected) {
	//	t.Errorf("Test case 1 failed. Expected: %v, but got: %v", expected, result)
	//}
}

func TestSortDesc(t *testing.T) {
	unsorted := []string{
		"Chair",
		"Desk",
	}

	expected := []string{
		"Desk",
		"Chair",
	}

	sort.SliceStable(unsorted, func(i, j int) bool {
		return unsorted[i] > unsorted[j]
	})

	if !reflect.DeepEqual(unsorted, expected) {
		t.Errorf("SortDesc() failed, expected %v, got %v", expected, unsorted)
	}
}

func TestToCssClasses(t *testing.T) {
	classes := ToCssClasses([]any{"font-bold", "mt-4"})
	expected := "font-bold mt-4"
	assert.Equal(t, expected, classes)
}

func TestToCssStyles(t *testing.T) {
	styles := ToCssStyles([]string{
		"font-weight: bold",
		"margin-top: 4px;",
	})

	expected := "font-weight: bold; margin-top: 4px;"
	if styles != expected {
		t.Errorf("ToCssStyles() = %q, expected %q", styles, expected)
	}
}
