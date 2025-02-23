package http

import (
	"testing"

	"github.com/stretchr/testify/suite"

	contractshttp "github.com/goravel/framework/contracts/testing/http"
	"github.com/goravel/framework/foundation/json"
)

type AssertableJsonTestSuite struct {
	suite.Suite
}

func TestAssertableJsonTestSuite(t *testing.T) {
	suite.Run(t, new(AssertableJsonTestSuite))
}

func (s *AssertableJsonTestSuite) SetupTest() {
}

func (s *AssertableJsonTestSuite) TestNewAssertableJSON() {
	validJSON := `{"key1": "value1", "key2": [1, 2, 3]}`
	invalidJSON := `{"key1": "value1", "key2": [1, 2, 3]`

	assertable, err := NewAssertableJSON(s.T(), json.NewJson(), validJSON)
	s.NoError(err)
	s.NotNil(assertable)

	assertable, err = NewAssertableJSON(s.T(), json.NewJson(), invalidJSON)
	s.Error(err)
	s.Nil(assertable)
}

func (s *AssertableJsonTestSuite) TestCount() {
	jsonStr := `{"items": [1, 2, 3], "otherKey": "value"}`
	assertable, err := NewAssertableJSON(s.T(), json.NewJson(), jsonStr)
	s.NoError(err)

	assertable.Count("items", 3)

	//assertable.Count("items", 4)
}

func (s *AssertableJsonTestSuite) TestHas() {
	jsonStr := `{
		"key1": "value1",
		"key2": [1, 2, 3],
		"nested": {"deep": "value"},
		"nullKey": null
	}`
	assertable, err := NewAssertableJSON(s.T(), json.NewJson(), jsonStr)
	s.NoError(err)

	assertable.Has("key1")
	assertable.Has("nullKey")

	//assertable.Has("nonExistingKey")
}

func (s *AssertableJsonTestSuite) TestHasAll() {
	jsonStr := `{"key1": "value1", "key2": [1, 2, 3]}`
	assertable, err := NewAssertableJSON(s.T(), json.NewJson(), jsonStr)
	s.NoError(err)

	// Test all keys exist
	assertable.HasAll([]string{"key1", "key2"})

	// Test one key does not exist
	//assertable.HasAll([]string{"key1", "nonExistingKey"})
}

func (s *AssertableJsonTestSuite) TestHasAny() {
	jsonStr := `{"key1": "value1", "key2": [1, 2, 3]}`
	assertable, err := NewAssertableJSON(s.T(), json.NewJson(), jsonStr)
	s.NoError(err)

	// Test at least one key exists
	assertable.HasAny([]string{"key1", "key2"})

	// Test no keys exist
	//assertable.HasAny([]string{"nonExistingKey1", "nonExistingKey2"})
}

func (s *AssertableJsonTestSuite) TestMissing() {
	jsonStr := `{"key1": "value1"}`
	assertable, err := NewAssertableJSON(s.T(), json.NewJson(), jsonStr)
	s.NoError(err)

	// Test key is missing
	assertable.Missing("nonExistingKey")

	// Test key exists
	//assertable.Missing("key1")
}

func (s *AssertableJsonTestSuite) TestMissingAll() {
	jsonStr := `{"key1": "value1"}`
	assertable, err := NewAssertableJSON(s.T(), json.NewJson(), jsonStr)
	s.NoError(err)

	// Test all keys are missing
	assertable.MissingAll([]string{"nonExistingKey1", "nonExistingKey2"})

	// Test one key exists
	//assertable.MissingAll([]string{"key1"})
}

func (s *AssertableJsonTestSuite) TestWhere() {
	jsonStr := `{
		"key1": "value1",
		"intKey": 42,
     	"floatKey": 42.0,
		"nullKey": null,
		"objKey": {"nested": "value"},
		"arrayKey": [1, 2, 3]
	}`
	assertable, err := NewAssertableJSON(s.T(), json.NewJson(), jsonStr)
	s.NoError(err)

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

func (s *AssertableJsonTestSuite) TestWhereNot() {
	jsonStr := `{"key1": "value1", "key2": [1, 2, 3]}`
	assertable, err := NewAssertableJSON(s.T(), json.NewJson(), jsonStr)
	s.NoError(err)

	// Test value is not as expected
	assertable.WhereNot("key1", "wrongValue")

	// Test value is as expected
	//assertable.WhereNot("key1", "value1")
}

func (s *AssertableJsonTestSuite) TestFirst() {
	jsonStr := `{"items": [{"id": 1}, {"id": 2}]}`

	assertable, err := NewAssertableJSON(s.T(), json.NewJson(), jsonStr)
	s.NoError(err)

	// Test fetching the first item
	assertable.First("items", func(item contractshttp.AssertableJSON) {
		item.Where("id", float64(1))
	})

	// Test with a non-existing key
	//assertable.First("nonExistingKey", func(item contractstesting.AssertableJSON) {})

	// Test with an empty array
	//emptyJsonStr := `{"items": []}`
	//emptyAssertable, _ := NewAssertableJSON(t, emptyJsonStr)
	//emptyAssertable.First("items", func(item contractstesting.AssertableJSON) {})
}

func (s *AssertableJsonTestSuite) TestHasWithScope() {
	jsonStr := `{"items": [{"id": 1}, {"id": 2}]}`

	assertable, err := NewAssertableJSON(s.T(), json.NewJson(), jsonStr)
	s.NoError(err)

	// Test has with correct length
	assertable.HasWithScope("items", 2, func(item contractshttp.AssertableJSON) {
		item.Where("id", float64(1))
	})

	// Test incorrect length
	//assertable.HasWithScope("items", 3, func(item contractstesting.AssertableJSON) {})

	// Test with a non-existing key
	//assertable.HasWithScope("nonExistingKey", 0, func(item contractstesting.AssertableJSON) {})
}

func (s *AssertableJsonTestSuite) TestEach() {
	jsonStr := `{
		"items": [{"id": 1}, {"id": 2}],
		"mixedTypes": [42, "string", {"key": "value"}],
		"nonArray": "value"
	}`

	assertable, err := NewAssertableJSON(s.T(), json.NewJson(), jsonStr)
	s.NoError(err)

	// Test iterating over each item
	callCount := 0
	assertable.Each("items", func(item contractshttp.AssertableJSON) {
		item.Where("id", float64(callCount+1))
		callCount++
	})
	s.Equal(2, callCount)

	// Test with a non-existing key
	//assertable.Each("nonExistingKey", func(item contractstesting.AssertableJSON) {})

	// Test with an empty array
	emptyJsonStr := `{"items": []}`
	emptyAssertable, err := NewAssertableJSON(s.T(), json.NewJson(), emptyJsonStr)
	s.NoError(err)

	emptyCallCount := 0
	emptyAssertable.Each("items", func(item contractshttp.AssertableJSON) {
		emptyCallCount++
	})
	s.Equal(0, emptyCallCount)
}
