package console

import (
	"fmt"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/support/file"
)

func GetName(ctx console.Context, ttype, name string, getPath func(string) string) (string, error) {
	if name == "" {
		var err error
		name, err = ctx.Ask(fmt.Sprintf("Enter the %s name", ttype), console.AskOption{
			Validate: func(s string) error {
				if s == "" {
					return fmt.Errorf("the %s name cannot be empty", ttype)
				}

				return nil
			},
		})
		if err != nil {
			return "", err
		}
	}

	if !ctx.OptionBool("force") && file.Exists(getPath(name)) {
		return "", fmt.Errorf("the %s already exists. Use the --force or -f flag to overwrite", ttype)
	}

	return name, nil
}
