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
	assertable, err := NewAssertableJSON(t, jsonStr)
	assert.NoError(t, err)

	assertable.Count("items", 3)

	//assertable.Count("items", 4)
}

func TestHas(t *testing.T) {
	jsonStr := `{
		"key1": "value1",
		"key2": [1, 2, 3],
		"nested": {"deep": "value"},
		"nullKey": null
	}`
	assertable, err := NewAssertableJSON(t, jsonStr)
	assert.NoError(t, err)

	assertable.Has("key1")
	assertable.Has("nullKey")

	//assertable.Has("nonExistingKey")
}

func TestHasAll(t *testing.T) {
	jsonStr := `{"key1": "value1", "key2": [1, 2, 3]}`
	assertable, err := NewAssertableJSON(t, jsonStr)
	assert.NoError(t, err)

	// Test all keys exist
	assertable.HasAll([]string{"key1", "key2"})

	// Test one key does not exist
	//assertable.HasAll([]string{"key1", "nonExistingKey"})
}

func TestHasAny(t *testing.T) {
	jsonStr := `{"key1": "value1", "key2": [1, 2, 3]}`
	assertable, err := NewAssertableJSON(t, jsonStr)
	assert.NoError(t, err)

	// Test at least one key exists
	assertable.HasAny([]string{"key1", "key2"})

	// Test no keys exist
	//assertable.HasAny([]string{"nonExistingKey1", "nonExistingKey2"})
}

func TestMissing(t *testing.T) {
	jsonStr := `{"key1": "value1"}`
	assertable, err := NewAssertableJSON(t, jsonStr)
	assert.NoError(t, err)

	// Test key is missing
	assertable.Missing("nonExistingKey")

	// Test key exists
	//assertable.Missing("key1")
}

func TestMissingAll(t *testing.T) {
	jsonStr := `{"key1": "value1"}`
	assertable, err := NewAssertableJSON(t, jsonStr)
	assert.NoError(t, err)

	// Test all keys are missing
	assertable.MissingAll([]string{"nonExistingKey1", "nonExistingKey2"})

	// Test one key exists
	//assertable.MissingAll([]string{"key1"})
}

func TestWhere(t *testing.T) {
	jsonStr := `{
		"key1": "value1",
		"intKey": 42,
     	"floatKey": 42.0,
		"nullKey": null,
		"objKey": {"nested": "value"},
		"arrayKey": [1, 2, 3]
	}`
	assertable, err := NewAssertableJSON(t, jsonStr)
	assert.NoError(t, err)

	// Test correct value
	assertable.Where("key1", "value1")
	// Test number type handling
	assertable.Where("intKey", float64(42))
	assertable.Where("floatKey", float64(42))
	// Test null
	assertable.Where("nullKey", nil)
	// Test object equality
	assertable.Where("objKey", map[string]any{"nested": "value"})
	// Test array equality
	assertable.Where("arrayKey", []any{float64(1), float64(2), float64(3)})

	// Test incorrect value
	//assertable.Where("key1", "wrongValue")
}

func TestWhereNot(t *testing.T) {
	jsonStr := `{"key1": "value1", "key2": [1, 2, 3]}`
	assertable, err := NewAssertableJSON(t, jsonStr)
	assert.NoError(t, err)

	// Test value is not as expected
	assertable.WhereNot("key1", "wrongValue")

	// Test value is as expected
	//assertable.WhereNot("key1", "value1")
}

func TestFirst(t *testing.T) {
	jsonStr := `{"items": [{"id": 1}, {"id": 2}]}`

	assertable, err := NewAssertableJSON(t, jsonStr)
	assert.NoError(t, err)

	// Test fetching the first item
	assertable.First("items", func(item contractstesting.AssertableJSON) {
		item.Where("id", float64(1))
	})

	// Test with a non-existing key
	//assertable.First("nonExistingKey", func(item contractstesting.AssertableJSON) {})

	// Test with an empty array
	//emptyJsonStr := `{"items": []}`
	//emptyAssertable, _ := NewAssertableJSON(t, emptyJsonStr)
	//emptyAssertable.First("items", func(item contractstesting.AssertableJSON) {})
}

func TestHasWithScope(t *testing.T) {
	jsonStr := `{"items": [{"id": 1}, {"id": 2}]}`

	assertable, err := NewAssertableJSON(t, jsonStr)
	assert.NoError(t, err)

	// Test has with correct length
	assertable.HasWithScope("items", 2, func(item contractstesting.AssertableJSON) {
		item.Where("id", float64(1))
	})

	// Test incorrect length
	//assertable.HasWithScope("items", 3, func(item contractstesting.AssertableJSON) {})

	// Test with a non-existing key
	//assertable.HasWithScope("nonExistingKey", 0, func(item contractstesting.AssertableJSON) {})
}

func TestEach(t *testing.T) {
	jsonStr := `{
		"items": [{"id": 1}, {"id": 2}],
		"mixedTypes": [42, "string", {"key": "value"}],
		"nonArray": "value"
	}`

	assertable, err := NewAssertableJSON(t, jsonStr)
	assert.NoError(t, err)

	// Test iterating over each item
	callCount := 0
	assertable.Each("items", func(item contractstesting.AssertableJSON) {
		item.Where("id", float64(callCount+1))
		callCount++
	})
	assert.Equal(t, 2, callCount)

	// Test with a non-existing key
	//assertable.Each("nonExistingKey", func(item contractstesting.AssertableJSON) {})

	// Test with an empty array
	emptyJsonStr := `{"items": []}`
	emptyAssertable, err := NewAssertableJSON(t, emptyJsonStr)
	assert.NoError(t, err)
	emptyCallCount := 0
	emptyAssertable.Each("items", func(item contractstesting.AssertableJSON) {
		emptyCallCount++
	})
	assert.Equal(t, 0, emptyCallCount)
}
