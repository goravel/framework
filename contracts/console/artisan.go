package console

import "github.com/goravel/framework/console/support"

type Artisan interface {
	//Register Register commands.
	Register(commands []support.Command)

	//Call Run an Artisan console command by name.
	Call(command string)

	//Run Run a command. args include: ["./main", "artisan", "command"]
	Run(args []string)
}
