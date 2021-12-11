package console

import (
	"github.com/goravel/framework/console/support"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

var consoleInstance *cli.App

func init() {
	consoleInstance = cli.NewApp()
}

type Application struct {
}

//Init Listen to artisan, Run the registered commands.
func (app *Application) Init() {
	args := os.Args
	app.run(args)
}

//GetInstance Get CLI instance.
func (app *Application) GetInstance() *cli.App {
	return consoleInstance
}

//Register Register all of the commands.
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

		consoleInstance.Commands = append(consoleInstance.Commands, &cliCommand)
	}
}

//run Run the command. args include: ["./main", "artisan", "command"]
func (app *Application) run(args []string) {
	if len(args) > 2 {
		if args[1] == "artisan" {
			cliApp := app.GetInstance()
			var cliArgs []string
			cliArgs = append(cliArgs, args[0])

			for i := 2; i < len(args); i++ {
				cliArgs = append(cliArgs, args[i])
			}

			if err := cliApp.Run(cliArgs); err != nil {
				log.Fatalln(err.Error())
			}

			os.Exit(0)
		}
	}
}
