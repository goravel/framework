package console

import (
	"os"
	"strings"

	"github.com/gookit/color"
	"github.com/goravel/framework/contracts/console"
	"github.com/urfave/cli/v2"
)

const Version string = "0.6.0"
const EnvironmentFile string = ".env"

type Application struct {
	cli *cli.App
}

//Init Listen to artisan, Run the registered commands.
func (app *Application) Init() *Application {
	app.cli = cli.NewApp()
	app.cli.Name = "Goravel Framework"
	app.cli.Usage = Version
	app.cli.UsageText = "artisan [global options] command [options] [arguments...]"

	return app
}

//Register Register commands.
func (app *Application) Register(commands []console.Command) {
	for _, command := range commands {
		command := command
		cliCommand := cli.Command{
			Name:  command.Signature(),
			Usage: command.Description(),
			Action: func(c *cli.Context) error {
				return command.Handle(c)
			},
		}

		cliCommand.Category = command.Extend().Category
		cliCommand.Flags = command.Extend().Flags
		cliCommand.Subcommands = command.Extend().Subcommands

		app.cli.Commands = append(app.cli.Commands, &cliCommand)
	}
}

//Call Run an Artisan console command by name.
func (app *Application) Call(command string) {
	app.Run(append([]string{os.Args[0], "artisan"}, strings.Split(command, " ")...), false)
}

//CallAndExit Run an Artisan console command by name and exit.
func (app *Application) CallAndExit(command string) {
	app.Run(append([]string{os.Args[0], "artisan"}, strings.Split(command, " ")...), true)
}

//Run a command. Args come from os.Args.
func (app *Application) Run(args []string, exitIfArtisan bool) {
	if len(args) >= 2 {
		if args[1] == "artisan" {
			if len(args) == 2 {
				args = append(args, "--help")
			}

			if args[2] != "-V" && args[2] != "--version" {
				cliArgs := append([]string{args[0]}, args[2:]...)
				if err := app.cli.Run(cliArgs); err != nil {
					panic(err.Error())
				}
			}

			printResult(args[2])

			if exitIfArtisan {
				os.Exit(0)
			}
		}
	}
}

func printResult(command string) {
	switch command {
	case "make:command":
		color.Greenln("Console command created successfully")
	case "-V", "--version":
		color.Greenln("Goravel Framework " + Version)
	}
}
