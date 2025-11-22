package migration

import (
	"database/sql"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/structmeta"
)

type BasicModel struct {
	ID         uint `gorm:"primaryKey;autoIncrement"`
	Name       string
	Email      *string `gorm:"unique"`
	unexported string
}

type TypeMappingModel struct {
	Int    int
	Int32  int32
	Int64  int64
	Uint   uint `gorm:"unsigned"`
	Uint32 uint32
	Uint64 uint64
	Bool   bool
	Time   time.Time
	//CarbonTime carbon.DateTime
	//Carbon     carbon.Carbon
	Float64    float64
	Float32    float32
	Int16      int16
	Int8       int8
	Uint16     uint16
	Uint8      uint8
	Bytes      []byte
	Uint8Slice []uint8
	DeletedAt  gorm.DeletedAt
	MapData    map[string]string `gorm:"type:json"`
	SliceData  []int             `gorm:"type:json"`
	NullStr    sql.NullString
	NullInt    sql.NullInt64
	NullBool   sql.NullBool
	NullFloat  sql.NullFloat64
	NullTime   sql.NullTime
}

type TagParsingModel struct {
	DefaultID        uint   `gorm:"primaryKey;autoIncrement"`
	CustomID         string `gorm:"primaryKey;size:36;type:char(36)"`
	SimpleString     string
	WithSize         string `gorm:"size:100"`
	WithColName      string `gorm:"column:custom_col"`
	WithDefault      string `gorm:"default:'hello'"`
	WithComment      string `gorm:"comment:'this is a comment'"`
	NotNullable      string `gorm:"not null"`
	ExplicitNull     string `gorm:"nullable"`
	PointerIsNull    *int
	UnsignedInt      uint
	SignedInt        int     `gorm:"unsigned"`
	Decimal          float64 `gorm:"type:decimal(12,4)"`
	Enum             string  `gorm:"type:enum('a','b',1);default:'a'"`
	DBType           string  `gorm:"type:varchar(50)"`
	IgnoreMe         string  `gorm:"-"`
	UniqueField      string  `gorm:"unique"`
	IndexField       int     `gorm:"index"`
	NamedIndex       int     `gorm:"index:my_idx"`
	UniqueIndex      string  `gorm:"uniqueIndex"`
	NamedUniqIdx     string  `gorm:"uniqueIndex:my_uidx"`
	Composite1       string  `gorm:"index:comp_idx;uniqueIndex:comp_uidx"`
	Composite2       int     `gorm:"index:comp_idx"`
	CompUnique2      int     `gorm:"uniqueIndex:comp_uidx"`
	IndexWithOptions string  `gorm:"index:idx_opts,unique"`
}

type RelationModel struct {
	ID          uint `gorm:"primaryKey;autoIncrement"`
	Profile     BasicModel
	Profiles    []BasicModel
	ProfilePtr  *BasicModel
	ProfilePtrs []*BasicModel
}

type TableNameModel struct {
	ID uint `gorm:"primaryKey;autoIncrement"`
}

func (t *TableNameModel) TableName() string {
	return "custom_table_name"
}

func findFieldMeta(meta structmeta.StructMetadata, fieldName string) *structmeta.FieldMetadata {
	for i := range meta.Fields {
		if meta.Fields[i].Name == fieldName {
			return &meta.Fields[i]
		}
	}
	return nil
}

func TestGenerate(t *testing.T) {
	testCases := []struct {
		name        string
		model       any
		wantTable   string
		wantColumns []string
		wantIndexes []string
		wantErr     error
	}{
		{
			name:      "Basic Model",
			model:     &BasicModel{},
			wantTable: "basic_models",
			wantColumns: []string{
				`table.Increments("id")`,
				`table.Text("name")`,
				`table.Text("email").Nullable()`,
			},
			wantIndexes: []string{
				`table.Unique("email")`,
			},
			wantErr: nil,
		},
		{
			name:      "Type Mapping Model",
			model:     &TypeMappingModel{},
			wantTable: "type_mapping_models",
			wantColumns: []string{
				`table.Integer("int")`,
				`table.Integer("int32")`,
				`table.BigInteger("int64")`,
				`table.UnsignedInteger("uint").Unsigned()`,
				`table.UnsignedInteger("uint32")`,
				`table.UnsignedBigInteger("uint64")`,
				`table.Boolean("bool")`,
				`table.DateTimeTz("time")`,
				//`table.DateTimeTz("carbon_time")`,
				//`table.DateTimeTz("carbon")`,
				`table.Double("float64")`,
				`table.Float("float32")`,
				`table.SmallInteger("int16")`,
				`table.TinyInteger("int8")`,
				`table.UnsignedSmallInteger("uint16")`,
				`table.UnsignedTinyInteger("uint8")`,
				`table.Binary("bytes")`,
				`table.Binary("uint8_slice")`,
				//`table.SoftDeletesTz("deleted_at").Nullable()`,
				`table.Json("map_data")`,
				`table.Json("slice_data")`,
				`table.Text("null_str").Nullable()`,
				`table.BigInteger("null_int").Nullable()`,
				`table.Boolean("null_bool").Nullable()`,
				`table.Double("null_float").Nullable()`,
				`table.DateTimeTz("null_time").Nullable()`,
			},
			wantIndexes: nil,
			wantErr:     nil,
		},
		{
			name:      "Tag Parsing Model",
			model:     &TagParsingModel{},
			wantTable: "tag_parsing_models",
			wantColumns: []string{
				`table.Increments("default_id")`,
				`table.String("custom_id", 36)`,
				`table.Text("simple_string")`,
				`table.String("with_size", 100)`,
				`table.Text("custom_col")`,
				`table.Text("with_default").Default("hello")`,
				`table.Text("with_comment").Comment("this is a comment")`,
				`table.Text("not_nullable")`,
				`table.Text("explicit_null").Nullable()`,
				`table.Integer("pointer_is_null").Nullable()`,
				`table.UnsignedInteger("unsigned_int")`,
				`table.Integer("signed_int").Unsigned()`,
				`table.Decimal("decimal").Places(4).Total(12)`,
				`table.Enum("enum", []any{"a", "b", 1}).Default("a")`,
				`table.String("db_type", 50)`,
				`table.Text("unique_field")`,
				`table.Integer("index_field")`,
				`table.Integer("named_index")`,
				`table.Text("unique_index")`,
				`table.Text("named_uniq_idx")`,
				`table.Text("composite1")`,
				`table.Integer("composite2")`,
				`table.Integer("comp_unique2")`,
				`table.Text("index_with_options")`,
			},
			wantIndexes: []string{
				`table.Primary("custom_id")`,
				`table.Index("composite1", "composite2")`,
				`table.Index("index_field")`,
				`table.Index("named_index")`,
				`table.Unique("comp_unique2", "composite1")`,
				`table.Unique("index_with_options")`,
				`table.Unique("named_uniq_idx")`,
				`table.Unique("unique_field")`,
				`table.Unique("unique_index")`,
			},
			wantErr: nil,
		},
		{
			name:      "Relation Model",
			model:     &RelationModel{},
			wantTable: "relation_models",
			wantColumns: []string{
				`table.Increments("id")`,
			},
			wantIndexes: nil,
			wantErr:     nil,
		},
		{
			name:      "Custom Table Name Model",
			model:     &TableNameModel{},
			wantTable: "custom_table_name",
			wantColumns: []string{
				`table.Increments("id")`,
			},
			wantIndexes: nil,
			wantErr:     nil,
		},
		{
			name:        "Error - Nil model",
			model:       nil,
			wantTable:   "",
			wantColumns: nil,
			wantIndexes: nil,
			wantErr:     errors.SchemaInvalidModel,
		},
		{
			name:        "Error - Non-struct model",
			model:       123,
			wantTable:   "",
			wantColumns: nil,
			wantIndexes: nil,
			wantErr:     errors.SchemaInvalidModel,
		},
		{
			name:        "Error - Empty struct model",
			model:       struct{}{},
			wantTable:   "",
			wantColumns: nil,
			wantIndexes: nil,
			wantErr:     errors.SchemaInvalidModel,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotTable, gotLines, gotErr := Generate(tc.model)

			assert.Equal(t, tc.wantTable, gotTable)
			assert.ErrorIs(t, gotErr, tc.wantErr)

			if tc.wantErr != nil {
				return
			}

			var gotColumns, gotIndexes []string
			separatorFound := false
			for _, line := range gotLines {
				if line == "" {
					separatorFound = true
					continue
				}
				if separatorFound {
					gotIndexes = append(gotIndexes, line)
				} else {
					gotColumns = append(gotColumns, line)
				}
			}

			assert.Equal(t, tc.wantColumns, gotColumns, "Generated columns mismatch or are out of order")

			if len(tc.wantIndexes) > 0 || len(gotIndexes) > 0 {
				assert.ElementsMatch(t, tc.wantIndexes, gotIndexes, "Generated indexes mismatch")
			}
		})
	}
}

func TestParseTag(t *testing.T) {
	type TagTestStruct struct {
		FieldA string  `gorm:"-"`
		FieldB string  `gorm:"column:my_col"`
		FieldC string  `gorm:"type:varchar(100);unsigned"`
		FieldD string  `gorm:"size:255"`
		FieldE string  `gorm:"default:'abc'"`
		FieldF string  `gorm:"comment:'a comment'"`
		FieldG string  `gorm:"unique"`
		FieldH int     `gorm:"index"`
		FieldI int     `gorm:"index:my_idx"`
		FieldJ string  `gorm:"index:my_idx,unique"`
		FieldK string  `gorm:"uniqueIndex"`
		FieldL string  `gorm:"uniqueIndex:my_uidx"`
		FieldM string  `gorm:"not null"`
		FieldN string  `gorm:"nullable"`
		FieldO int     `gorm:"unsigned"`
		FieldP uint    `gorm:"primaryKey"`
		FieldQ int     `gorm:"autoIncrement"`
		FieldR float64 `gorm:"precision:10;scale:2"`
		FieldS float64 `gorm:"type:decimal(8, 3)"`
		FieldT string  `gorm:"type:enum('x','y')"`
		FieldU string  `gorm:"enum:1,2,'c'"`
		FieldV string  `gorm:"serializer:json"`
		FieldW string  `gorm:"column:the_col;size:50;index;not null"`
	}
	meta := structmeta.WalkStruct(TagTestStruct{})

	testCases := []struct {
		fieldName string
		want      *tag
	}{
		{"FieldA", &tag{ignore: true}},
		{"FieldB", &tag{column: "my_col"}},
		{"FieldC", &tag{dbType: "varchar(100)", unsigned: true}},
		{"FieldD", &tag{size: 255}},
		{"FieldE", &tag{defaultVal: "abc"}},
		{"FieldF", &tag{comment: "a comment"}},
		{"FieldG", &tag{unique: true}},
		{"FieldH", &tag{index: "idx_field_h"}},
		{"FieldI", &tag{index: "my_idx"}},
		{"FieldJ", &tag{index: "my_idx", indexUnique: true}},
		{"FieldK", &tag{uniqueIndex: "uidx_field_k"}},
		{"FieldL", &tag{uniqueIndex: "my_uidx"}},
		{"FieldM", &tag{notNull: true}},
		{"FieldN", &tag{nullable: true}},
		{"FieldO", &tag{unsigned: true}},
		{"FieldP", &tag{primaryKey: true}},
		{"FieldQ", &tag{autoIncrement: true}},
		{"FieldR", &tag{precision: 10, scale: 2}},
		{"FieldS", &tag{dbType: "decimal(8, 3)", precision: 8, scale: 3}},
		{"FieldT", &tag{dbType: "enum('x','y')", enumValues: []any{"x", "y"}}},
		{"FieldU", &tag{enumValues: []any{int64(1), int64(2), "c"}}},
		{"FieldV", &tag{dbType: "json"}},
		{"FieldW", &tag{column: "the_col", size: 50, index: "idx_field_w", notNull: true}},
	}

	for _, tc := range testCases {
		t.Run(tc.fieldName, func(t *testing.T) {
			field := findFieldMeta(meta, tc.fieldName)
			assert.NotNil(t, field, "Field %s not found", tc.fieldName)
			if field == nil {
				return
			}
			got := parseTag(field)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestMapType(t *testing.T) {
	meta := structmeta.WalkStruct(TypeMappingModel{})

	testCases := []struct {
		fieldName  string
		tag        *tag
		wantMethod string
		wantArgs   []any
	}{
		{"Int", &tag{}, "Integer", nil},
		{"Uint", &tag{}, "UnsignedInteger", nil},
		{"Bool", &tag{}, "Boolean", nil},
		{"Time", &tag{}, "DateTimeTz", nil},
		{"Bytes", &tag{}, "Binary", nil},
		{"DeletedAt", &tag{}, "SoftDeletesTz", nil},
		{"MapData", &tag{dbType: "json"}, "Json", nil},
		{"SliceData", &tag{dbType: "json"}, "Json", nil},
		{"NullStr", &tag{}, "Text", nil},
		{"NullInt", &tag{}, "BigInteger", nil},
		{"Int", &tag{enumValues: []any{"a"}}, "Enum", nil},
		{"Int", &tag{dbType: "varchar(10)"}, "String", []any{10}},
		{"Int", &tag{dbType: "FLOAT"}, "Float", nil},
		{"Uint8Slice", &tag{}, "Binary", nil},
		{"MapData", &tag{}, "Json", nil},
		{"SliceData", &tag{}, "Json", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.fieldName, func(t *testing.T) {
			field := findFieldMeta(meta, tc.fieldName)
			assert.NotNil(t, field, "Field %s not found", tc.fieldName)
			if field == nil {
				return
			}

			tagToUse := tc.tag
			if tagToUse == nil {
				tagToUse = parseTag(field)
			}

			gotMethod, gotArgs := mapType(field, tagToUse)
			assert.Equal(t, tc.wantMethod, gotMethod)
			assert.Equal(t, tc.wantArgs, gotArgs)
		})
	}
}

func TestMapDBType(t *testing.T) {
	testCases := []struct {
		name       string
		dbType     string
		size       int
		wantMethod string
		wantArgs   []any
	}{
		{"VarcharWithSize", "varchar(100)", 0, "String", []any{100}},
		{"VarcharNoSize", "varchar", 0, "String", nil},
		{"Char", "char(10)", 0, "String", []any{10}},
		{"Text", "TEXT", 0, "Text", nil},
		{"Int", "INT", 0, "Integer", nil},
		{"IntegerUnsigned", "integer unsigned", 0, "Integer", nil},
		{"BigInt", "bigint", 0, "BigInteger", nil},
		{"TinyInt(1)", "tinyint(1)", 0, "Boolean", nil},
		{"TinyInt", "tinyint", 0, "TinyInteger", nil},
		{"Decimal", "decimal(8,2)", 0, "Decimal", nil},
		{"Json", "json", 0, "Json", nil},
		{"Timestamp", "timestamp", 0, "DateTimeTz", nil},
		{"Date", "date", 0, "Date", nil},
		{"Time", "time", 0, "Time", nil},
		{"Blob", "blob", 0, "Binary", nil},
		{"BinaryWithSize", "binary(16)", 0, "Binary", nil},
		{"Uuid", "uuid", 0, "Uuid", nil},
		{"Unknown", "weirdtype", 0, "Column", []any{`weirdtype`}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotMethod, gotArgs := mapDBType(tc.dbType, tc.size)
			assert.Equal(t, tc.wantMethod, gotMethod)
			assert.Equal(t, tc.wantArgs, gotArgs)
		})
	}
}

func TestIsNullable(t *testing.T) {
	type TestStruct struct {
		NotNullTag    string `gorm:"not null"`
		NullableTag   string `gorm:"nullable"`
		PointerType   *string
		ValueType     string
		SQLNullString sql.NullString
		GormDeletedAt gorm.DeletedAt
	}
	meta := structmeta.WalkStruct(TestStruct{})

	testCases := []struct {
		fieldName   string
		tagOverride *tag
		want        bool
	}{
		{"NotNullTag", nil, false},
		{"NullableTag", nil, true},
		{"PointerType", nil, true},
		{"ValueType", nil, false},
		{"SQLNullString", nil, true},
		{"GormDeletedAt", nil, true},
		{"ValueType", &tag{nullable: true}, true},
		{"PointerType", &tag{notNull: true}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.fieldName, func(t *testing.T) {
			field := findFieldMeta(meta, tc.fieldName)
			assert.NotNil(t, field, "Field %s not found", tc.fieldName)
			if field == nil {
				return
			}

			tagToUse := tc.tagOverride
			if tagToUse == nil {
				tagToUse = parseTag(field)
			}

			got := isNullable(field, tagToUse)
			assert.Equal(t, tc.want, got)
		})
	}
}

type AnonEmbed struct {
	InnerField string
}

type TestStruct struct {
	Exported   string
	unexported string
	Relation   BasicModel
	TimeField  time.Time
	AnonEmbed  `gorm:"embedded"`
}

func TestShouldInclude(t *testing.T) {
	meta := structmeta.WalkStruct(TestStruct{})

	testCases := []struct {
		fieldName string
		want      bool
	}{
		{"Exported", true},
		{"Relation", false},
		{"TimeField", true},
		{"AnonEmbed", false},
		{"InnerField", true},
	}

	for _, tc := range testCases {
		t.Run(tc.fieldName, func(t *testing.T) {
			fieldPtr := findFieldMeta(meta, tc.fieldName)
			if !assert.NotNil(t, fieldPtr, "Field %s not found in parsed metadata", tc.fieldName) {
				t.FailNow()
			}
			got := shouldInclude(fieldPtr)
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("unexported", func(t *testing.T) {
		unexportedField := structmeta.FieldMetadata{Name: "unexported"}
		assert.False(t, shouldInclude(&unexportedField))
	})
}

func TestTableName(t *testing.T) {
	testCases := []struct {
		name  string
		model any
		want  string
	}{
		{"Default Plural Snake", &RelationModel{}, "relation_models"},
		{"Custom TableName Method", &TableNameModel{}, "custom_table_name"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			meta := structmeta.WalkStruct(tc.model)
			got := tableName(meta)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestCollectFieldIndexes(t *testing.T) {
	testCases := []struct {
		name    string
		colName string
		tag     *tag
		want    map[string]index
	}{
		{"No Indexes", "col", &tag{}, map[string]index{}},
		{"Simple Unique", "col", &tag{unique: true}, map[string]index{
			"col_unique": {columns: []string{"col"}, unique: true},
		}},
		{"Simple Index", "col", &tag{index: "idx_col"}, map[string]index{
			"idx_col": {columns: []string{"col"}, unique: false},
		}},
		{"Named Unique Index", "col", &tag{uniqueIndex: "uidx_col"}, map[string]index{
			"uidx_col": {columns: []string{"col"}, unique: true},
		}},
		{"Index with Unique Option", "col", &tag{index: "idx_col", indexUnique: true}, map[string]index{
			"idx_col": {columns: []string{"col"}, unique: true},
		}},
		{"Multiple (Unique and Index)", "col", &tag{unique: true, index: "idx_col"}, map[string]index{
			"col_unique": {columns: []string{"col"}, unique: true},
			"idx_col":    {columns: []string{"col"}, unique: false},
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := collectFieldIndexes(tc.tag, tc.colName)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestBuildColumn(t *testing.T) {
	type TestStruct struct {
		ID        uint   `gorm:"primaryKey;autoIncrement"`
		UUID      string `gorm:"primaryKey;size:36;type:char(36)"`
		Name      string
		Email     *string `gorm:"size:100;default:'a@b.c';comment:'Email Addr'"`
		Age       int     `gorm:"unsigned"`
		Price     float64 `gorm:"type:decimal(10,2)"`
		EnumField string  `gorm:"type:enum('x','y')"`
	}
	meta := structmeta.WalkStruct(TestStruct{})

	testCases := []struct {
		name  string
		field *structmeta.FieldMetadata
		tag   *tag
		want  *column
	}{
		{
			name:  "AutoIncrement PK (uint)",
			field: findFieldMeta(meta, "ID"),
			tag:   parseTag(findFieldMeta(meta, "ID")),
			want:  &column{name: "id", method: "Increments"},
		},
		{
			name:  "String PK (char)",
			field: findFieldMeta(meta, "UUID"),
			tag:   parseTag(findFieldMeta(meta, "UUID")),
			want: &column{
				name:      "uuid",
				method:    "String",
				args:      []any{36},
				modifiers: []modifier{},
			},
		},
		{
			name:  "Simple String (Text)",
			field: findFieldMeta(meta, "Name"),
			tag:   parseTag(findFieldMeta(meta, "Name")),
			want:  &column{name: "name", method: "Text", modifiers: []modifier{}},
		},
		{
			name:  "Pointer String with Modifiers",
			field: findFieldMeta(meta, "Email"),
			tag:   parseTag(findFieldMeta(meta, "Email")),
			want: &column{
				name:   "email",
				method: "String",
				args:   []any{100},
				modifiers: []modifier{
					{name: "Comment", arg: "Email Addr"},
					{name: "Default", arg: "a@b.c"},
					{name: "Nullable"},
				},
			},
		},
		{
			name:  "Unsigned Int",
			field: findFieldMeta(meta, "Age"),
			tag:   parseTag(findFieldMeta(meta, "Age")),
			want: &column{
				name:      "age",
				method:    "Integer",
				modifiers: []modifier{{name: "Unsigned"}},
			},
		},
		{
			name:  "Decimal",
			field: findFieldMeta(meta, "Price"),
			tag:   parseTag(findFieldMeta(meta, "Price")),
			want: &column{
				name:   "price",
				method: "Decimal",
				modifiers: []modifier{
					{name: "Places", arg: "2"},
					{name: "Total", arg: "10"},
				},
			},
		},
		{
			name:  "Enum",
			field: findFieldMeta(meta, "EnumField"),
			tag:   parseTag(findFieldMeta(meta, "EnumField")),
			want: &column{
				name:      "enum_field",
				method:    "Enum",
				enum:      []any{"x", "y"},
				modifiers: []modifier{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc.field, "Field metadata is nil for %s", tc.name)
			if tc.field == nil {
				return
			}

			got := buildColumn(tc.field, tc.tag)
			if got != nil {
				sort.SliceStable(got.modifiers, func(i, j int) bool {
					return got.modifiers[i].name < got.modifiers[j].name
				})
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestRenderColumn(t *testing.T) {
	testCases := []struct {
		name string
		col  *column
		want string
	}{
		{
			name: "Simple Text",
			col:  &column{name: "description", method: "Text"},
			want: `table.Text("description")`,
		},
		{
			name: "String with Size",
			col:  &column{name: "code", method: "String", args: []any{50}},
			want: `table.String("code", 50)`,
		},
		{
			name: "Integer Unsigned Nullable",
			col: &column{name: "count", method: "Integer", modifiers: []modifier{
				{name: "Nullable"}, {name: "Unsigned"},
			}},
			want: `table.Integer("count").Nullable().Unsigned()`,
		},
		{
			name: "Decimal with Default and Comment",
			col: &column{name: "amount", method: "Decimal", modifiers: []modifier{
				{name: "Comment", arg: "The amount"}, {name: "Default", arg: "0.00"}, // Sorted
				{name: "Places", arg: "3"}, {name: "Total", arg: "12"},
			}},
			want: `table.Decimal("amount").Comment("The amount").Default("0.00").Places(3).Total(12)`,
		},
		{
			name: "Enum",
			col:  &column{name: "status", method: "Enum", enum: []any{"pending", "done", int64(1)}},
			want: `table.Enum("status", []any{"pending", "done", 1})`,
		},
		{
			name: "Increments",
			col:  &column{name: "id", method: "BigIncrements"},
			want: `table.BigIncrements("id")`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.col != nil {
				sort.SliceStable(tc.col.modifiers, func(i, j int) bool {
					return tc.col.modifiers[i].name < tc.col.modifiers[j].name
				})
			}
			got := renderColumn(tc.col)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestRenderIndex(t *testing.T) {
	testCases := []struct {
		name string
		idx  index
		want string
	}{
		{
			name: "Primary Single",
			idx:  index{columns: []string{"id"}, isPrimary: true},
			want: `table.Primary("id")`,
		},
		{
			name: "Primary Composite (Unsorted Input)",
			idx:  index{columns: []string{"col_b", "col_a"}, isPrimary: true},
			want: `table.Primary("col_a", "col_b")`,
		},
		{
			name: "Unique Single",
			idx:  index{columns: []string{"email"}, unique: true},
			want: `table.Unique("email")`,
		},
		{
			name: "Unique Composite (Unsorted Input)",
			idx:  index{columns: []string{"org_id", "email"}, unique: true},
			want: `table.Unique("email", "org_id")`,
		},
		{
			name: "Index Single",
			idx:  index{columns: []string{"created_at"}},
			want: `table.Index("created_at")`,
		},
		{
			name: "Index Composite (Unsorted Input)",
			idx:  index{columns: []string{"zip_code", "user_id"}},
			want: `table.Index("user_id", "zip_code")`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := renderIndex(tc.idx)
			assert.Equal(t, tc.want, got)
		})
	}
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
		{"Uint", uint(1), `1`},
		{"Uint64", uint64(2), `2`},
		{"Float64", 12.34, `12.34`},
		{"Float32", float32(5.6), `5.6`},
		{"Bool True", true, `true`},
		{"Bool False", false, `false`},
		{"Nil", nil, `nil`},
		{"String Ptr", func() *string { s := "ptr"; return &s }(), `"ptr"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var b strings.Builder
			writeValue(&b, tc.value)
			assert.Equal(t, tc.want, b.String())
		})
	}
}
