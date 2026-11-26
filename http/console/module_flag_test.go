package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

// Every make command of the package accepts --module.
func TestMakeCommands_ModuleFlag(t *testing.T) {
	commands := []console.Command{
		&ControllerMakeCommand{},
		&MiddlewareMakeCommand{},
		&RequestMakeCommand{},
	}

	for _, cmd := range commands {
		var names []string
		for _, flag := range cmd.Extend().Flags {
			if f, ok := flag.(*command.StringFlag); ok {
				names = append(names, f.Name)
			}
		}

		assert.Contains(t, names, "module", cmd.Signature())
	}
}
