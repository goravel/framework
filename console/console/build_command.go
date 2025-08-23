package console

import (
	"fmt"
	"os/exec"
	"slices"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type BuildCommand struct {
	config config.Config
}

func NewBuildCommand(config config.Config) *BuildCommand {
	return &BuildCommand{
		config: config,
	}
}

// Signature The name and signature of the console command.
func (r *BuildCommand) Signature() string {
	return "build"
}

// Description The console command description.
func (r *BuildCommand) Description() string {
	return "Build the application"
}

// Extend The console command extend.
func (r *BuildCommand) Extend() command.Extend {
	return command.Extend{
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "os",
				Aliases: []string{"o"},
				Value:   "",
				Usage:   "Target os",
			},
			&command.BoolFlag{
				Name:    "static",
				Aliases: []string{"s"},
				Value:   false,
				Usage:   "Static compilation",
			},
			&command.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Value:   "",
				Usage:   "Output binary name",
			},
		},
	}
}

// Handle Execute the console command.
func (r *BuildCommand) Handle(ctx console.Context) error {
	var err error
	if r.config.GetString("app.env") == "production" {
		ctx.Warning("**************************************")
		ctx.Warning("*     Application In Production!     *")
		ctx.Warning("**************************************")

		if !ctx.Confirm("Do you really wish to run this command?") {
			ctx.Warning("Command cancelled!")
			return nil
		}
	}

	os := ctx.Option("os")
	if os == "" {
		os, err = ctx.Choice("Select target os", []console.Choice{
			{Key: "Linux", Value: "linux"},
			{Key: "Windows", Value: "windows"},
			{Key: "Darwin", Value: "darwin"},
		})
		if err != nil {
			ctx.Error(fmt.Sprintf("Select target os error: %v", err))
			return nil
		}
	}

	validOs := []string{"linux", "windows", "darwin"}
	if !slices.Contains(validOs, os) {
		ctx.Error(fmt.Sprintf("Invalid os '%s' specified. Allowed values are: %v", os, validOs))
		return nil
	}

	if err := ctx.Spinner("Building...", console.SpinnerOption{
		Action: func() error {
			return r.build(os, generateCommand(ctx.Option("name"), ctx.OptionBool("static")))
		},
	}); err != nil {
		ctx.Error(fmt.Sprintf("Build error: %v", err))
	} else {
		ctx.Info("Built successfully.")
	}

	return nil
}

func (r *BuildCommand) build(system string, command []string) error {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Env = append(cmd.Environ(), "CGO_ENABLED=0", "GOARCH=amd64", "GOOS="+system)
	_, err := cmd.Output()
	return err
}

func generateCommand(name string, static bool) []string {
	commands := []string{"go", "build"}

	if static {
		commands = append(commands, "-ldflags", "-extldflags -static")
	}

	if name != "" {
		commands = append(commands, "-o", name)
	}

	return append(commands, ".")
}
