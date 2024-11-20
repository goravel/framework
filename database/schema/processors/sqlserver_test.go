package processors

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/database/schema"
)

func TestGetType(t *testing.T) {
	tests := []struct {
		name     string
		dbColumn DBColumn
		expected string
	}{
		{
			name:     "BinaryWithMaxLength",
			dbColumn: DBColumn{TypeName: "binary", Length: -1},
			expected: "binary(max)",
		},
		{
			name:     "VarbinaryWithSpecificLength",
			dbColumn: DBColumn{TypeName: "varbinary", Length: 255},
			expected: "varbinary(255)",
		},
		{
			name:     "CharWithSpecificLength",
			dbColumn: DBColumn{TypeName: "char", Length: 10},
			expected: "char(10)",
		},
		{
			name:     "DecimalWithPrecisionAndScale",
			dbColumn: DBColumn{TypeName: "decimal", Precision: 10, Places: 2},
			expected: "decimal(10,2)",
		},
		{
			name:     "FloatWithPrecision",
			dbColumn: DBColumn{TypeName: "float", Precision: 5},
			expected: "float(5)",
		},
		{
			name:     "DefaultTypeName",
			dbColumn: DBColumn{TypeName: "int"},
			expected: "int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getType(tt.dbColumn)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSqlserverProcessColumns(t *testing.T) {
	tests := []struct {
		name      string
		dbColumns []DBColumn
		expected  []schema.Column
	}{
		{
			name: "ValidInput",
			dbColumns: []DBColumn{
				{Name: "id", TypeName: "int", Nullable: "false", Autoincrement: true, Collation: "utf8_general_ci", Comment: "primary key", Default: "1"},
				{Name: "name", TypeName: "varchar", Nullable: "true", Collation: "utf8_general_ci", Comment: "user name", Default: "default_name", Length: 10},
			},
			expected: []schema.Column{
				{Autoincrement: true, Collation: "utf8_general_ci", Comment: "primary key", Default: "1", Name: "id", Nullable: false, Type: "int", TypeName: "int"},
				{Autoincrement: false, Collation: "utf8_general_ci", Comment: "user name", Default: "default_name", Name: "name", Nullable: true, Type: "varchar(10)", TypeName: "varchar"},
			},
		},
		{
			name:      "EmptyInput",
			dbColumns: []DBColumn{},
		},
		{
			name: "NullableColumn",
			dbColumns: []DBColumn{
				{Name: "description", TypeName: "text", Nullable: "true", Collation: "utf8_general_ci", Comment: "description", Default: "default_description"},
			},
			expected: []schema.Column{
				{Autoincrement: false, Collation: "utf8_general_ci", Comment: "description", Default: "default_description", Name: "description", Nullable: true, Type: "text", TypeName: "text"},
			},
		},
		{
			name: "NonNullableColumn",
			dbColumns: []DBColumn{
				{Name: "created_at", TypeName: "timestamp", Nullable: "false", Collation: "", Comment: "creation time", Default: "CURRENT_TIMESTAMP"},
			},
			expected: []schema.Column{
				{Autoincrement: false, Collation: "", Comment: "creation time", Default: "CURRENT_TIMESTAMP", Name: "created_at", Nullable: false, Type: "timestamp", TypeName: "timestamp"},
			},
		},
	}

	sqlserver := NewSqlserver()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sqlserver.ProcessColumns(tt.dbColumns)
			assert.Equal(t, tt.expected, result)
		})
	}
}
