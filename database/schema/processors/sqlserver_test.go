package processors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/schema"
)

type SqlserverTestSuite struct {
	suite.Suite
	sqlserver Sqlserver
}

func TestSqlserverTestSuite(t *testing.T) {
	suite.Run(t, new(SqlserverTestSuite))
}

func (s *SqlserverTestSuite) SetupTest() {
	s.sqlserver = NewSqlserver()
}

func (s *SqlserverTestSuite) TestProcessColumns() {
	tests := []struct {
		name      string
		dbColumns []schema.DBColumn
		expected  []schema.Column
	}{
		{
			name: "ValidInput",
			dbColumns: []schema.DBColumn{
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
			dbColumns: []schema.DBColumn{},
		},
		{
			name: "NullableColumn",
			dbColumns: []schema.DBColumn{
				{Name: "description", TypeName: "text", Nullable: "true", Collation: "utf8_general_ci", Comment: "description", Default: "default_description"},
			},
			expected: []schema.Column{
				{Autoincrement: false, Collation: "utf8_general_ci", Comment: "description", Default: "default_description", Name: "description", Nullable: true, Type: "text", TypeName: "text"},
			},
		},
		{
			name: "NonNullableColumn",
			dbColumns: []schema.DBColumn{
				{Name: "created_at", TypeName: "timestamp", Nullable: "false", Collation: "", Comment: "creation time", Default: "CURRENT_TIMESTAMP"},
			},
			expected: []schema.Column{
				{Autoincrement: false, Collation: "", Comment: "creation time", Default: "CURRENT_TIMESTAMP", Name: "created_at", Nullable: false, Type: "timestamp", TypeName: "timestamp"},
			},
		},
	}

	sqlserver := NewSqlserver()
	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := sqlserver.ProcessColumns(tt.dbColumns)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *SqlserverTestSuite) TestProcessForeignKeys() {
	tests := []struct {
		name          string
		dbForeignKeys []schema.DBForeignKey
		expected      []schema.ForeignKey
	}{
		{
			name: "ValidInput",
			dbForeignKeys: []schema.DBForeignKey{
				{Name: "fk_user_id", Columns: "user_id", ForeignSchema: "dbo", ForeignTable: "users", ForeignColumns: "id", OnUpdate: "CASCADE", OnDelete: "SET_NULL"},
			},
			expected: []schema.ForeignKey{
				{Name: "fk_user_id", Columns: []string{"user_id"}, ForeignSchema: "dbo", ForeignTable: "users", ForeignColumns: []string{"id"}, OnUpdate: "cascade", OnDelete: "set null"},
			},
		},
		{
			name:          "EmptyInput",
			dbForeignKeys: []schema.DBForeignKey{},
		},
	}

	sqlserver := NewSqlserver()
	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := sqlserver.ProcessForeignKeys(tt.dbForeignKeys)
			s.Equal(tt.expected, result)
		})
	}
}

func TestGetType(t *testing.T) {
	tests := []struct {
		name     string
		dbColumn schema.DBColumn
		expected string
	}{
		{
			name:     "BinaryWithMaxLength",
			dbColumn: schema.DBColumn{TypeName: "binary", Length: -1},
			expected: "binary(max)",
		},
		{
			name:     "VarbinaryWithSpecificLength",
			dbColumn: schema.DBColumn{TypeName: "varbinary", Length: 255},
			expected: "varbinary(255)",
		},
		{
			name:     "CharWithSpecificLength",
			dbColumn: schema.DBColumn{TypeName: "char", Length: 10},
			expected: "char(10)",
		},
		{
			name:     "DecimalWithPrecisionAndScale",
			dbColumn: schema.DBColumn{TypeName: "decimal", Precision: 10, Places: 2},
			expected: "decimal(10,2)",
		},
		{
			name:     "FloatWithPrecision",
			dbColumn: schema.DBColumn{TypeName: "float", Precision: 5},
			expected: "float(5)",
		},
		{
			name:     "DefaultTypeName",
			dbColumn: schema.DBColumn{TypeName: "int"},
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
