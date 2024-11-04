package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractstesting "github.com/goravel/framework/contracts/testing"
)

func TestNewAssertableJSON(t *testing.T) {
	validJSON := `{"key1": "value1", "key2": [1, 2, 3]}`
	invalidJSON := `{"key1": "value1", "key2": [1, 2, 3]`

	assertable, err := NewAssertableJSON(t, validJSON)
	assert.NoError(t, err)
	assert.NotNil(t, assertable)

	assertable, err = NewAssertableJSON(t, invalidJSON)
	assert.Error(t, err)
	assert.Nil(t, assertable)
}

func TestCount(t *testing.T) {
	jsonStr := `{"items": [1, 2, 3], "otherKey": "value"}`
	assertable, _ := NewAssertableJSON(t, jsonStr)

	assertable.Count("items", 3)

	//assert.Panics(t, func() {
	//	assertable.Count("items", 4)
	//})
}

func TestHas(t *testing.T) {
	jsonStr := `{"key1": "value1", "key2": [1, 2, 3]}`
	assertable, _ := NewAssertableJSON(t, jsonStr)

	assertable.Has("key1")

	//assert.Panics(t, func() {
	//	assertable.Has("nonExistingKey")
	//})
}

func TestHasAll(t *testing.T) {
	jsonStr := `{"key1": "value1", "key2": [1, 2, 3]}`
	assertable, _ := NewAssertableJSON(t, jsonStr)

	// Test all keys exist
	assertable.HasAll([]string{"key1", "key2"})

	// Test one key does not exist
	//assert.Panics(t, func() {
	//	assertable.HasAll([]string{"key1", "nonExistingKey"})
	//})
}

func TestHasAny(t *testing.T) {
	jsonStr := `{"key1": "value1", "key2": [1, 2, 3]}`
	assertable, _ := NewAssertableJSON(t, jsonStr)

	// Test at least one key exists
	assertable.HasAny([]string{"key1", "key2"})

	// Test no keys exist
	//assert.Panics(t, func() {
	//	assertable.HasAny([]string{"nonExistingKey1", "nonExistingKey2"})
	//})
}

func TestMissing(t *testing.T) {
	jsonStr := `{"key1": "value1"}`
	assertable, _ := NewAssertableJSON(t, jsonStr)

	// Test key is missing
	assertable.Missing("nonExistingKey")

	// Test key exists
	//assert.Panics(t, func() {
	//	assertable.Missing("key1")
	//})
}

func TestMissingAll(t *testing.T) {
	jsonStr := `{"key1": "value1"}`
	assertable, _ := NewAssertableJSON(t, jsonStr)

	// Test all keys are missing
	assertable.MissingAll([]string{"nonExistingKey1", "nonExistingKey2"})

	// Test one key exists
	//assert.Panics(t, func() {
	//	assertable.MissingAll([]string{"key1"})
	//})
}

func TestWhere(t *testing.T) {
	jsonStr := `{"key1": "value1", "key2": [1, 2, 3]}`
	assertable, _ := NewAssertableJSON(t, jsonStr)

	// Test correct value
	assertable.Where("key1", "value1")

	// Test incorrect value
	//assert.Panics(t, func() {
	//	assertable.Where("key1", "wrongValue").
	//		Where("key2", []any{1.0, 2.0, 3.0})
	//})
}

func TestWhereNot(t *testing.T) {
	jsonStr := `{"key1": "value1", "key2": [1, 2, 3]}`
	assertable, _ := NewAssertableJSON(t, jsonStr)

	// Test value is not as expected
	assertable.WhereNot("key1", "wrongValue")

	// Test value is as expected
	//assert.Panics(t, func() {
	//	assertable.WhereNot("key1", "value1")
	//})
}

func TestFirst(t *testing.T) {
	jsonStr := `{"items": [{"id": 1}, {"id": 2}]}`

	assertable, _ := NewAssertableJSON(t, jsonStr)

	// Test fetching the first item
	assertable.First("items", func(item contractstesting.AssertableJSON) {
		item.Where("id", float64(1)) // Verify the first item's id
	})

	// Test with a non-existing key
	//assert.Panics(t, func() {
	//	assertable.First("nonExistingKey", func(item contractstesting.AssertableJSON) {})
	//})

	// Test with an empty array
	//emptyJsonStr := `{"items": []}`
	//emptyAssertable, _ := NewAssertableJSON(t, emptyJsonStr)
	//assert.Panics(t, func() {
	//	emptyAssertable.First("items", func(item contractstesting.AssertableJSON) {})
	//})
}

func TestHasWithScope(t *testing.T) {
	jsonStr := `{"items": [{"id": 1}, {"id": 2}]}`

	assertable, _ := NewAssertableJSON(t, jsonStr)

	// Test has with correct length
	assertable.HasWithScope("items", 2, func(item contractstesting.AssertableJSON) {
		item.Where("id", float64(1))
	})

	// Test incorrect length
	//assert.Panics(t, func() {
	//	assertable.HasWithScope("items", 3, func(item contractstesting.AssertableJSON) {})
	//})

	// Test with a non-existing key
	//assert.Panics(t, func() {
	//	assertable.HasWithScope("nonExistingKey", 0, func(item contractstesting.AssertableJSON) {})
	//})
}

func TestEach(t *testing.T) {
	jsonStr := `{"items": [{"id": 1}, {"id": 2}]}`

	assertable, _ := NewAssertableJSON(t, jsonStr)

	// Test iterating over each item
	i := 1
	assertable.Each("items", func(item contractstesting.AssertableJSON) {
		item.Where("id", float64(i))
		i++
	})

	// Test with a non-existing key
	//assert.Panics(t, func() {
	//	assertable.Each("nonExistingKey", func(item contractstesting.AssertableJSON) {})
	//})

	// Test with an empty array
	emptyJsonStr := `{"items": []}`
	emptyAssertable, _ := NewAssertableJSON(t, emptyJsonStr)
	emptyAssertable.Each("items", func(item contractstesting.AssertableJSON) {})
}
