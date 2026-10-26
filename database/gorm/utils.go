package gorm

import (
	"reflect"
	"regexp"

	"gorm.io/gorm/clause"
)

// plainIdentifierRegex matches ASCII identifiers only, optionally qualified by a table. A name with other characters
// (non ASCII letters, "$", etc.) is treated as an expression and is not quoted.
var plainIdentifierRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`)

// columnCondition returns the SQL for a column in a condition and the arguments it binds.
// A plain identifier becomes a placeholder bound to a clause.Column, so gorm quotes it when the query is built,
// with the dialect of the connection that runs it. Anything else is an expression and is used as is.
func columnCondition(column string) (string, []any) {
	if !isPlainIdentifier(column) {
		return column, nil
	}

	return "?", []any{clause.Column{Name: column}}
}

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
