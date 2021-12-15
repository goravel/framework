package console

import (
	"github.com/goravel/framework/console/support"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

type Application struct {
	cli *cli.App
}

//Init Listen to artisan, Run the registered commands.
func (app *Application) Init() *Application {
	app.cli = cli.NewApp()
	args := os.Args
	app.run(args)

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
func (app *Application) Call(command string)  {

}

//run Run a command. args include: ["./main", "artisan", "command"]
func (app *Application) run(args []string) {
	if len(args) > 2 {
		if args[1] == "artisan" {
			var cliArgs []string
			cliArgs = append(cliArgs, args[0])

			for i := 2; i < len(args); i++ {
				cliArgs = append(cliArgs, args[i])
			}

			if err := app.cli.Run(cliArgs); err != nil {
				log.Fatalln(err.Error())
			}

			os.Exit(0)
		}
	}
}
