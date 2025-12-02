package migration

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/errors"
)

type User struct {
	orm.Model
	Name string
}

type Product struct {
	ID    uint   `gorm:"primaryKey"`
	Code  string `gorm:"uniqueIndex"`
	Price float64
}

type CompositeKey struct {
	OrderID   uint `gorm:"primaryKey"`
	ProductID uint `gorm:"primaryKey"`
	Note      string
}

type IgnoredField struct {
	ID     uint
	Secret string `gorm:"-:migration"`
}

type CustomTypes struct {
	ID       uint
	Status   string         `gorm:"type:enum('on','off')"`
	Settings map[string]any `gorm:"type:json"`
}

type PtrFields struct {
	ID   uint
	Name *string
	Age  *int
}

type NullableFields struct {
	ID      uint
	NullStr sql.NullString
	NullInt sql.NullInt64
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name        string
		model       any
		wantTable   string
		wantColumns []string
		wantIndexes []string
		wantErr     error
		wantAnyErr  bool
	}{
		{
			name:      "ORM Model (embedded)",
			model:     &User{},
			wantTable: "users",
			wantColumns: []string{
				`table.TimestampTz("created_at").Nullable()`,
				`table.TimestampTz("updated_at").Nullable()`,
				`table.BigIncrements("id")`,
				`table.String("name")`,
			},
			wantIndexes: []string{},
		},
		{
			name:      "Simple Product with Unique Index",
			model:     &Product{},
			wantTable: "products",
			wantColumns: []string{
				`table.BigIncrements("id")`,
				`table.String("code")`,
				`table.Double("price")`,
			},
			wantIndexes: []string{
				`table.Unique("code").Name("idx_products_code")`,
			},
		},
		{
			name:      "Composite Primary Key",
			model:     &CompositeKey{},
			wantTable: "composite_keys",
			wantColumns: []string{
				`table.UnsignedBigInteger("order_id")`,
				`table.UnsignedBigInteger("product_id")`,
				`table.String("note")`,
			},
			wantIndexes: []string{
				`table.Primary("order_id", "product_id")`,
			},
		},
		{
			name:      "Ignored Field",
			model:     &IgnoredField{},
			wantTable: "ignored_fields",
			wantColumns: []string{
				`table.BigIncrements("id")`,
			},
		},
		{
			name:      "Custom Types (Enum, JSON)",
			model:     &CustomTypes{},
			wantTable: "custom_types",
			wantColumns: []string{
				`table.BigIncrements("id")`,
				`table.Enum("status", []any{"on", "off"})`,
				`table.Json("settings")`,
			},
		},
		{
			name:      "Pointer Fields (Nullable)",
			model:     &PtrFields{},
			wantTable: "ptr_fields",
			wantColumns: []string{
				`table.BigIncrements("id")`,
				`table.String("name").Nullable()`,
				`table.BigInteger("age").Nullable()`,
			},
		},
		{
			name:      "SQL Null Types",
			model:     &NullableFields{},
			wantTable: "nullable_fields",
			wantColumns: []string{
				`table.BigIncrements("id")`,
				`table.String("null_str").Nullable()`,
				`table.BigInteger("null_int").Nullable()`,
			},
		},
		{
			name:       "Error: Nil Model",
			model:      nil,
			wantAnyErr: true,
			wantErr:    errors.SchemaInvalidModel,
		},
		{
			name:       "Error: Non-Struct Model",
			model:      123,
			wantAnyErr: true,
			wantErr:    errors.SchemaInvalidModel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotTable, gotLines, gotErr := Generate(tc.model)

			if tc.wantAnyErr {
				assert.Error(t, gotErr)
				return
			}
			if tc.wantErr != nil {
				assert.ErrorIs(t, gotErr, tc.wantErr)
				return
			}

			assert.NoError(t, gotErr)
			assert.Equal(t, tc.wantTable, gotTable)

			gotColumns, gotIndexes := splitLines(gotLines)
			assert.Equal(t, tc.wantColumns, gotColumns, "Columns mismatch")
			assert.ElementsMatch(t, tc.wantIndexes, gotIndexes, "Indexes mismatch")
		})
	}
}

func splitLines(lines []string) (columns, indexes []string) {
	seenSeparator := false
	for _, line := range lines {
		if line == "" {
			seenSeparator = true
			continue
		}
		if seenSeparator {
			indexes = append(indexes, line)
		} else {
			columns = append(columns, line)
		}
	}
	return
}

func TestWriteValue(t *testing.T) {
	testCases := []struct {
		name  string
		value any
		want  string
	}{
		{"String", "hello", `"hello"`},
		{"Int", 123, `123`},
		{"Int64", int64(999), `999`},
		{"Float64", 12.34, `12.34`},
		{"Bool True", true, `true`},
		{"Bool False", false, `false`},
		{"Nil", nil, `nil`},
		{"SliceAny", []any{"a", int64(1)}, `[]any{"a", 1}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := &atom{Builder: strings.Builder{}}
			b.WriteValue(tc.value)
			assert.Equal(t, tc.want, b.String())
		})
	}
}

func TestIntMethod(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		unsigned bool
		want     string
	}{
		{"TinyInt signed", 8, false, "TinyInteger"},
		{"TinyInt unsigned", 8, true, "UnsignedTinyInteger"},
		{"SmallInt signed", 16, false, "SmallInteger"},
		{"SmallInt unsigned", 16, true, "UnsignedSmallInteger"},
		{"Int signed", 32, false, "Integer"},
		{"Int unsigned", 32, true, "UnsignedInteger"},
		{"BigInt signed", 64, false, "BigInteger"},
		{"BigInt unsigned", 64, true, "UnsignedBigInteger"},
		{"Size 0 falls into TinyInt", 0, false, "TinyInteger"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := intMethod(tc.size, tc.unsigned)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestParseEnum(t *testing.T) {
	tests := []struct {
		name string
		def  string
		want []any
	}{
		{"Simple strings", "enum('on','off')", []any{"on", "off"}},
		{"With spaces", "enum('active', 'inactive', 'pending')", []any{"active", "inactive", "pending"}},
		{"Integers", "enum(1, 2, 3)", []any{int64(1), int64(2), int64(3)}},
		{"Mixed types", "enum('a', 1, 2.5)", []any{"a", int64(1), 2.5}},
		{"Empty", "enum()", nil},
		{"No parentheses", "enum", nil},
		{"Single value", "enum('only')", []any{"only"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseEnum(tc.def)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestParseVal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  any
	}{
		{"Integer", "123", int64(123)},
		{"Negative integer", "-456", int64(-456)},
		{"Float", "12.34", 12.34},
		{"Negative float", "-98.76", -98.76},
		{"String", "hello", "hello"},
		{"String with spaces", "  hello  ", "hello"},
		{"Zero", "0", int64(0)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseVal(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Single quotes", "'hello'", "hello"},
		{"Double quotes", `"world"`, "world"},
		{"No quotes", "plain", "plain"},
		{"Mismatched quotes", "'mixed\"", "'mixed\""},
		{"Empty string", "", ""},
		{"Single char", "a", "a"},
		{"Only quotes", "''", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := trimQuotes(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestParseTypeSize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"varchar(255)", "varchar(255)", 255},
		{"char(50)", "char(50)", 50},
		{"decimal(10,2)", "decimal(10,2)", 10},
		{"No size", "varchar", 0},
		{"Empty parens", "varchar()", 0},
		{"Invalid number", "varchar(abc)", 0},
		{"Spaces", "varchar( 100 )", 100},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseTypeSize(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsSQLNullType(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"sql.NullString", sql.NullString{}, true},
		{"sql.NullInt64", sql.NullInt64{}, true},
		{"sql.NullBool", sql.NullBool{}, true},
		{"sql.NullFloat64", sql.NullFloat64{}, true},
		{"Regular string", "", false},
		{"Pointer", (*string)(nil), false},
		{"Struct", struct{ Name string }{}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isSQLNullType(reflect.TypeOf(tc.value))
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestWriteMethod(t *testing.T) {
	tests := []struct {
		name string
		call func(b *atom)
		want string
	}{
		{
			name: "Method with no args",
			call: func(b *atom) { b.WriteMethod("Nullable") },
			want: "Nullable()",
		},
		{
			name: "Method with string arg",
			call: func(b *atom) { b.WriteMethod("String", "name") },
			want: `String("name")`,
		},
		{
			name: "Method with multiple args",
			call: func(b *atom) { b.WriteMethod("Decimal", "price", 10, 2) },
			want: `Decimal("price", 10, 2)`,
		},
		{
			name: "Chained methods",
			call: func(b *atom) {
				b.WriteString("table.")
				b.WriteMethod("String", "name")
				b.WriteMethod("Nullable")
			},
			want: `table.String("name").Nullable()`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := &atom{Builder: strings.Builder{}}
			tc.call(b)
			assert.Equal(t, tc.want, b.String())
		})
	}
}
