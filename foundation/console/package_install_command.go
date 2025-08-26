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
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/maps"
)

type PackageInstallCommand struct {
	baseFacades        []string
	facadeDependencies map[string][]string
	facadeToPath       map[string]string
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
		ctx.Error(fmt.Sprintf("failed to get package: %s", err))

		return nil
	}

	// install package
	if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "run", setup, "install")); err != nil {
		ctx.Error(fmt.Sprintf("failed to install package: %s", err))

		return nil
	}

	// tidy go.mod file
	if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "mod", "tidy")); err != nil {
		ctx.Error(fmt.Sprintf("failed to tidy go.mod file: %s", err))

		return nil
	}

	ctx.Success(fmt.Sprintf("Package %s installed successfully", pkg))

	return nil
}

func (r *PackageInstallCommand) installFacade(ctx console.Context, name string) error {
	facadeDependencies, exists := r.facadeDependencies[name]
	if !exists {
		ctx.Warning(errors.PackageFacadeNotFound.Args(name).Error())
		ctx.Info(fmt.Sprintf("Available facades: %s", strings.Join(filterBaseFacades(maps.Keys(r.facadeDependencies), r.baseFacades), ", ")))
		return nil
	}

	filterFacadeDependencies := filterBaseFacades(facadeDependencies, r.baseFacades)
	ctx.Info(fmt.Sprintf("%s depends on %s, they will be installed simultaneously", name, strings.Join(filterFacadeDependencies, ", ")))

	facadesThatNeedInstall := append(filterFacadeDependencies, name)

	for _, facade := range facadesThatNeedInstall {
		if slices.Contains(r.baseFacades, facade) {
			continue
		}

		setup := r.facadeToPath[facade] + "/setup"

		if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "run", setup, "install")); err != nil {
			ctx.Error(fmt.Sprintf("Failed to install facade %s, error: %s", facade, err.Error()))

			return nil
		}

		ctx.Success(fmt.Sprintf("Facade %s installed successfully", facade))
	}

	return nil
}

func filterBaseFacades(facades, baseFacades []string) []string {
	return collect.Filter(facades, func(facade string, _ int) bool {
		return !slices.Contains(baseFacades, facade)
	})
}

func isPackage(pkg string) bool {
	return strings.Contains(pkg, "/")
}
