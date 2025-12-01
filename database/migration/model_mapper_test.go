package migration

import (
	"database/sql"
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
			b.writeValue(tc.value)
			assert.Equal(t, tc.want, b.String())
		})
	}
}
