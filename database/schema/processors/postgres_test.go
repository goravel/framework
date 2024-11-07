package processors

import (
	"testing"

	"github.com/gookit/goutil/testutil/assert"

	"github.com/goravel/framework/contracts/database/schema"
)

func TestProcessIndexes(t *testing.T) {
	// Test with valid indexes
	input := []schema.Index{
		{Name: "INDEX_A", Type: "BTREE"},
		{Name: "INDEX_B", Type: "HASH"},
	}
	expected := []schema.Index{
		{Name: "index_a", Type: "btree"},
		{Name: "index_b", Type: "hash"},
	}

	postgres := NewPostgres()
	result := postgres.ProcessIndexes(input)

	assert.Equal(t, expected, result)

	// Test with empty input
	input = []schema.Index{}
	expected = []schema.Index{}

	result = postgres.ProcessIndexes(input)

	assert.Equal(t, expected, result)
}

func TestProcessTypes(t *testing.T) {
	// ValidTypes_ReturnsProcessedTypes
	input := []schema.Type{
		{Type: "b", Category: "a"},
		{Type: "c", Category: "b"},
		{Type: "d", Category: "c"},
	}
	expected := []schema.Type{
		{Type: "base", Category: "array"},
		{Type: "composite", Category: "boolean"},
		{Type: "domain", Category: "composite"},
	}

	postgres := NewPostgres()
	result := postgres.ProcessTypes(input)

	assert.Equal(t, expected, result)

	// UnknownType_ReturnsEmptyString
	input = []schema.Type{
		{Type: "unknown", Category: "a"},
	}
	expected = []schema.Type{
		{Type: "", Category: "array"},
	}

	result = postgres.ProcessTypes(input)

	assert.Equal(t, expected, result)

	// UnknownCategory_ReturnsEmptyString
	input = []schema.Type{
		{Type: "b", Category: "unknown"},
	}
	expected = []schema.Type{
		{Type: "base", Category: ""},
	}

	result = postgres.ProcessTypes(input)

	assert.Equal(t, expected, result)

	// EmptyInput_ReturnsEmptyOutput
	input = []schema.Type{}
	expected = []schema.Type{}

	result = postgres.ProcessTypes(input)

	assert.Equal(t, expected, result)
}
