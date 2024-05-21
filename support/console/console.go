package console

import (
	"errors"
	"fmt"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/support/file"
)

type Option struct {
	Question string
	Field    string
	Required bool
	GetPath  func(string) string
	Type     string
}

// GetArgument Get the argument from the console context.(options: question, field)
func GetArgument(ctx console.Context, index int, options ...Option) (string, error) {
	name := ctx.Argument(index)
	var required bool
	t := "file"
	if name == "" {
		var err error
		question := "Enter the argument"
		field := "argument"
		if len(options) > 0 {
			if options[0].Question != "" {
				question = options[0].Question
			}

			if options[0].Field != "" {
				field = options[0].Field
			}

			required = options[0].Required
		}

		name, err = ctx.Ask(question, console.AskOption{
			Validate: func(s string) error {
				if s == "" && required {
					return errors.New(field + " cannot be empty")
				}

				return nil
			},
		})
		if err != nil {
			return "", err
		}
	}

	force := ctx.OptionBool("force")
	path := name
	if len(options) > 0 && options[0].GetPath != nil {
		if options[0].Type != "" {
			t = options[0].Type
		}
		if options[0].GetPath != nil {
			path = options[0].GetPath(name)
		}
	}

	if !force && file.Exists(path) {
		return "", errors.New(fmt.Sprintf("the %s already exists. Use the --force flag to overwrite", t))
	}

	return name, nil
}
