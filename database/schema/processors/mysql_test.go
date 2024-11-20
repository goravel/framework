package processors

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/database/schema"
)

func TestMysqlProcessColumns(t *testing.T) {
	tests := []struct {
		name      string
		dbColumns []DBColumn
		expected  []schema.Column
	}{
		{
			name: "ValidInput",
			dbColumns: []DBColumn{
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
			dbColumns: []DBColumn{},
		},
		{
			name: "NullableColumn",
			dbColumns: []DBColumn{
				{Name: "description", Type: "text", TypeName: "TEXT", Nullable: "YES", Extra: "", Collation: "utf8_general_ci", Comment: "description", Default: ""},
			},
			expected: []schema.Column{
				{Autoincrement: false, Collation: "utf8_general_ci", Comment: "description", Default: "", Name: "description", Nullable: true, Type: "text", TypeName: "TEXT"},
			},
		},
		{
			name: "NonNullableColumn",
			dbColumns: []DBColumn{
				{Name: "created_at", Type: "timestamp", TypeName: "TIMESTAMP", Nullable: "NO", Extra: "", Collation: "", Comment: "creation time", Default: "CURRENT_TIMESTAMP"},
			},
			expected: []schema.Column{
				{Autoincrement: false, Collation: "", Comment: "creation time", Default: "CURRENT_TIMESTAMP", Name: "created_at", Nullable: false, Type: "timestamp", TypeName: "TIMESTAMP"},
			},
		},
	}

	mysql := NewMysql()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mysql.ProcessColumns(tt.dbColumns)
			assert.Equal(t, tt.expected, result)
		})
	}
}
