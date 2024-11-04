package grammars

import (
	"fmt"
	"reflect"
	"unicode"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/database/schema"
)

func addModify(modifiers []func(schema.Blueprint, schema.ColumnDefinition) string, sql string, blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	for _, modifier := range modifiers {
		sql += modifier(blueprint, column)
	}

	return sql
}

func getColumn(grammar schema.Grammar, blueprint schema.Blueprint, column schema.ColumnDefinition) string {
	sql := fmt.Sprintf("%s %s", column.GetName(), getType(grammar, column))

	return addModify(grammar.GetModifiers(), sql, blueprint, column)
}

func getColumns(grammar schema.Grammar, blueprint schema.Blueprint) []string {
	var columns []string
	for _, column := range blueprint.GetAddedColumns() {
		columns = append(columns, getColumn(grammar, blueprint, column))
	}

	return columns
}

func getDefaultValue(def any) string {
	switch def.(type) {
	case bool:
		return "'" + cast.ToString(cast.ToInt(def)) + "'"
	}

	return "'" + cast.ToString(def) + "'"
}

func getType(grammar schema.Grammar, column schema.ColumnDefinition) string {
	t := []rune(column.GetType())
	if len(t) == 0 {
		return ""
	}

	t[0] = unicode.ToUpper(t[0])
	methodName := fmt.Sprintf("Type%s", string(t))
	methodValue := reflect.ValueOf(grammar).MethodByName(methodName)
	if methodValue.IsValid() {
		args := []reflect.Value{reflect.ValueOf(column)}
		callResult := methodValue.Call(args)

		return callResult[0].String()
	}

	return ""
}

func prefixArray(prefix string, values []string) []string {
	for i, value := range values {
		values[i] = prefix + " " + value
	}

	return values
}
