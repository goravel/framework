package migration

import (
	"reflect"
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

	embeddedFieldCapacity = 4
)

// Generate generates database migration schema lines from a Go struct model.
// It returns the table name, migration lines, and any error encountered.
//
// The model parameter should be a struct or pointer to struct with GORM tags.
// Supported GORM tags: column, type, size, precision, scale, default, comment,
// primaryKey, unique, index, not null, nullable, unsigned, enum.
//
// Example:
//
//	type User struct {
//	    ID    uint   `gorm:"primaryKey"`
//	    Name  string `gorm:"size:100;not null"`
//	    Email string `gorm:"unique;index"`
//	}
//
//	tableName, lines, err := Generate(&User{})
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
	table      string
	columns    []column
	indexes    map[string]index
	hasID      bool
	timestamps bool
	softDelete bool
}

type column struct {
	name      string
	method    string
	args      []any
	enum      []any
	modifiers []modifier
}

type index struct {
	columns []string
	unique  bool
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

	// Scan for embedded fields like gorm.Model
	embedded := scanEmbedded(meta)
	s.hasID = embedded["ID"]
	s.timestamps = embedded["CreatedAt"] && embedded["UpdatedAt"]
	s.softDelete = embedded["DeletedAt"]

	for _, field := range meta.Fields {
		if !shouldInclude(&field, embedded) {
			continue
		}

		if field.Name == "ID" {
			s.hasID = true
			continue
		}

		col := buildColumn(&field)
		if col == nil {
			continue
		}

		s.columns = append(s.columns, *col)
		collectIndexes(&field, col.name, s.indexes)
	}

	if !s.hasID && len(s.columns) == 0 && !s.timestamps && !s.softDelete {
		return nil, errors.SchemaInvalidModel
	}

	return s, nil
}

func buildColumn(field *structmeta.FieldMetadata) *column {
	t := parseTag(field)
	if t.ignore {
		return nil
	}

	name := t.column
	if name == "" {
		name = str.Of(field.Name).Snake().String()
	}

	method, args := mapType(field, t)
	if method == "" {
		return nil
	}

	col := &column{
		name:   name,
		method: method,
		args:   args,
		enum:   t.enumValues,
	}

	// Build modifiers slice - common case is 0-2 modifiers
	mods := make([]modifier, 0, 2)

	if t.unsigned {
		mods = append(mods, modifier{"Unsigned", ""})
	}

	// Decimal precision/scale modifiers
	if method == "Decimal" {
		if t.precision > 0 {
			mods = append(mods, modifier{"Total", strconv.Itoa(t.precision)})
		}
		if t.scale > 0 {
			mods = append(mods, modifier{"Places", strconv.Itoa(t.scale)})
		}
	}

	if isNullable(field, t) {
		mods = append(mods, modifier{"Nullable", ""})
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

// renderLines converts tableSchema to migration code lines.
func renderLines(s *tableSchema) []string {
	capacity := len(s.columns) + len(s.indexes) + 3
	if len(s.indexes) > 0 {
		capacity++
	}
	lines := make([]string, 0, capacity)

	if s.hasID {
		lines = append(lines, "table.ID()")
	}

	for i := range s.columns {
		lines = append(lines, renderColumn(&s.columns[i]))
	}

	if s.timestamps {
		lines = append(lines, "table.Timestamps()")
	}

	if s.softDelete {
		lines = append(lines, "table.SoftDeletes()")
	}

	if len(s.indexes) > 0 {
		lines = append(lines, "")
		for _, idx := range s.indexes {
			lines = append(lines, renderIndex(idx))
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
		buf := strconv.AppendInt(nil, val, 10)
		b.Write(buf)
	case uint:
		buf := strconv.AppendUint(nil, uint64(val), 10)
		b.Write(buf)
	case uint64:
		buf := strconv.AppendUint(nil, val, 10)
		b.Write(buf)
	case int32:
		buf := strconv.AppendInt(nil, int64(val), 10)
		b.Write(buf)
	case uint32:
		buf := strconv.AppendUint(nil, uint64(val), 10)
		b.Write(buf)
	case float64:
		buf := strconv.AppendFloat(nil, val, 'g', -1, 64)
		b.Write(buf)
	case float32:
		buf := strconv.AppendFloat(nil, float64(val), 'g', -1, 32)
		b.Write(buf)
	case bool:
		if val {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	default:
		b.WriteString("nil")
	}
}

func renderIndex(idx index) string {
	var b strings.Builder
	b.Grow(avgIndexLine + len(idx.columns)*avgColNameLen)

	b.WriteString("table.")
	if idx.unique {
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
	column     string
	dbType     string
	defaultVal string
	comment    string
	enumValues []any
	size       int
	precision  int
	scale      int
	primaryKey bool
	unique     bool
	index      bool
	notNull    bool
	nullable   bool
	unsigned   bool
	ignore     bool
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

		if idx := strings.IndexByte(part, ':'); idx != -1 {
			key := strings.ToLower(strings.TrimSpace(part[:idx]))
			val := strings.TrimSpace(part[idx+1:])
			parseKeyValue(t, key, val)
		} else {
			parseKeyValue(t, strings.ToLower(part), "")
		}
	}

	return t
}

func parseKeyValue(t *tag, key, val string) {
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
	case "size":
		t.size, _ = strconv.Atoi(val)
	case "default":
		t.defaultVal = strings.Trim(val, `"'`)
	case "comment":
		t.comment = strings.Trim(val, `"'`)
	case "unique":
		t.unique = true
	case "index":
		t.index = true
	case "not null":
		t.notNull = true
	case "null", "nullable":
		t.nullable = true
	case "unsigned":
		t.unsigned = true
	case "primarykey", "primary_key":
		t.primaryKey = true
	case "precision":
		t.precision, _ = strconv.Atoi(val)
	case "scale":
		t.scale, _ = strconv.Atoi(val)
	case "enum":
		parseEnum(t, val)
	}
}

// parseDecimal extracts precision and scale from decimal type strings.
// Handles formats like "decimal(10,2)" or "DECIMAL(10, 2)".
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

	// Both precision and scale
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

	if goType == "string" {
		if t.size > 0 {
			return "String", []any{t.size}
		}
		return "Text", nil
	}

	switch goType {
	case "uint":
		return "UnsignedInteger", nil
	case "int", "int32":
		return "Integer", nil
	case "int64":
		return "BigInteger", nil
	case "uint64":
		return "UnsignedBigInteger", nil
	case "bool":
		return "Boolean", nil
	case "time.Time", "carbon.DateTime":
		return "DateTimeTz", nil
	case "float64":
		return "Double", nil
	case "float32":
		return "Float", nil
	case "int16":
		return "SmallInteger", nil
	case "int8":
		return "TinyInteger", nil
	case "uint32":
		return "UnsignedInteger", nil
	case "uint16":
		return "UnsignedSmallInteger", nil
	case "uint8":
		return "UnsignedTinyInteger", nil
	case "[]byte", "[]uint8":
		return "Binary", nil
	case "gorm.DeletedAt":
		return "SoftDeletesTz", nil
	}

	kind := field.Kind
	if kind == reflect.Ptr && field.ReflectType != nil {
		kind = field.ReflectType.Elem().Kind()
	}

	if kind == reflect.Map || kind == reflect.Slice {
		return "Json", nil
	}

	return "String", nil
}

func mapDBType(dbType string, size int) (method string, args []any) {
	s := strings.ToLower(dbType)
	if size == 0 {
		size = extractSize(s)
	}

	if strings.HasPrefix(s, "varchar") || strings.HasPrefix(s, "char") {
		if size > 0 {
			return "String", []any{size}
		}
		return "String", nil
	}
	if strings.HasPrefix(s, "decimal") {
		return "Decimal", nil
	}
	if strings.HasPrefix(s, "enum") {
		return "Enum", nil
	}
	if strings.HasPrefix(s, "text") || strings.Contains(s, "text") { // tinytext, mediumtext
		return "Text", nil
	}
	if strings.HasPrefix(s, "uuid") {
		return "Uuid", nil
	}
	if strings.HasPrefix(s, "ulid") {
		return "Ulid", nil
	}
	if strings.HasPrefix(s, "bigint") {
		return "BigInteger", nil
	}
	if strings.HasPrefix(s, "int") { // int, integer, smallint, tinyint
		return "Integer", nil
	}
	if strings.HasPrefix(s, "json") { // json, jsonb
		return "Json", nil
	}
	if strings.HasPrefix(s, "bool") {
		return "Boolean", nil
	}
	if strings.HasPrefix(s, "time") || strings.HasPrefix(s, "date") { // time, date, datetime, timestamp
		return "DateTimeTz", nil
	}
	if strings.HasPrefix(s, "binary") || strings.HasPrefix(s, "blob") {
		return "Binary", nil
	}

	// Fallback: quote the dbType as it's a string literal
	return "Column", []any{strconv.Quote(dbType)}
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
	if t.nullable {
		return true
	}
	if t.notNull {
		return false
	}
	return field.Kind == reflect.Ptr
}

func shouldInclude(field *structmeta.FieldMetadata, embedded map[string]bool) bool {
	if len(field.Name) == 0 {
		return false
	}

	if !unicode.IsUpper(rune(field.Name[0])) {
		return false
	}

	if embedded[field.Name] {
		return false
	}

	return !isRelation(field)
}

func isRelation(field *structmeta.FieldMetadata) bool {
	// Time types are not relations despite being structs
	t := field.Type
	if strings.Contains(t, "time.Time") || strings.Contains(t, "gorm.DeletedAt") {
		return false
	}

	// Check kind - struct types are usually relations
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

func scanEmbedded(meta structmeta.StructMetadata) map[string]bool {
	embedded := make(map[string]bool, embeddedFieldCapacity)

	for _, field := range meta.Fields {
		if !field.Anonymous {
			continue
		}

		t := field.Type
		if t == "gorm.Model" || t == "*gorm.Model" || t == "orm.Model" || t == "*orm.Model" {
			embedded["ID"] = true
			embedded["CreatedAt"] = true
			embedded["UpdatedAt"] = true
			embedded["DeletedAt"] = true
			continue
		}

		if t == "orm.SoftDeletes" || t == "*orm.SoftDeletes" {
			embedded["DeletedAt"] = true
			continue
		}

		if t == "orm.Timestamps" || t == "*orm.Timestamps" {
			embedded["CreatedAt"] = true
			embedded["UpdatedAt"] = true
		}
	}

	return embedded
}

func tableName(meta structmeta.StructMetadata) string {
	// Check for TableName() method
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

	if len(typeName) > 0 && typeName[0] == '*' {
		typeName = typeName[1:]
	}

	if idx := strings.LastIndexByte(typeName, '.'); idx != -1 {
		typeName = typeName[idx+1:]
	}

	return str.Of(typeName).Plural().Snake().String()
}

func collectIndexes(field *structmeta.FieldMetadata, colName string, indexes map[string]index) {
	if field.Tag == nil {
		return
	}

	tagStr := field.Tag.Get("gorm")
	if tagStr == "" {
		return
	}

	parts := strings.Split(tagStr, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		var key, val string
		if idx := strings.IndexByte(part, ':'); idx != -1 {
			key = strings.ToLower(strings.TrimSpace(part[:idx]))
			val = strings.TrimSpace(part[idx+1:])
		} else {
			key = strings.ToLower(part)
		}

		switch key {
		case "unique":
			if val == "" {
				indexes[colName+"_unique"] = index{columns: []string{colName}, unique: true}
			}
		case "index":
			if val == "" {
				indexes[colName+"_index"] = index{columns: []string{colName}, unique: false}
			} else {
				// Extract index name (before comma if present)
				name := val
				if comma := strings.IndexByte(val, ','); comma != -1 {
					name = val[:comma]
				}
				idx := indexes[name]
				idx.columns = append(idx.columns, colName)
				idx.unique = false
				indexes[name] = idx
			}
		case "uniqueindex", "unique_index":
			if val != "" {
				name := val
				if comma := strings.IndexByte(val, ','); comma != -1 {
					name = val[:comma]
				}
				idx := indexes[name]
				idx.columns = append(idx.columns, colName)
				idx.unique = true
				indexes[name] = idx
			}
		}
	}
}
