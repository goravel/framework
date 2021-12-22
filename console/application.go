package console

import (
	"github.com/goravel/framework/console/support"
	"github.com/goravel/framework/support/testing"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

type Application struct {
	cli *cli.App
}

//Init Listen to artisan, Run the registered commands.
func (app *Application) Init() *Application {
	app.cli = cli.NewApp()

	return app
}

//Register Register commands.
func (app *Application) Register(commands []support.Command) {
	for _, command := range commands {
		command := command
		cliCommand := cli.Command{
			Name:  command.Signature(),
			Usage: command.Description(),
			Action: func(c *cli.Context) error {
				return command.Handle(c)
			},
		}

		if len(command.Flags()) > 0 {
			cliCommand.Flags = command.Flags()
		}

		if len(command.Subcommands()) > 0 {
			cliCommand.Subcommands = command.Subcommands()
		}

		app.cli.Commands = append(app.cli.Commands, &cliCommand)
	}
}

//Call Run an Artisan console command by name.
func (app *Application) Call(command string) {
	app.Run(append([]string{os.Args[0], "artisan"}, strings.Split(command, " ")...))
}

//Run Run a command. Args come from os.Args.
func (app *Application) Run(args []string) {
	if len(args) > 2 {
		if args[1] == "artisan" {
			cliArgs := append([]string{args[0]}, args[2:]...)
			if err := app.cli.Run(cliArgs); err != nil {
				panic(err.Error())
			}

			if !testing.RunInTest() {
				os.Exit(0)
			}
		}
	}
}
