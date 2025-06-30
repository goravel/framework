package console

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/maps"
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

	return r.uninstallPackage(ctx, pkg)

	// TODO: Implement this in v1.17 https://github.com/goravel/goravel/issues/719
	// if isPackage(pkg) {
	// 	return r.uninstallPackage(ctx, pkg)
	// }

	// return r.uninstallFacade(ctx, pkg)
}

func (r *PackageUninstallCommand) uninstallPackage(ctx console.Context, pkg string) error {
	pkgPath, _, _ := strings.Cut(pkg, "@")
	setup := pkgPath + "/setup"

	// uninstall package
	uninstall := exec.Command("go", "run", setup, "uninstall")
	if ctx.OptionBool("force") {
		uninstall.Args = append(uninstall.Args, "--force")
	}

	if err := supportconsole.ExecuteCommand(ctx, uninstall); err != nil {
		color.Errorln("failed to uninstall package:")
		color.Red().Println(err.Error())

		return nil
	}

	// tidy go.mod file
	if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "mod", "tidy")); err != nil {
		color.Errorln("failed to tidy go.mod file:")
		color.Red().Println(err.Error())

		return nil
	}

	color.Successf("Package %s uninstalled successfully\n", pkg)

	return nil
}

func (r *PackageUninstallCommand) uninstallFacade(ctx console.Context, facade string) error {
	path, exists := binding.FacadeToPath[facade]
	if !exists {
		ctx.Warning(errors.PackageFacadeNotFound.Args(facade).Error())
		ctx.Info(fmt.Sprintf("Available facades: %s", strings.Join(maps.Keys(binding.FacadeToPath), ", ")))
		return nil
	}

	setup := path + "/setup"

	if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "run", setup, "uninstall")); err != nil {
		color.Red().Println(err.Error())

		return nil
	}

	color.Successf("Facade %s uninstalled successfully\n", facade)

	return nil
}
