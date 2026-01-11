package http

import (
	"fmt"
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

	assertable, err := NewAssertableJSON(s.T(), json.New(), validJSON)
	s.NoError(err)
	s.NotNil(assertable)

	assertable, err = NewAssertableJSON(s.T(), json.New(), invalidJSON)
	s.Error(err)
	s.Nil(assertable)
}

func (s *AssertableJsonTestSuite) TestCount() {
	jsonStr := `{"items": [1, 2, 3], "otherKey": "value"}`
	assertable, err := NewAssertableJSON(s.T(), json.New(), jsonStr)
	s.NoError(err)

	assertable.Count("items", 3)

	mockT := &MockTestingT{}
	mockAssertable, _ := NewAssertableJSON(mockT, json.New(), jsonStr)
	mockAssertable.Count("items", 4)
	s.True(mockT.Failed)
	s.Contains(mockT.ErrorMessages[0], "expected size")
}

func (s *AssertableJsonTestSuite) TestHas() {
	jsonStr := `{
       "key1": "value1",
       "key2": [1, 2, 3],
       "nested": {"deep": "value"},
       "nullKey": null
    }`
	assertable, err := NewAssertableJSON(s.T(), json.New(), jsonStr)
	s.NoError(err)

	assertable.Has("key1").Has("nullKey")

	mockT := &MockTestingT{}
	mockAssertable, _ := NewAssertableJSON(mockT, json.New(), jsonStr)
	mockAssertable.Has("nonExistingKey")
	s.True(mockT.Failed)
	s.Contains(mockT.ErrorMessages[0], "does not exist")
}

func (s *AssertableJsonTestSuite) TestHasAll() {
	jsonStr := `{"key1": "value1", "key2": [1, 2, 3]}`
	assertable, err := NewAssertableJSON(s.T(), json.New(), jsonStr)
	s.NoError(err)

	assertable.HasAll([]string{"key1", "key2"})

	mockT := &MockTestingT{}
	mockAssertable, _ := NewAssertableJSON(mockT, json.New(), jsonStr)
	mockAssertable.HasAll([]string{"key1", "nonExistingKey"})
	s.True(mockT.Failed)
	s.Contains(mockT.ErrorMessages[0], "does not exist")
}

func (s *AssertableJsonTestSuite) TestHasAny() {
	jsonStr := `{"key1": "value1", "key2": [1, 2, 3]}`
	assertable, err := NewAssertableJSON(s.T(), json.New(), jsonStr)
	s.NoError(err)

	assertable.HasAny([]string{"key1", "key2"})

	mockT := &MockTestingT{}
	mockAssertable, _ := NewAssertableJSON(mockT, json.New(), jsonStr)
	mockAssertable.HasAny([]string{"nonExistingKey1", "nonExistingKey2"})
	s.True(mockT.Failed)
	s.Contains(mockT.ErrorMessages[0], "None of properties")
}

func (s *AssertableJsonTestSuite) TestMissing() {
	jsonStr := `{"key1": "value1"}`
	assertable, err := NewAssertableJSON(s.T(), json.New(), jsonStr)
	s.NoError(err)

	assertable.Missing("nonExistingKey")

	mockT := &MockTestingT{}
	mockAssertable, _ := NewAssertableJSON(mockT, json.New(), jsonStr)
	mockAssertable.Missing("key1")
	s.True(mockT.Failed)
	s.Contains(mockT.ErrorMessages[0], "expected to be missing")
}

func (s *AssertableJsonTestSuite) TestMissingAll() {
	jsonStr := `{"key1": "value1"}`
	assertable, err := NewAssertableJSON(s.T(), json.New(), jsonStr)
	s.NoError(err)

	assertable.MissingAll([]string{"nonExistingKey1", "nonExistingKey2"})

	mockT := &MockTestingT{}
	mockAssertable, _ := NewAssertableJSON(mockT, json.New(), jsonStr)
	mockAssertable.MissingAll([]string{"key1"})
	s.True(mockT.Failed)
	s.Contains(mockT.ErrorMessages[0], "expected to be missing")
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
	assertable, err := NewAssertableJSON(s.T(), json.New(), jsonStr)
	s.NoError(err)

	assertable.Where("key1", "value1").
		Where("intKey", float64(42)).
		Where("floatKey", float64(42)).
		Where("nullKey", nil).
		Where("objKey", map[string]any{"nested": "value"}).
		Where("arrayKey", []any{float64(1), float64(2), float64(3)})

	mockT := &MockTestingT{}
	mockAssertable, err := NewAssertableJSON(mockT, json.New(), jsonStr)
	s.NoError(err)
	mockAssertable.Where("key1", "wrongValue")
	s.True(mockT.Failed)
	s.Contains(mockT.ErrorMessages[0], "Expected property [key1]")
}

func (s *AssertableJsonTestSuite) TestWhereNot() {
	jsonStr := `{"key1": "value1", "key2": [1, 2, 3]}`
	assertable, err := NewAssertableJSON(s.T(), json.New(), jsonStr)
	s.NoError(err)

	assertable.WhereNot("key1", "wrongValue")

	mockT := &MockTestingT{}
	mockAssertable, _ := NewAssertableJSON(mockT, json.New(), jsonStr)
	mockAssertable.WhereNot("key1", "value1")
	s.True(mockT.Failed)
	s.Contains(mockT.ErrorMessages[0], "to not have value")
}

func (s *AssertableJsonTestSuite) TestFirst() {
	jsonStr := `{"items": [{"id": 1}, {"id": 2}]}`
	assertable, err := NewAssertableJSON(s.T(), json.New(), jsonStr)
	s.NoError(err)

	assertable.First("items", func(item contractshttp.AssertableJSON) {
		item.Where("id", float64(1))
	})

	mockT := &MockTestingT{}
	mockAssertable, _ := NewAssertableJSON(mockT, json.New(), jsonStr)
	mockAssertable.First("nonExistingKey", func(item contractshttp.AssertableJSON) {})
	s.True(mockT.Failed)
	s.Contains(mockT.ErrorMessages[0], "does not exist")

	emptyJsonStr := `{"items": []}`
	mockT2 := &MockTestingT{}
	emptyAssertable, _ := NewAssertableJSON(mockT2, json.New(), emptyJsonStr)
	emptyAssertable.First("items", func(item contractshttp.AssertableJSON) {})
	s.True(mockT2.Failed)
	s.Contains(mockT2.ErrorMessages[0], "non-empty array")
}

func (s *AssertableJsonTestSuite) TestHasWithScope() {
	jsonStr := `{"items": [{"id": 1}, {"id": 2}]}`
	assertable, err := NewAssertableJSON(s.T(), json.New(), jsonStr)
	s.NoError(err)

	assertable.HasWithScope("items", 2, func(item contractshttp.AssertableJSON) {
		item.Where("id", float64(1))
	})

	mockT := &MockTestingT{}
	mockAssertable, _ := NewAssertableJSON(mockT, json.New(), jsonStr)
	mockAssertable.HasWithScope("items", 3, func(item contractshttp.AssertableJSON) {})
	s.True(mockT.Failed)
	s.Contains(mockT.ErrorMessages[0], "expected length")
}

func (s *AssertableJsonTestSuite) TestEach() {
	jsonStr := `{
       "items": [{"id": 1}, {"id": 2}],
       "mixedTypes": [42, "string", {"key": "value"}],
       "nonArray": "value"
    }`
	assertable, err := NewAssertableJSON(s.T(), json.New(), jsonStr)
	s.NoError(err)

	callCount := 0
	assertable.Each("items", func(item contractshttp.AssertableJSON) {
		item.Where("id", float64(callCount+1))
		callCount++
	})
	s.Equal(2, callCount)

	mockT := &MockTestingT{}
	mockAssertable, _ := NewAssertableJSON(mockT, json.New(), jsonStr)
	mockAssertable.Each("nonExistingKey", func(item contractshttp.AssertableJSON) {})
	s.True(mockT.Failed)
	s.Contains(mockT.ErrorMessages[0], "does not exist")
}

// MockTestingT captures assertions instead of failing the test suite.
type MockTestingT struct {
	Failed        bool
	ErrorMessages []string
}

func (r *MockTestingT) Errorf(format string, args ...interface{}) {
	r.Failed = true
	r.ErrorMessages = append(r.ErrorMessages, fmt.Sprintf(format, args...))
}

func (r *MockTestingT) FailNow() {
	r.Failed = true
}
