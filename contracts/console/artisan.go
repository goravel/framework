package console

//go:generate mockery --name=Artisan
type Artisan interface {
	// Register commands.
	Register(commands []Command)

	// Call run an Artisan console command by name.
	Call(command string)

	// CallAndExit run an Artisan console command by name and exit.
	CallAndExit(command string)

	// Run a command. args include: ["./main", "artisan", "command"]
	Run(args []string, exitIfArtisan bool)
}
