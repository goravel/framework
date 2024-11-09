package processors

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/database/schema"
)

func TestSqliteProcessIndexes(t *testing.T) {
	// Test with valid indexes
	input := []DBIndex{
		{Name: "INDEX_A", Type: "BTREE", Columns: "a,b"},
		{Name: "INDEX_B", Type: "HASH", Columns: "c,d"},
		{Name: "INDEX_C", Type: "HASH", Columns: "e,f", Primary: true},
	}
	expected := []schema.Index{
		{Name: "index_a", Columns: []string{"a", "b"}},
		{Name: "index_b", Columns: []string{"c", "d"}},
		{Name: "index_c", Columns: []string{"e", "f"}, Primary: true},
	}

	sqlite := NewSqlite()
	result := sqlite.ProcessIndexes(input)

	assert.Equal(t, expected, result)

	// Test with valid indexes with multiple primary keys
	input = []DBIndex{
		{Name: "INDEX_A", Type: "BTREE", Columns: "a,b"},
		{Name: "INDEX_B", Type: "HASH", Columns: "c,d"},
		{Name: "INDEX_C", Type: "HASH", Columns: "e,f", Primary: true},
		{Name: "INDEX_D", Type: "HASH", Columns: "g,h", Primary: true},
	}
	expected = []schema.Index{
		{Name: "index_a", Columns: []string{"a", "b"}},
		{Name: "index_b", Columns: []string{"c", "d"}},
	}

	result = sqlite.ProcessIndexes(input)

	assert.Equal(t, expected, result)

	// Test with empty input
	input = []DBIndex{}

	result = sqlite.ProcessIndexes(input)

	assert.Nil(t, result)
}
