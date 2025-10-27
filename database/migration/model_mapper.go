package migration

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/str"
	"github.com/goravel/framework/support/structmeta"
)

const (
	avgColumnLine = 64
	avgIndexLine  = 40
	avgColNameLen = 10
)

var goTypeToMethod = map[string]string{
	"uint":                "UnsignedInteger",
	"int":                 "Integer",
	"int32":               "Integer",
	"int64":               "BigInteger",
	"uint64":              "UnsignedBigInteger",
	"bool":                "Boolean",
	"time.Time":           "DateTimeTz",
	"carbon.DateTime":     "DateTimeTz",
	"carbon.Carbon":       "DateTimeTz",
	"float64":             "Double",
	"float32":             "Float",
	"int16":               "SmallInteger",
	"int8":                "TinyInteger",
	"uint32":              "UnsignedInteger",
	"uint16":              "UnsignedSmallInteger",
	"uint8":               "UnsignedTinyInteger",
	"[]byte":              "Binary",
	"[]uint8":             "Binary",
	"gorm.DeletedAt":      "SoftDeletesTz",
	"datatypes.JSON":      "Json",
	"datatypes.JSONMap":   "Json",
	"datatypes.JSONArray": "Json",
	"datatypes.Date":      "Date",
	"datatypes.Time":      "Time",
	"sql.NullString":      "Text",
	"sql.NullInt64":       "BigInteger",
	"sql.NullInt32":       "Integer",
	"sql.NullInt16":       "SmallInteger",
	"sql.NullBool":        "Boolean",
	"sql.NullFloat64":     "Double",
	"sql.NullTime":        "DateTimeTz",
}

var goNullableTypes = map[string]bool{
	"sql.NullString":  true,
	"sql.NullInt64":   true,
	"sql.NullInt32":   true,
	"sql.NullInt16":   true,
	"sql.NullBool":    true,
	"sql.NullFloat64": true,
	"sql.NullTime":    true,
}

var nonRelationTypes = map[string]bool{
	"time.Time":           true,
	"gorm.DeletedAt":      true,
	"carbon.DateTime":     true,
	"carbon.Carbon":       true,
	"datatypes.JSON":      true,
	"datatypes.JSONMap":   true,
	"datatypes.JSONArray": true,
	"datatypes.Date":      true,
	"datatypes.Time":      true,
	"sql.NullString":      true,
	"sql.NullInt64":       true,
	"sql.NullInt32":       true,
	"sql.NullInt16":       true,
	"sql.NullBool":        true,
	"sql.NullFloat64":     true,
	"sql.NullTime":        true,
}

var dbTypeToMethod = map[string]string{
	"varchar":     "String",
	"char":        "String",
	"varbinary":   "String",
	"binary":      "Binary",
	"decimal":     "Decimal",
	"enum":        "Enum",
	"text":        "Text",
	"tinytext":    "Text",
	"mediumtext":  "Text",
	"longtext":    "Text",
	"clob":        "Text",
	"uuid":        "Uuid",
	"ulid":        "Ulid",
	"bigint":      "BigInteger",
	"tinyint(1)":  "Boolean",
	"tinyint":     "TinyInteger",
	"smallint":    "SmallInteger",
	"int":         "Integer",
	"integer":     "Integer",
	"json":        "Json",
	"jsonb":       "Json",
	"bool":        "Boolean",
	"boolean":     "Boolean",
	"timestamp":   "DateTimeTz",
	"timestamptz": "DateTimeTz",
	"datetime":    "DateTimeTz",
	"date":        "Date", // Changed from DateTimeTz to Date
	"time":        "Time", // Changed from DateTimeTz to Time
	"blob":        "Binary",
	"tinyblob":    "Binary",
	"mediumblob":  "Binary",
	"longblob":    "Binary",
	"double":      "Double",
	"real":        "Double",
	"float":       "Float",
}

var methodToIncrements = map[string]string{
	"UnsignedBigInteger":   "BigIncrements",
	"BigInteger":           "BigIncrements",
	"Integer":              "Increments",
	"UnsignedInteger":      "Increments",
	"SmallInteger":         "SmallIncrements",
	"UnsignedSmallInteger": "SmallIncrements",
	"TinyInteger":          "TinyIncrements",
	"UnsignedTinyInteger":  "TinyIncrements",
}

func Generate(model any) (tableName string, lines []string, err error) {
	meta := structmeta.WalkStruct(model)
	if meta.Name == "" {
		return "", nil, errors.SchemaInvalidModel
	}

	s, err := buildSchema(meta)
	if err != nil {
		return "", nil, err
	}

	return s.table, renderLines(s), nil
}

type tableSchema struct {
	table   string
	columns []column
	indexes map[string]index
}

type column struct {
	name      string
	method    string
	args      []any
	enum      []any
	modifiers []modifier
}

type index struct {
	columns   []string
	unique    bool
	isPrimary bool
}

type modifier struct {
	name string
	arg  string
}

func buildSchema(meta structmeta.StructMetadata) (*tableSchema, error) {
	s := &tableSchema{
		table:   tableName(meta),
		indexes: make(map[string]index),
		columns: make([]column, 0, len(meta.Fields)),
	}

	var primaryKeyColumns []string
	processedColumns := make(map[string]bool)

	for _, field := range meta.Fields {
		t := parseTag(&field)
		if t.ignore {
			continue
		}

		if !shouldInclude(&field) {
			continue
		}

		col := buildColumn(&field, t)
		if col == nil {
			continue
		}

		if processedColumns[col.name] {
			continue
		}
		processedColumns[col.name] = true

		s.columns = append(s.columns, *col)

		if t.primaryKey && !t.autoIncrement {
			primaryKeyColumns = append(primaryKeyColumns, col.name)
		}

		mergeIndexes(s.indexes, collectFieldIndexes(t, col.name))
	}

	if len(primaryKeyColumns) > 0 {
		s.indexes["_primary"] = index{
			columns:   primaryKeyColumns,
			isPrimary: true,
		}
	}

	if len(s.columns) == 0 {
		return nil, errors.SchemaInvalidModel
	}

	return s, nil
}

func mergeIndexes(main, new map[string]index) {
	for name, idx := range new {
		if existing, ok := main[name]; ok {
			existing.columns = append(existing.columns, idx.columns...)
			if idx.unique {
				existing.unique = true
			}
			main[name] = existing
		} else {
			main[name] = idx
		}
	}
}

func buildColumn(field *structmeta.FieldMetadata, t *tag) *column {
	name := t.column
	if name == "" {
		name = str.Of(field.Name).Snake().String()
	}

	method, args := mapType(field, t)
	if method == "" {
		return nil
	}

	if t.primaryKey && t.autoIncrement {
		incrementsMethod, ok := methodToIncrements[method]
		if !ok {
			incrementsMethod = "BigIncrements"
		}
		return &column{name: name, method: incrementsMethod}
	}

	col := &column{
		name:   name,
		method: method,
		args:   args,
		enum:   t.enumValues,
	}

	mods := make([]modifier, 0, 2)

	if t.unsigned {
		mods = append(mods, modifier{name: "Unsigned"})
	}
	if method == "Decimal" {
		if t.precision > 0 {
			mods = append(mods, modifier{"Total", strconv.Itoa(t.precision)})
		}
		if t.scale > 0 {
			mods = append(mods, modifier{"Places", strconv.Itoa(t.scale)})
		}
	}
	if isNullable(field, t) {
		mods = append(mods, modifier{name: "Nullable"})
	}
	if t.defaultVal != "" {
		mods = append(mods, modifier{"Default", t.defaultVal})
	}
	if t.comment != "" {
		mods = append(mods, modifier{"Comment", t.comment})
	}

	col.modifiers = mods
	return col
}

func renderLines(s *tableSchema) []string {
	capacity := len(s.columns) + len(s.indexes)
	if len(s.indexes) > 0 {
		capacity++
	}
	lines := make([]string, 0, capacity)

	for i := range s.columns {
		lines = append(lines, renderColumn(&s.columns[i]))
	}

	if len(s.indexes) > 0 {
		lines = append(lines, "")
		names := make([]string, 0, len(s.indexes))
		for name := range s.indexes {
			names = append(names, name)
		}
		sort.Strings(names)

		if len(names) > 1 {
			for i, n := range names {
				if n == "_primary" {
					names = append([]string{"_primary"}, append(names[:i], names[i+1:]...)...)
					break
				}
			}
		}

		for _, name := range names {
			lines = append(lines, renderIndex(s.indexes[name]))
		}
	}

	return lines
}

func renderColumn(col *column) string {
	var b strings.Builder
	b.Grow(avgColumnLine)

	b.WriteString("table.")
	b.WriteString(col.method)
	b.WriteByte('(')
	b.WriteByte('"')
	b.WriteString(col.name)
	b.WriteByte('"')

	if col.method == "Enum" && len(col.enum) > 0 {
		b.WriteString(", []any{")
		for i, v := range col.enum {
			if i > 0 {
				b.WriteString(", ")
			}
			writeValue(&b, v)
		}
		b.WriteByte('}')
	} else {
		for _, arg := range col.args {
			b.WriteString(", ")
			writeValue(&b, arg)
		}
	}

	b.WriteByte(')')

	for _, m := range col.modifiers {
		b.WriteByte('.')
		b.WriteString(m.name)
		b.WriteByte('(')
		if m.arg != "" {
			if m.name == "Total" || m.name == "Places" {
				b.WriteString(m.arg)
			} else {
				b.WriteString(strconv.Quote(m.arg))
			}
		}
		b.WriteByte(')')
	}

	return b.String()
}

func writeValue(b *strings.Builder, v any) {
	switch val := v.(type) {
	case int:
		b.WriteString(strconv.Itoa(val))
	case string:
		b.WriteString(strconv.Quote(val))
	case int64:
		b.Write(strconv.AppendInt(nil, val, 10))
	case uint:
		b.Write(strconv.AppendUint(nil, uint64(val), 10))
	case uint64:
		b.Write(strconv.AppendUint(nil, val, 10))
	case int32:
		b.Write(strconv.AppendInt(nil, int64(val), 10))
	case uint32:
		b.Write(strconv.AppendUint(nil, uint64(val), 10))
	case float64:
		b.Write(strconv.AppendFloat(nil, val, 'g', -1, 64))
	case float32:
		b.Write(strconv.AppendFloat(nil, float64(val), 'g', -1, 32))
	case bool:
		b.WriteString(strconv.FormatBool(val))
	default:
		b.WriteString("nil")
	}
}

func renderIndex(idx index) string {
	var b strings.Builder
	b.Grow(avgIndexLine + len(idx.columns)*avgColNameLen)

	b.WriteString("table.")

	if idx.isPrimary {
		b.WriteString("Primary(")
	} else if idx.unique {
		b.WriteString("Unique(")
	} else {
		b.WriteString("Index(")
	}

	for i := range idx.columns {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteByte('"')
		b.WriteString(idx.columns[i])
		b.WriteByte('"')
	}

	b.WriteByte(')')
	return b.String()
}

type tag struct {
	column        string
	dbType        string
	defaultVal    string
	comment       string
	enumValues    []any
	size          int
	precision     int
	scale         int
	primaryKey    bool
	autoIncrement bool
	unique        bool
	index         string
	uniqueIndex   string
	notNull       bool
	nullable      bool
	unsigned      bool
	ignore        bool
	indexUnique   bool
}

func parseTag(field *structmeta.FieldMetadata) *tag {
	t := &tag{}

	if field.Tag == nil {
		return t
	}
	tagStr := field.Tag.Get("gorm")
	if tagStr == "" {
		return t
	}
	if tagStr == "-" {
		t.ignore = true
		return t
	}

	parts := strings.Split(tagStr, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		key, val := strings.ToLower(part), ""
		if idx := strings.IndexByte(part, ':'); idx != -1 {
			key = strings.ToLower(strings.TrimSpace(part[:idx]))
			val = strings.TrimSpace(part[idx+1:])
		}

		switch key {
		case "column":
			t.column = val
		case "type":
			t.dbType = val
			lowerVal := strings.ToLower(val)
			if strings.HasPrefix(lowerVal, "decimal") {
				parseDecimal(t, val)
			}
			if strings.HasPrefix(lowerVal, "enum") {
				start := strings.IndexByte(val, '(')
				end := strings.LastIndexByte(val, ')')
				if start != -1 && end != -1 && end > start {
					parseEnum(t, val[start+1:end])
				}
			}
			if strings.Contains(lowerVal, "unsigned") {
				t.unsigned = true
			}
		case "size":
			t.size, _ = strconv.Atoi(val)
		case "default":
			t.defaultVal = strings.Trim(val, `"'`)
		case "comment":
			t.comment = strings.Trim(val, `"'`)
		case "unique":
			t.unique = true
		case "index":
			name, uniq := parseIndexSpec(val, "idx_"+str.Of(field.Name).Snake().String())
			t.index = name
			t.indexUnique = uniq
		case "uniqueindex", "unique_index":
			name, _ := parseIndexSpec(val, "uidx_"+str.Of(field.Name).Snake().String())
			t.uniqueIndex = name
		case "not null":
			t.notNull = true
		case "null", "nullable":
			t.nullable = true
		case "unsigned":
			t.unsigned = true
		case "primarykey", "primary_key":
			t.primaryKey = true
		case "auto_increment", "autoincrement", "autoIncrement":
			t.autoIncrement = true
		case "precision":
			t.precision, _ = strconv.Atoi(val)
		case "scale":
			t.scale, _ = strconv.Atoi(val)
		case "enum":
			parseEnum(t, val)
		case "serializer":
			if strings.EqualFold(val, "json") {
				t.dbType = "json"
			}
		}
	}
	return t
}

func parseIndexSpec(spec, defaultName string) (name string, unique bool) {
	s := strings.TrimSpace(spec)
	if s == "" {
		return defaultName, false
	}
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.EqualFold(p, "unique") {
			unique = true
			continue
		}
		if name == "" {
			name = p
		}
	}
	if name == "" {
		name = defaultName
	}
	return name, unique
}

func parseDecimal(t *tag, typeStr string) {
	start := strings.IndexByte(typeStr, '(')
	if start == -1 {
		return
	}
	end := strings.IndexByte(typeStr[start:], ')')
	if end == -1 {
		return
	}
	end += start

	params := typeStr[start+1 : end]
	comma := strings.IndexByte(params, ',')

	if comma == -1 {
		if p, err := strconv.Atoi(strings.TrimSpace(params)); err == nil {
			t.precision = p
		}
		return
	}

	if p, err := strconv.Atoi(strings.TrimSpace(params[:comma])); err == nil {
		t.precision = p
	}
	if s, err := strconv.Atoi(strings.TrimSpace(params[comma+1:])); err == nil {
		t.scale = s
	}
}

func parseEnum(t *tag, val string) {
	parts := strings.Split(val, ",")
	t.enumValues = make([]any, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		trimmed = strings.Trim(trimmed, `"'`)

		if num, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			t.enumValues = append(t.enumValues, num)
		} else if num, err := strconv.ParseFloat(trimmed, 64); err == nil {
			t.enumValues = append(t.enumValues, num)
		} else {
			t.enumValues = append(t.enumValues, trimmed)
		}
	}
}

func mapType(field *structmeta.FieldMetadata, t *tag) (method string, args []any) {
	if len(t.enumValues) > 0 {
		return "Enum", nil
	}

	if t.dbType != "" {
		return mapDBType(t.dbType, t.size)
	}

	goType := field.Type
	if field.Kind == reflect.Ptr && len(goType) > 0 && goType[0] == '*' {
		goType = goType[1:]
	}

	if method, ok := goTypeToMethod[goType]; ok {
		return method, nil
	}

	if goType == "string" {
		if t.size > 0 {
			return "String", []any{t.size}
		}
		return "Text", nil
	}

	kind := field.Kind
	if kind == reflect.Ptr && field.ReflectType != nil {
		kind = field.ReflectType.Elem().Kind()
	}

	if kind == reflect.Map || (kind == reflect.Slice && goType != "[]byte" && goType != "[]uint8") {
		return "Json", nil
	}

	return "String", nil
}

func mapDBType(dbType string, size int) (method string, args []any) {
	s := strings.ToLower(dbType)
	method = "Column"
	args = []any{strconv.Quote(dbType)}

	if m, ok := dbTypeToMethod[s]; ok {
		method = m
		args = nil
	} else {
		// Try prefix matching for types like varchar(100), int unsigned, etc.
		for prefix, m := range dbTypeToMethod {
			if strings.HasPrefix(s, prefix) {
				method = m
				args = nil
				break
			}
		}
	}

	if method == "String" {
		if size == 0 {
			size = extractSize(s) // Extract size like varchar(255)
		}
		if size > 0 {
			args = []any{size}
		}
	}

	return method, args
}

func extractSize(typeStr string) int {
	start := strings.IndexByte(typeStr, '(')
	if start == -1 {
		return 0
	}
	end := strings.IndexByte(typeStr[start:], ')')
	if end == -1 {
		return 0
	}

	sizeStr := typeStr[start+1 : start+end]
	size, _ := strconv.Atoi(sizeStr)
	return size
}

func isNullable(field *structmeta.FieldMetadata, t *tag) bool {
	if t.nullable { // Explicit `gorm:"nullable"`
		return true
	}
	if t.notNull { // Explicit `gorm:"not null"`
		return false
	}
	if field.Kind == reflect.Ptr { // Pointers are nullable
		return true
	}
	// Check if Go type implies nullability (e.g., sql.NullString)
	goType := field.Type
	if field.Kind == reflect.Ptr && len(goType) > 0 && goType[0] == '*' {
		goType = goType[1:]
	}
	if goNullableTypes[goType] {
		return true
	}

	return false
}

func shouldInclude(field *structmeta.FieldMetadata) bool {
	if len(field.Name) == 0 {
		return false
	}
	if !unicode.IsUpper(rune(field.Name[0])) {
		return false
	}
	if field.Anonymous {
		return false
	}
	return !isRelation(field)
}

func isRelation(field *structmeta.FieldMetadata) bool {
	t := field.Type
	if field.Kind == reflect.Ptr {
		t = strings.TrimPrefix(t, "*")
	}

	if nonRelationTypes[t] {
		return false
	}

	kind := field.Kind
	if kind == reflect.Ptr || kind == reflect.Slice || kind == reflect.Array {
		if field.ReflectType != nil {
			elem := field.ReflectType.Elem()
			if elem != nil {
				kind = elem.Kind()
			}
		}
	}

	return kind == reflect.Struct
}

func tableName(meta structmeta.StructMetadata) string {
	for _, method := range meta.Methods {
		if method.Name != "TableName" {
			continue
		}
		if len(method.Returns) != 1 || method.Returns[0] != "string" {
			continue
		}
		if !method.ReflectValue.IsValid() || method.ReflectValue.IsZero() {
			continue
		}

		results := method.ReflectValue.Call(nil)
		if len(results) > 0 && results[0].Kind() == reflect.String {
			return results[0].String()
		}
	}

	typeName := meta.Name
	if idx := strings.LastIndexByte(typeName, '.'); idx != -1 {
		typeName = typeName[idx+1:]
	}
	return str.Of(typeName).Plural().Snake().String()
}

func collectFieldIndexes(t *tag, colName string) map[string]index {
	indexes := make(map[string]index)

	if t.unique {
		indexes[colName+"_unique"] = index{columns: []string{colName}, unique: true}
	}

	if t.index != "" {
		name := t.index
		idx := indexes[name]
		idx.columns = append(idx.columns, colName)
		if t.indexUnique {
			idx.unique = true
		}
		indexes[name] = idx
	}

	if t.uniqueIndex != "" {
		name := t.uniqueIndex
		idx := indexes[name]
		idx.columns = append(idx.columns, colName)
		idx.unique = true
		indexes[name] = idx
	}

	return indexes
}
