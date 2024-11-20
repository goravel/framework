package processors

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/database/schema"
)

func TestPostgresProcessColumns(t *testing.T) {
	tests := []struct {
		name      string
		dbColumns []DBColumn
		expected  []schema.Column
	}{
		{
			name: "ValidInput",
			dbColumns: []DBColumn{
				{Name: "id", Type: "int", TypeName: "INT", Nullable: "NO", Extra: "auto_increment", Collation: "utf8_general_ci", Comment: "primary key", Default: "nextval('id_seq'::regclass)"},
				{Name: "name", Type: "varchar", TypeName: "VARCHAR", Nullable: "true", Extra: "", Collation: "utf8_general_ci", Comment: "user name", Default: ""},
			},
			expected: []schema.Column{
				{Autoincrement: true, Collation: "utf8_general_ci", Comment: "primary key", Default: "nextval('id_seq'::regclass)", Name: "id", Nullable: false, Type: "int", TypeName: "INT"},
				{Autoincrement: false, Collation: "utf8_general_ci", Comment: "user name", Default: "", Name: "name", Nullable: true, Type: "varchar", TypeName: "VARCHAR"},
			},
		},
		{
			name:      "EmptyInput",
			dbColumns: []DBColumn{},
		},
		{
			name: "NullableColumn",
			dbColumns: []DBColumn{
				{Name: "description", Type: "text", TypeName: "TEXT", Nullable: "true", Extra: "", Collation: "utf8_general_ci", Comment: "description", Default: ""},
			},
			expected: []schema.Column{
				{Autoincrement: false, Collation: "utf8_general_ci", Comment: "description", Default: "", Name: "description", Nullable: true, Type: "text", TypeName: "TEXT"},
			},
		},
		{
			name: "NonNullableColumn",
			dbColumns: []DBColumn{
				{Name: "created_at", Type: "timestamp", TypeName: "TIMESTAMP", Nullable: "false", Extra: "", Collation: "", Comment: "creation time", Default: "CURRENT_TIMESTAMP"},
			},
			expected: []schema.Column{
				{Autoincrement: false, Collation: "", Comment: "creation time", Default: "CURRENT_TIMESTAMP", Name: "created_at", Nullable: false, Type: "timestamp", TypeName: "TIMESTAMP"},
			},
		},
	}

	postgres := NewPostgres()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := postgres.ProcessColumns(tt.dbColumns)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPostgresProcessTypes(t *testing.T) {
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
