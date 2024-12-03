package processors

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/schema"
)

type MysqlTestSuite struct {
	suite.Suite
	mysql Mysql
}

func TestMysqlTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}

func (s *MysqlTestSuite) SetupTest() {
	s.mysql = NewMysql()
}

func (s *MysqlTestSuite) TestProcessColumns() {
	tests := []struct {
		name      string
		dbColumns []schema.DBColumn
		expected  []schema.Column
	}{
		{
			name: "ValidInput",
			dbColumns: []schema.DBColumn{
				{Name: "id", Type: "int", TypeName: "INT", Nullable: "NO", Extra: "auto_increment", Collation: "utf8_general_ci", Comment: "primary key", Default: "0"},
				{Name: "name", Type: "varchar", TypeName: "VARCHAR", Nullable: "YES", Extra: "", Collation: "utf8_general_ci", Comment: "user name", Default: ""},
			},
			expected: []schema.Column{
				{Autoincrement: true, Collation: "utf8_general_ci", Comment: "primary key", Default: "0", Name: "id", Nullable: false, Type: "int", TypeName: "INT"},
				{Autoincrement: false, Collation: "utf8_general_ci", Comment: "user name", Default: "", Name: "name", Nullable: true, Type: "varchar", TypeName: "VARCHAR"},
			},
		},
		{
			name:      "EmptyInput",
			dbColumns: []schema.DBColumn{},
		},
		{
			name: "NullableColumn",
			dbColumns: []schema.DBColumn{
				{Name: "description", Type: "text", TypeName: "TEXT", Nullable: "YES", Extra: "", Collation: "utf8_general_ci", Comment: "description", Default: ""},
			},
			expected: []schema.Column{
				{Autoincrement: false, Collation: "utf8_general_ci", Comment: "description", Default: "", Name: "description", Nullable: true, Type: "text", TypeName: "TEXT"},
			},
		},
		{
			name: "NonNullableColumn",
			dbColumns: []schema.DBColumn{
				{Name: "created_at", Type: "timestamp", TypeName: "TIMESTAMP", Nullable: "NO", Extra: "", Collation: "", Comment: "creation time", Default: "CURRENT_TIMESTAMP"},
			},
			expected: []schema.Column{
				{Autoincrement: false, Collation: "", Comment: "creation time", Default: "CURRENT_TIMESTAMP", Name: "created_at", Nullable: false, Type: "timestamp", TypeName: "TIMESTAMP"},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := s.mysql.ProcessColumns(tt.dbColumns)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *MysqlTestSuite) TestProcessForeignKeys() {
	tests := []struct {
		name          string
		dbForeignKeys []schema.DBForeignKey
		expected      []schema.ForeignKey
	}{
		{
			name: "ValidInput",
			dbForeignKeys: []schema.DBForeignKey{
				{Name: "fk_user_id", Columns: "user_id", ForeignSchema: "public", ForeignTable: "users", ForeignColumns: "id", OnUpdate: "CASCADE", OnDelete: "SET NULL"},
			},
			expected: []schema.ForeignKey{
				{Name: "fk_user_id", Columns: []string{"user_id"}, ForeignSchema: "public", ForeignTable: "users", ForeignColumns: []string{"id"}, OnUpdate: "cascade", OnDelete: "set null"},
			},
		},
		{
			name:          "EmptyInput",
			dbForeignKeys: []schema.DBForeignKey{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := s.mysql.ProcessForeignKeys(tt.dbForeignKeys)
			s.Equal(tt.expected, result)
		})
	}
}
