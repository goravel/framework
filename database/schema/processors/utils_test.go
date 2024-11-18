package processors

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/database/schema"
)

func TestProcessIndexes(t *testing.T) {
	// Test with valid indexes
	input := []DBIndex{
		{Name: "INDEX_A", Type: "BTREE", Columns: "a,b"},
		{Name: "INDEX_B", Type: "HASH", Columns: "c,d"},
	}
	expected := []schema.Index{
		{Name: "index_a", Type: "btree", Columns: []string{"a", "b"}},
		{Name: "index_b", Type: "hash", Columns: []string{"c", "d"}},
	}

	result := processIndexes(input)

	assert.Equal(t, expected, result)

	// Test with empty input
	input = []DBIndex{}

	result = processIndexes(input)

	assert.Nil(t, result)
}
