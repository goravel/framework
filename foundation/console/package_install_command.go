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

type PackageInstallCommand struct {
}

func NewPackageInstallCommand() *PackageInstallCommand {
	return &PackageInstallCommand{}
}

// Signature The name and signature of the console command.
func (r *PackageInstallCommand) Signature() string {
	return "package:install"
}

// Description The console command description.
func (r *PackageInstallCommand) Description() string {
	return "Install a package or a facade"
}

// Extend The console command extend.
func (r *PackageInstallCommand) Extend() command.Extend {
	return command.Extend{
		ArgsUsage: " <package@version> or <facade>",
		Category:  "package",
	}
}

// Handle Execute the console command.
func (r *PackageInstallCommand) Handle(ctx console.Context) error {
	pkg := ctx.Argument(0)
	if pkg == "" {
		var err error
		pkg, err = ctx.Ask("Enter the package/facade name to install", console.AskOption{
			Description: "If no version is specified, install the latest",
			Placeholder: " E.g example.com/pkg or example.com/pkg@v1.0.0 or cache",
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

	if isPackage(pkg) {
		return r.installPackage(ctx, pkg)
	}

	return r.installFacade(ctx, pkg)
}

func (r *PackageInstallCommand) installPackage(ctx console.Context, pkg string) error {
	pkgPath, _, _ := strings.Cut(pkg, "@")
	setup := pkgPath + "/setup"

	// get package
	if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "get", pkg)); err != nil {
		color.Red().Println(err.Error())

		return nil
	}

	// install package
	if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "run", setup, "install")); err != nil {
		color.Red().Println(err.Error())

		return nil
	}

	// tidy go.mod file
	if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "mod", "tidy")); err != nil {
		color.Red().Println(err.Error())

		return nil
	}

	color.Successf("Package %s installed successfully\n", pkg)

	return nil
}

func (r *PackageInstallCommand) installFacade(ctx console.Context, facade string) error {
	path, exists := binding.FacadeToPath[facade]
	if !exists {
		ctx.Warning(errors.PackageFacadeNotFound.Args(facade).Error())
		ctx.Info(fmt.Sprintf("Available facades: %s", strings.Join(maps.Keys(binding.FacadeToPath), ", ")))
		return nil
	}

	setup := path + "/setup"

	if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "run", setup, "install")); err != nil {
		color.Red().Println(err.Error())

		return nil
	}

	color.Successf("Facade %s installed successfully\n", facade)

	return nil
}

func isPackage(pkg string) bool {
	return strings.Contains(pkg, "/")
}
