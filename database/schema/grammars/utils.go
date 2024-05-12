package grammars

import (
	"fmt"
	"reflect"
	"unicode"

	"github.com/spf13/cast"

	schemacontract "github.com/goravel/framework/contracts/database/schema"
)

func addModify(grammar schemacontract.Grammar, sql string, blueprint schemacontract.Blueprint, column schemacontract.ColumnDefinition) string {
	for _, modifier := range grammar.GetModifiers() {
		sql += modifier(blueprint, column)
	}

	return sql
}

func getColumns(grammar schemacontract.Grammar, blueprint schemacontract.Blueprint) []string {
	var columns []string
	for _, column := range blueprint.GetAddedColumns() {
		sql := fmt.Sprintf("%s %s", column.GetName(), getType(grammar, column))

		columns = append(columns, addModify(grammar, sql, blueprint, column))
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

func getType(grammar schemacontract.Grammar, column schemacontract.ColumnDefinition) string {
	t := []rune(column.GetType())
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

func quoteString(value []string) []string {
	for i, v := range value {
		value[i] = "'" + v + "'"
	}

	return value
}
