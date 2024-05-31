package console

import (
	"fmt"
	"os"
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
func (receiver *BuildCommand) Signature() string {
	return "build"
}

// Description The console command description.
func (receiver *BuildCommand) Description() string {
	return "Build the application"
}

// Extend The console command extend.
func (receiver *BuildCommand) Extend() command.Extend {
	return command.Extend{
		Category: "build",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "system",
				Aliases: []string{"s"},
				Value:   "",
				Usage:   "target system os",
			},
			&command.BoolFlag{
				Name:  "static",
				Value: false,
				Usage: "Static compilation",
			},
			&command.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Value:   "",
				Usage:   "output binary name",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *BuildCommand) Handle(ctx console.Context) error {
	var err error
	if receiver.config.GetString("app.env") == "production" {
		ctx.Warning("**************************************")
		ctx.Warning("*     Application In Production!     *")
		ctx.Warning("**************************************")

		answer, err := ctx.Confirm("Do you really wish to run this command?")
		if err != nil {
			ctx.Error(fmt.Sprintf("Confirm error: %v", err))
			return nil
		}

		if !answer {
			ctx.Warning("Command cancelled!")
			return nil
		}
	}

	system := ctx.Option("system")
	if system == "" {
		system, err = ctx.Choice("Select the target system os", []console.Choice{
			{Key: "Linux", Value: "linux"},
			{Key: "Darwin", Value: "windows"},
			{Key: "Windows", Value: "darwin"},
		})
		if err != nil {
			ctx.Error(fmt.Sprintf("Select the target system error: %v", err))
			return nil
		}
	}

	validSystems := []string{"linux", "windows", "darwin"}
	if !slices.Contains(validSystems, system) {
		err := fmt.Sprintf("Invalid system '%s' specified. Allowed values are: %v", system, validSystems)
		ctx.Error(err)
		return nil
	}

	ctx.Spinner("Building...", console.SpinnerOption{
		Action: func() {
			err = receiver.build(system, generateCommand(ctx.Option("name"), ctx.OptionBool("static")))
		},
	})

	if err != nil {
		ctx.Error(fmt.Sprintf("Build error: %v", err))
	} else {
		ctx.Info("Built successfully.")
	}

	return nil
}

func (receiver *BuildCommand) build(system string, command []string) error {
	os.Setenv("CGO_ENABLED", "0")
	os.Setenv("GOOS", system)
	os.Setenv("GOARCH", "amd64")

	cmd := exec.Command(command[0], command[1:]...)
	_, err := cmd.Output()
	return err
}

func generateCommand(name string, static bool) []string {
	command := []string{"go", "build"}

	if static {
		command = append(command, "-ldflags", "-extldflags -static")
	}

	if name != "" {
		command = append(command, "-o", name)
	}

	return append(command, ".")
}
