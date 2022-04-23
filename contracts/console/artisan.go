package console

type Artisan interface {
	//Register commands.
	Register(commands []Command)

	//Call Run an Artisan console command by name.
	Call(command string)

	//CallDontExit Run an Artisan console command by name and don't exit.
	CallDontExit(command string)

	//Run a command. args include: ["./main", "artisan", "command"]
	Run(args []string, exitIfArtisan bool)
}
