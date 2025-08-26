package console

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	supportconsole "github.com/goravel/framework/support/console"
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
				Name:    "arch",
				Aliases: []string{"a"},
				Usage:   "Target architecture",
				Value:   "amd64",
			},
			&command.StringFlag{
				Name:    "os",
				Aliases: []string{"o"},
				Usage:   "Target os",
			},
			&command.BoolFlag{
				Name:               "static",
				Aliases:            []string{"s"},
				Value:              false,
				Usage:              "Static compilation",
				DisableDefaultText: true,
			},
			&command.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
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
		if os, err = ctx.Choice("Select target os", []console.Choice{
			{Key: "Linux", Value: "linux"},
			{Key: "Windows", Value: "windows"},
			{Key: "Darwin", Value: "darwin"},
		}, console.ChoiceOption{Default: runtime.GOOS}); err != nil {
			ctx.Error(fmt.Sprintf("Select target os error: %v", err))
			return nil
		}
	}

	if err = supportconsole.ExecuteCommand(ctx, generateCommand(ctx.Option("name"), os, ctx.Option("arch"), ctx.OptionBool("static")), "Building..."); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	ctx.Info("Built successfully.")

	return nil
}

func generateCommand(name, os, arch string, static bool) *exec.Cmd {
	args := []string{"build"}

	if static {
		args = append(args, "-ldflags", "-extldflags -static")
	}

	if name != "" {
		args = append(args, "-o", name)
	}

	args = append(args, ".")

	cmd := exec.Command("go", args...)
	cmd.Env = append(cmd.Environ(), "CGO_ENABLED=0", "GOARCH="+arch, "GOOS="+os)

	return cmd
}
