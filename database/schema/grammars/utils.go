package grammars

import (
	"fmt"
	"reflect"
	"unicode"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/database/schema"
)

func getCommandByName(commands []*schema.Command, name string) *schema.Command {
	commands = getCommandsByName(commands, name)
	if len(commands) == 0 {
		return nil
	}

	return commands[0]
}

func getCommandsByName(commands []*schema.Command, name string) []*schema.Command {
	var filteredCommands []*schema.Command
	for _, command := range commands {
		if command.Name == name {
			filteredCommands = append(filteredCommands, command)
		}
	}

	return filteredCommands
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
