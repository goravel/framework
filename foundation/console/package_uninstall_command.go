package console

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
)

type PackageUninstallCommand struct {
}

func NewPackageUninstallCommand() *PackageUninstallCommand {
	return &PackageUninstallCommand{}
}

// Signature The name and signature of the console command.
func (r *PackageUninstallCommand) Signature() string {
	return "package:uninstall"
}

// Description The console command description.
func (r *PackageUninstallCommand) Description() string {
	return "Uninstall a package"
}

// Extend The console command extend.
func (r *PackageUninstallCommand) Extend() command.Extend {
	return command.Extend{
		ArgsUsage: " <package>",
		Category:  "package",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:               "force",
				Aliases:            []string{"f"},
				Usage:              "Continue uninstalling process even if an error occurs",
				DisableDefaultText: true,
			},
		},
	}
}

// Handle Execute the console command.
func (r *PackageUninstallCommand) Handle(ctx console.Context) error {
	var (
		err error
		pkg = ctx.Argument(0)
	)
	if pkg == "" {
		pkg, err = ctx.Ask("Enter the package name to uninstall", console.AskOption{
			Placeholder: " E.g example.com/pkg",
			Prompt:      ">",
			Validate: func(s string) error {
				if s == "" {
					return errors.CommandEmptyPackageName
				}

				return nil
			},
		})
		if err != nil {
			ctx.Error(err.Error())
			return nil
		}
	}

	pkg, _, _ = strings.Cut(pkg, "@")
	manager := pkg + "/manager"

	// uninstall package
	var output []byte
	uninstall := exec.Command("go", "run", manager, "uninstall")
	if ctx.OptionBool("force") {
		uninstall.Args = append(uninstall.Args, "--force")
	}

	if err = ctx.Spinner(fmt.Sprintf("> @%s", strings.Join(uninstall.Args, " ")), console.SpinnerOption{
		Action: func() error {
			output, err = uninstall.CombinedOutput()

			return err
		},
	}); err != nil {
		color.Errorf("failed to uninstall package: %s\n", err.Error())
		if len(output) > 0 {
			color.Red().Println(string(output))
		}

		return nil
	}

	// tidy go.mod file
	tidy := exec.Command("go", "mod", "tidy")
	if err = ctx.Spinner(fmt.Sprintf("> @%s", strings.Join(tidy.Args, " ")), console.SpinnerOption{
		Action: func() error {
			output, err = tidy.CombinedOutput()

			return err
		},
	}); err != nil {
		color.Errorf("failed to tidy go.mod file: %s\n", err.Error())
		if len(output) > 0 {
			color.Red().Println(string(output))
		}

		return nil
	}

	color.Successf("Package %s uninstalled successfully\n", pkg)

	return nil
}
