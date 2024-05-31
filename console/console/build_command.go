package console

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
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
	if receiver.config.GetString("app.env") == "production" {
		color.Yellow().Println("**************************************")
		color.Yellow().Println("*     Application In Production!     *")
		color.Yellow().Println("**************************************")

		answer, err := ctx.Confirm("Do you really wish to run this command?")
		if err != nil {
			return err
		}

		if !answer {
			color.Yellow().Println("Command cancelled!")
			return nil
		}
	}

	system := ctx.Option("system")
	if system == "" {
		var err error
		system, err = ctx.Choice("Select the target system os", []console.Choice{
			{Key: "Linux", Value: "linux"},
			{Key: "Darwin", Value: "windows"},
			{Key: "Windows", Value: "darwin"},
		})
		if err != nil {
			color.Red().Println(err)
			return nil
		}
	}

	validSystems := []string{"linux", "windows", "darwin"}
	if !slices.Contains(validSystems, system) {
		err := fmt.Sprintf("Invalid system '%s' specified. Allowed values are: %v", system, validSystems)
		color.Red().Println(err)
		return nil
	}

	if err := receiver.build(system, generateCommand(ctx.Option("name"), ctx.OptionBool("static"))); err != nil {
		color.Red().Println(err.Error())

		return nil
	}

	color.Green().Println("Built successfully.")

	return nil
}

func (receiver *BuildCommand) build(system, command string) error {
	os.Setenv("CGO_ENABLED", "0")
	os.Setenv("GOOS", system)
	os.Setenv("GOARCH", "amd64")

	commandArr := strings.Split(command, " ")

	cmd := exec.Command(commandArr[0], commandArr[1:]...)
	_, err := cmd.Output()
	return err
}

func generateCommand(name string, static bool) string {
	command := "go build"

	if static {
		command += " -ldflags -extldflags -static"
	}

	if name != "" {
		command += " -o " + name
	}

	return command + " ."
}
