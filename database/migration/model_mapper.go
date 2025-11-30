package migration

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"gorm.io/gorm/schema"

	"github.com/goravel/framework/errors"
)

// Blueprint method names for column definitions.
const (
	methodBigIncrements        = "BigIncrements"
	methodIncrements           = "Increments"
	methodBoolean              = "Boolean"
	methodTinyInteger          = "TinyInteger"
	methodSmallInteger         = "SmallInteger"
	methodInteger              = "Integer"
	methodBigInteger           = "BigInteger"
	methodUnsignedTinyInteger  = "UnsignedTinyInteger"
	methodUnsignedSmallInteger = "UnsignedSmallInteger"
	methodUnsignedInteger      = "UnsignedInteger"
	methodUnsignedBigInteger   = "UnsignedBigInteger"
	methodFloat                = "Float"
	methodDouble               = "Double"
	methodDecimal              = "Decimal"
	methodString               = "String"
	methodText                 = "Text"
	methodBinary               = "Binary"
	methodJson                 = "Json"
	methodJsonb                = "Jsonb"
	methodEnum                 = "Enum"
	methodUuid                 = "Uuid"
	methodUlid                 = "Ulid"
	methodDate                 = "Date"
	methodTime                 = "Time"
	methodTimestamp            = "Timestamp"
	methodTimestampTz          = "TimestampTz"
)

// Blueprint method names for index definitions.
const (
	methodPrimary  = "Primary"
	methodUnique   = "Unique"
	methodIndex    = "Index"
	methodFullText = "FullText"
)

// Blueprint method names for modifiers.
const (
	methodUnsigned  = "Unsigned"
	methodNullable  = "Nullable"
	methodDefault   = "Default"
	methodComment   = "Comment"
	methodPlaces    = "Places"
	methodTotal     = "Total"
	methodName      = "Name"
	methodAlgorithm = "Algorithm"
)

// Index class constants matching GORM's index classification.
const (
	classUnique   = "UNIQUE"
	classPrimary  = "PRIMARY"
	classFullText = "FULLTEXT"
)

const tablePrefix = "table."

// dataTypeMapping maps GORM DataType strings to blueprint method names.
// Used for custom TYPE tags that aren't standard GORM DataTypes.
var dataTypeMapping = map[string]string{
	"jsonb":     methodJsonb,
	"json":      methodJson,
	"text":      methodText,
	"binary":    methodBinary,
	"varbinary": methodBinary,
	"blob":      methodBinary,
	"decimal":   methodDecimal,
	"numeric":   methodDecimal,
	"uuid":      methodUuid,
	"ulid":      methodUlid,
	"date":      methodDate,
	"time":      methodTime,
}

// stringTypePrefixes contains SQL string type prefixes for type detection.
var stringTypePrefixes = []string{"char", "varchar", "nvarchar", "varchar2", "nchar"}

var (
	schemaCache           = &sync.Map{}
	defaultNamingStrategy = schema.NamingStrategy{}
)

func Generate(model any) (string, []string, error) {
	sch, err := schema.Parse(model, schemaCache, defaultNamingStrategy)
	if err != nil {
		return "", nil, err
	}

	lines := renderSchema(sch)
	if len(lines) == 0 {
		return "", nil, errors.SchemaInvalidModel
	}

	return sch.Table, lines, nil
}

func renderSchema(sch *schema.Schema) []string {
	var lines []string
	seen := make(map[string]bool)
	var fields []*schema.Field

	// 1. Render Columns
	for _, field := range sch.Fields {
		if shouldSkipField(field) || seen[field.DBName] {
			continue
		}
		seen[field.DBName] = true
		fields = append(fields, field)

		if line := renderField(field); line != "" {
			lines = append(lines, line)
		}
	}

	// 2. Render Indexes
	if idxLines := renderIndexes(sch, fields); len(idxLines) > 0 {
		lines = append(lines, "", strings.Join(idxLines, "\n"))
	}

	return lines
}

func shouldSkipField(field *schema.Field) bool {
	// Skip ignored or embedded fields
	if field.IgnoreMigration || field.EmbeddedSchema != nil || field.DBName == "" {
		return true
	}

	// Skip relation fields
	rels := &field.Schema.Relationships
	rels.Mux.RLock()
	_, isRel := rels.Relations[field.Name]
	rels.Mux.RUnlock()
	if isRel {
		return true
	}

	// Skip foreign key fields (belong to a relation)
	for _, rel := range field.Schema.Relationships.Relations {
		for _, ref := range rel.References {
			if ref.ForeignKey != nil && ref.ForeignKey.DBName == field.DBName {
				return true
			}
		}
	}

	return false
}

func renderField(f *schema.Field) string {
	method, args := fieldToMethod(f)
	if method == "" {
		return ""
	}

	b := &atom{Builder: strings.Builder{}}
	b.Grow(64)
	b.WriteString(tablePrefix)

	// Chain: table.Method("name", args...).Nullable()...
	b.WriteMethod(method, append([]any{f.DBName}, args...)...)

	// Check for unsigned modifier
	rawType := strings.ToLower(string(f.DataType))
	if !strings.Contains(method, "Unsigned") {
		if strings.Contains(rawType, "unsigned") || f.TagSettings["UNSIGNED"] != "" {
			b.WriteMethod(methodUnsigned)
		}
	}

	if method == methodDecimal {
		if f.Scale > 0 {
			b.WriteMethod(methodPlaces, f.Scale)
		}
		if f.Precision > 0 {
			b.WriteMethod(methodTotal, f.Precision)
		}
	}
	if !f.NotNull && !f.PrimaryKey {
		b.WriteMethod(methodNullable)
	}
	if f.HasDefaultValue && f.DefaultValueInterface != nil {
		b.WriteMethod(methodDefault, f.DefaultValueInterface)
	}
	if f.Comment != "" {
		b.WriteMethod(methodComment, trimQuotes(f.Comment))
	}

	return b.String()
}

func fieldToMethod(f *schema.Field) (string, []any) {
	if f.PrimaryKey && f.AutoIncrement {
		if f.Size <= 32 {
			return methodIncrements, nil
		}
		return methodBigIncrements, nil
	}

	switch f.DataType {
	case schema.Bool:
		return methodBoolean, nil
	case schema.Int:
		return intMethod(f.Size, false), nil
	case schema.Uint:
		return intMethod(f.Size, true), nil
	case schema.Float:
		if f.Size <= 32 {
			return methodFloat, nil
		}
		return methodDouble, nil
	case schema.String:
		if f.Size > 0 {
			return methodString, []any{f.Size}
		}
		return methodString, nil
	case schema.Time:
		if f.Precision > 0 {
			return methodTimestampTz, []any{f.Precision}
		}
		return methodTimestampTz, nil
	case schema.Bytes:
		return methodBinary, nil
	}

	// String-based Type Inference (Enums, Custom types)
	sType := strings.ToLower(string(f.DataType))

	if strings.HasPrefix(sType, "enum") {
		return methodEnum, []any{parseEnum(string(f.DataType))}
	}

	// Helper to check prefixes fast
	for _, p := range stringTypePrefixes {
		if strings.HasPrefix(sType, p) {
			if size := parseTypeSize(sType); size > 0 {
				return methodString, []any{size}
			}
			return methodString, nil
		}
	}

	if strings.HasPrefix(sType, "timestamp") || strings.HasPrefix(sType, "datetime") {
		if strings.Contains(sType, "tz") {
			return methodTimestampTz, nil
		}
		return methodTimestamp, nil
	}

	// Map lookup for fixed types (json, uuid, etc)
	for k, v := range dataTypeMapping {
		if strings.Contains(sType, k) {
			return v, nil
		}
	}

	// Fallback to Go type name
	goType := strings.ToLower(f.FieldType.String())
	if strings.Contains(goType, "json") {
		return methodJson, nil
	}
	if strings.Contains(goType, "uuid") {
		return methodUuid, nil
	}
	if strings.Contains(goType, "ulid") {
		return methodUlid, nil
	}

	return methodText, nil
}

func renderIndexes(sch *schema.Schema, fields []*schema.Field) []string {
	var lines []string
	seen := make(map[string]bool)

	add := func(key, line string) {
		if !seen[key] {
			seen[key] = true
			lines = append(lines, line)
		}
	}

	// Composite Primary Keys (if > 1 PK field)
	if len(sch.PrimaryFields) > 1 {
		cols := getColNames(sch.PrimaryFields)
		add("PK:"+strings.Join(cols, ","), formatIndex(methodPrimary, cols, nil))
	}

	indexes := sch.ParseIndexes()
	sort.Slice(indexes, func(i, j int) bool { return indexes[i].Name < indexes[j].Name })

	for _, idx := range indexes {
		if len(idx.Fields) == 0 {
			continue
		}

		cols := make([]string, 0, len(idx.Fields))
		for _, f := range idx.Fields {
			if f.DBName != "" {
				cols = append(cols, f.DBName)
			}
		}
		if len(cols) == 0 {
			continue
		}

		method := methodIndex
		switch idx.Class {
		case classUnique:
			method = methodUnique
		case classFullText:
			method = methodFullText
		case classPrimary:
			method = methodPrimary
		}

		add(idx.Class+":"+strings.Join(cols, ","), formatIndex(method, cols, idx))
	}

	// Field Unique Constraints
	for _, f := range fields {
		if f.Unique && f.DBName != "" {
			add("UQ:"+f.DBName, formatIndex(methodUnique, []string{f.DBName}, nil))
		}
	}

	return lines
}

func intMethod(size int, unsigned bool) string {
	switch {
	case size <= 8:
		if unsigned {
			return methodUnsignedTinyInteger
		}
		return methodTinyInteger
	case size <= 16:
		if unsigned {
			return methodUnsignedSmallInteger
		}
		return methodSmallInteger
	case size <= 32:
		if unsigned {
			return methodUnsignedInteger
		}
		return methodInteger
	default:
		if unsigned {
			return methodUnsignedBigInteger
		}
		return methodBigInteger
	}
}

func formatIndex(method string, cols []string, idx *schema.Index) string {
	b := &atom{Builder: strings.Builder{}}
	b.WriteString(tablePrefix)

	args := make([]any, len(cols))
	for i, v := range cols {
		args[i] = v
	}
	b.WriteMethod(method, args...)

	if idx != nil {
		if idx.Type != "" {
			b.WriteMethod(methodAlgorithm, idx.Type)
		}
		// Only write name if not auto-generated (idx_table_col)
		if idx.Name != "" && !strings.HasSuffix(idx.Name, "_"+cols[0]) {
			b.WriteMethod(methodName, idx.Name)
		}
	}
	return b.String()
}

func parseEnum(def string) []any {
	start, end := strings.IndexByte(def, '('), strings.LastIndexByte(def, ')')
	if start == -1 || end <= start+1 {
		return nil
	}

	var values []any
	var buf strings.Builder
	inQuote := false

	for _, r := range def[start+1 : end] {
		if r == '\'' {
			inQuote = !inQuote
			continue
		}
		if r == ',' && !inQuote {
			values = append(values, parseVal(buf.String()))
			buf.Reset()
			continue
		}
		buf.WriteRune(r)
	}
	if buf.Len() > 0 {
		values = append(values, parseVal(buf.String()))
	}
	return values
}

func parseVal(s string) any {
	s = strings.TrimSpace(s)
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if strings.Contains(s, ".") {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
	}
	return s
}

func trimQuotes(s string) string {
	if len(s) >= 2 && (s[0] == '\'' || s[0] == '"') && s[0] == s[len(s)-1] {
		return s[1 : len(s)-1]
	}
	return s
}

func parseTypeSize(s string) int {
	start, end := strings.IndexByte(s, '('), strings.IndexByte(s, ')')
	if start > -1 && end > start {
		// Just grab the first number "varchar(255)" -> 255
		parts := strings.SplitN(s[start+1:end], ",", 2)
		if n, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
			return n
		}
	}
	return 0
}

func getColNames(fields []*schema.Field) []string {
	names := make([]string, 0, len(fields))
	for _, f := range fields {
		if f.DBName != "" {
			names = append(names, f.DBName)
		}
	}
	return names
}

type atom struct {
	strings.Builder
}

func (r *atom) WriteMethod(name string, args ...any) {
	if r.Len() > 0 && r.String()[r.Len()-1] != '.' {
		r.WriteByte('.')
	}
	r.WriteString(name)
	r.WriteByte('(')
	for i, arg := range args {
		if i > 0 {
			r.WriteString(", ")
		}
		r.writeValue(arg)
	}
	r.WriteByte(')')
}

func (r *atom) writeValue(v any) {
	switch val := v.(type) {
	case nil:
		r.WriteString("nil")
	case string:
		r.WriteString(strconv.Quote(val))
	case int:
		r.WriteString(strconv.Itoa(val))
	case int64:
		r.WriteString(strconv.FormatInt(val, 10))
	case float64:
		r.WriteString(strconv.FormatFloat(val, 'f', -1, 64))
	case bool:
		r.WriteString(strconv.FormatBool(val))
	case []any:
		r.WriteString("[]any{")
		for i, item := range val {
			if i > 0 {
				r.WriteString(", ")
			}
			r.writeValue(item)
		}
		r.WriteByte('}')
	default:
		_, _ = fmt.Fprintf(r, "%v", val)
	}
}
