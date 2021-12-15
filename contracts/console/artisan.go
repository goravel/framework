package console

import "github.com/goravel/framework/console/support"

type Artisan interface {
	//Call Run an Artisan console command by name.
	Call(command string)

	//Register Register commands.
	Register(commands []support.Command)
}
