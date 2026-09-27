package gorm

import (
	"reflect"
	"regexp"
)

var plainIdentifierRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`)

func copyStruct(dest any) reflect.Value {
	t := reflect.TypeOf(dest)
	v := reflect.ValueOf(dest)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		v = v.Elem()
	}

	destFields := make([]reflect.StructField, 0)
	for i := 0; i < t.NumField(); i++ {
		destFields = append(destFields, t.Field(i))
	}
	copyDestStruct := reflect.StructOf(destFields)

	return v.Convert(copyDestStruct)
}

// isPlainIdentifier reports whether the column is a bare identifier, optionally qualified by a table, e.g. "group" or "users.group".
// Anything else (functions, JSON selectors, already quoted names, etc.) is treated as an expression and must not be quoted.
func isPlainIdentifier(column string) bool {
	return plainIdentifierRegex.MatchString(column)
}
