package processors

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/database/schema"
)

func TestProcessColumns(t *testing.T) {
	tests := []struct {
		name      string
		dbColumns []DBColumn
		expected  []schema.Column
	}{
		{
			name: "ValidInput",
			dbColumns: []DBColumn{
				{Name: "id", Type: "integer", Nullable: "false", Primary: true, Default: "1"},
				{Name: "name", Type: "varchar", Nullable: "true", Default: "default_name"},
			},
			expected: []schema.Column{
				{Autoincrement: true, Default: "1", Name: "id", Nullable: false, Type: "integer"},
				{Autoincrement: false, Default: "default_name", Name: "name", Nullable: true, Type: "varchar"},
			},
		},
		{
			name:      "EmptyInput",
			dbColumns: []DBColumn{},
		},
		{
			name: "NullableColumn",
			dbColumns: []DBColumn{
				{Name: "description", Type: "text", Nullable: "true", Default: "default_description"},
			},
			expected: []schema.Column{
				{Autoincrement: false, Default: "default_description", Name: "description", Nullable: true, Type: "text"},
			},
		},
		{
			name: "NonNullableColumn",
			dbColumns: []DBColumn{
				{Name: "created_at", Type: "timestamp", Nullable: "false", Default: "CURRENT_TIMESTAMP"},
			},
			expected: []schema.Column{
				{Autoincrement: false, Default: "CURRENT_TIMESTAMP", Name: "created_at", Nullable: false, Type: "timestamp"},
			},
		},
	}

	sqlite := NewSqlite()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sqlite.ProcessColumns(tt.dbColumns)
			assert.Equal(t, tt.expected, result)
		})
	}
}

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
