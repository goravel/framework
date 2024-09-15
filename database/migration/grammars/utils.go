package grammars

import (
	"fmt"
	"reflect"
	"unicode"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/database/migration"
)

func addModify(grammar migration.Grammar, sql string, blueprint migration.Blueprint, column migration.ColumnDefinition) string {
	for _, modifier := range grammar.GetModifiers() {
		sql += modifier(blueprint, column)
	}

	return sql
}

func getColumns(grammar migration.Grammar, blueprint migration.Blueprint) []string {
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

func getType(grammar migration.Grammar, column migration.ColumnDefinition) string {
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
