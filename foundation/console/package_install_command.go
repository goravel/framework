package console

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/collect"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/maps"
)

type PackageInstallCommand struct {
	facadeDependencies map[string][]string
	facadeToPath       map[string]string
	baseFacades        []string
}

func NewPackageInstallCommand(facadeDependencies map[string][]string, facadeToPath map[string]string, baseFacades []string) *PackageInstallCommand {
	return &PackageInstallCommand{
		facadeDependencies: facadeDependencies,
		facadeToPath:       facadeToPath,
		baseFacades:        baseFacades,
	}
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
	names := ctx.Arguments()
	if len(names) == 0 {
		name, err := ctx.Ask("Enter the package/facade name to install", console.AskOption{
			Description: "If no version is specified, install the latest",
			Placeholder: " E.g example.com/pkg or example.com/pkg@v1.0.0 or Cache",
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

		names = append(names, name)
	}

	for _, name := range names {
		if isPackage(name) {
			if err := r.installPackage(ctx, name); err != nil {
				return err
			}
		} else {
			if err := r.installFacade(ctx, name); err != nil {
				return err
			}
		}
	}

	return nil
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

func (r *PackageInstallCommand) installFacade(ctx console.Context, facadeName string) error {
	facadeDependencies, exists := r.facadeDependencies[facadeName]
	if !exists {
		ctx.Warning(errors.PackageFacadeNotFound.Args(facadeName).Error())
		ctx.Info(fmt.Sprintf("Available facades: %s", strings.Join(maps.Keys(r.facadeDependencies), ", ")))
		return nil
	}

	filterFacadeDependencies := collect.Filter(facadeDependencies, func(facade string, _ int) bool {
		return !slices.Contains(r.baseFacades, facade)
	})
	ctx.Info(fmt.Sprintf("%s depends on %s, they will be installed simultaneously", facadeName, strings.Join(filterFacadeDependencies, ", ")))

	allFacades := append(facadeDependencies, facadeName)

	for _, facade := range allFacades {
		if slices.Contains(r.baseFacades, facade) {
			continue
		}

		setup := r.facadeToPath[facade] + "/setup"

		if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "run", setup, "install")); err != nil {
			ctx.Error(fmt.Sprintf("Failed to install facade %s, error: %s", facade, err.Error()))

			return nil
		}

		color.Successf("Facade %s installed successfully\n", facade)
	}

	return nil
}

func isPackage(pkg string) bool {
	return strings.Contains(pkg, "/")
}
