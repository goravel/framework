package migration

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	"gorm.io/gorm/schema"

	"github.com/goravel/framework/errors"
)

var (
	cache          = &sync.Map{}
	namingStrategy = schema.NamingStrategy{}
)

const (
	indexClassPrimary = "PRIMARY"
	indexClassUnique  = "UNIQUE"

	methodBigIncrements        = "BigIncrements"
	methodIncrements           = "Increments"
	methodSmallIncrements      = "SmallIncrements"
	methodTinyIncrements       = "TinyIncrements"
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
	methodString               = "String"
	methodText                 = "Text"
	methodTimestampTz          = "TimestampTz"
	methodBinary               = "Binary"
	methodJson                 = "Json"
	methodUuid                 = "Uuid"
	methodUlid                 = "Ulid"

	modifierUnique   = "Unique"
	modifierNullable = "Nullable"
	modifierDefault  = "Default"
	modifierComment  = "Comment"

	tablePrefix   = "table."
	primaryMethod = "Primary"
	uniqueMethod  = "Unique"
	indexMethod   = "Index"
)

var (
	knownScalarTypes = []string{
		"time.Time",
		"gorm.DeletedAt",
		"sql.Null",
		"datatypes.",
	}

	specialTypeKeywords = map[string]string{
		"deletedat": methodTimestampTz,
		"json":      methodJson,
		"uuid":      methodUuid,
		"ulid":      methodUlid,
	}
)

func Generate(model any) (tableName string, lines []string, err error) {
	if model == nil {
		return "", nil, errors.SchemaInvalidModel
	}

	s, err := schema.Parse(model, cache, namingStrategy)
	if err != nil {
		return "", nil, err
	}

	lines = renderSchema(s)
	if len(lines) == 0 {
		return "", nil, errors.SchemaInvalidModel
	}

	return s.Table, lines, nil
}

func renderSchema(sch *schema.Schema) []string {
	indexes := sch.ParseIndexes()
	lines := make([]string, 0, len(sch.Fields)+len(indexes)+2)

	for _, field := range sch.Fields {
		if isRelationField(field) {
			continue
		}

		line := renderField(field)
		if line != "" {
			lines = append(lines, line)
		}
	}

	if len(indexes) > 0 {
		lines = append(lines, "")
		indexLines := renderIndexes(indexes)
		lines = append(lines, indexLines...)
	}

	return lines
}

func isRelationField(field *schema.Field) bool {
	if field.EmbeddedSchema != nil {
		return true
	}

	if field.FieldType.Kind() == reflect.Struct {
		typeName := field.FieldType.String()
		for _, known := range knownScalarTypes {
			if strings.Contains(typeName, known) {
				return false
			}
		}
		return true
	}

	return false
}

func renderField(field *schema.Field) string {
	if field.IgnoreMigration {
		return ""
	}

	var b strings.Builder
	b.Grow(64)
	b.WriteString(tablePrefix)

	method, args := fieldToMethod(field)
	if method == "" {
		return ""
	}

	b.WriteString(method)
	b.WriteByte('(')
	b.WriteByte('"')
	b.WriteString(field.DBName)
	b.WriteByte('"')

	for _, arg := range args {
		b.WriteString(", ")
		writeValue(&b, arg)
	}

	b.WriteByte(')')

	if field.Unique {
		b.WriteByte('.')
		b.WriteString(modifierUnique)
		b.WriteString("()")
	}

	if !field.NotNull && !field.PrimaryKey {
		b.WriteByte('.')
		b.WriteString(modifierNullable)
		b.WriteString("()")
	}

	if field.HasDefaultValue && field.DefaultValue != "" {
		b.WriteByte('.')
		b.WriteString(modifierDefault)
		b.WriteByte('(')
		b.WriteString(strconv.Quote(field.DefaultValue))
		b.WriteByte(')')
	}

	if field.Comment != "" {
		b.WriteByte('.')
		b.WriteString(modifierComment)
		b.WriteByte('(')
		b.WriteString(strconv.Quote(field.Comment))
		b.WriteByte(')')
	}

	return b.String()
}

func fieldToMethod(field *schema.Field) (method string, args []any) {
	if field.PrimaryKey && field.AutoIncrement {
		switch field.DataType {
		case schema.Int:
			if field.Size <= 32 {
				return methodIncrements, nil
			}
			return methodBigIncrements, nil
		case schema.Uint:
			if field.Size <= 32 {
				return methodIncrements, nil
			}
			return methodBigIncrements, nil
		default:
			return methodBigIncrements, nil
		}
	}

	switch field.DataType {
	case schema.Bool:
		return methodBoolean, nil

	case schema.Int:
		if field.Size <= 8 {
			return methodTinyInteger, nil
		} else if field.Size <= 16 {
			return methodSmallInteger, nil
		} else if field.Size <= 32 {
			return methodInteger, nil
		}
		return methodBigInteger, nil

	case schema.Uint:
		if field.Size <= 8 {
			return methodUnsignedTinyInteger, nil
		} else if field.Size <= 16 {
			return methodUnsignedSmallInteger, nil
		} else if field.Size <= 32 {
			return methodUnsignedInteger, nil
		}
		return methodUnsignedBigInteger, nil

	case schema.Float:
		if field.Size <= 32 {
			return methodFloat, nil
		}
		return methodDouble, nil

	case schema.String:
		if field.Size > 0 && field.Size <= 255 {
			return methodString, []any{field.Size}
		}
		return methodText, nil

	case schema.Time:
		return methodTimestampTz, nil

	case schema.Bytes:
		return methodBinary, nil

	default:
		fieldType := strings.ToLower(field.FieldType.String())
		for keyword, method := range specialTypeKeywords {
			if strings.Contains(fieldType, keyword) {
				return method, nil
			}
		}
		return methodText, nil
	}
}

func renderIndexes(indexes []*schema.Index) []string {
	if len(indexes) == 0 {
		return nil
	}

	lines := make([]string, 0, len(indexes))
	sortedIndexes := make([]*schema.Index, len(indexes))
	copy(sortedIndexes, indexes)
	sort.Slice(sortedIndexes, func(i, j int) bool {
		return sortedIndexes[i].Name < sortedIndexes[j].Name
	})

	for _, idx := range sortedIndexes {
		if strings.EqualFold(idx.Class, indexClassPrimary) {
			lines = append(lines, renderIndex(idx))
			break
		}
	}

	for _, idx := range sortedIndexes {
		if !strings.EqualFold(idx.Class, indexClassPrimary) {
			lines = append(lines, renderIndex(idx))
		}
	}

	return lines
}

func renderIndex(idx *schema.Index) string {
	var b strings.Builder
	b.Grow(40 + len(idx.Fields)*10)
	b.WriteString(tablePrefix)

	isPrimary := strings.EqualFold(idx.Class, indexClassPrimary)
	isUnique := strings.EqualFold(idx.Class, indexClassUnique)

	if isPrimary {
		b.WriteString(primaryMethod)
	} else if isUnique {
		b.WriteString(uniqueMethod)
	} else {
		b.WriteString(indexMethod)
	}

	b.WriteByte('(')

	for i, field := range idx.Fields {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteByte('"')
		b.WriteString(field.Name)
		b.WriteByte('"')
	}

	b.WriteByte(')')
	return b.String()
}

func writeValue(b *strings.Builder, v any) {
	switch val := v.(type) {
	case int:
		b.WriteString(strconv.Itoa(val))
	case int64:
		b.WriteString(strconv.FormatInt(val, 10))
	case uint:
		b.WriteString(strconv.FormatUint(uint64(val), 10))
	case uint64:
		b.WriteString(strconv.FormatUint(val, 10))
	case string:
		b.WriteString(strconv.Quote(val))
	case bool:
		b.WriteString(strconv.FormatBool(val))
	default:
		b.WriteString(strconv.Quote(reflect.ValueOf(v).String()))
	}
}
