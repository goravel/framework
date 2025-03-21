package console

import (
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

	pkg := ctx.Argument(0)
	if pkg == "" {
		var err error
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

	pkgPath, _, _ := strings.Cut(pkg, "@")
	setup := pkgPath + "/setup"

	// uninstall package
	uninstall := exec.Command("go", "run", setup, "uninstall")
	if ctx.OptionBool("force") {
		uninstall.Args = append(uninstall.Args, "--force")
	}

	if err, msg := execute(ctx, uninstall); err != nil {
		color.Errorf("failed to uninstall package: %s\n", err.Error())
		color.Red().Println(msg)

		return nil
	}

	// tidy go.mod file
	if err, msg := execute(ctx, exec.Command("go", "mod", "tidy")); err != nil {
		color.Errorf("failed to tidy go.mod file: %s\n", err.Error())
		color.Red().Println(msg)

		return nil
	}

	color.Successf("Package %s uninstalled successfully\n", pkg)

	return nil
}
