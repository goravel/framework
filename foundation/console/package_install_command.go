package console

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/collect"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/str"
)

type PackageInstallCommand struct {
	facades          map[string]binding.FacadeInfo
	installedFacades []string
}

func NewPackageInstallCommand(facades map[string]binding.FacadeInfo, installedFacades []string) *PackageInstallCommand {
	return &PackageInstallCommand{
		facades:          facades,
		installedFacades: installedFacades,
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
			Prompt:      "> ",
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
	bindingName := convertFacadeToBinding(name)
	if _, exists := r.facades[bindingName]; !exists {
		ctx.Warning(errors.PackageFacadeNotFound.Args(name).Error())
		ctx.Info(fmt.Sprintf("Available facades: %s", strings.Join(getAvailableFacades(r.facades), ", ")))
		return nil
	}

	dependencies := r.getDependenciesThatNeedInstall(bindingName)
	if len(dependencies) > 0 {
		facadeNames := make([]string, len(dependencies))
		for i := range dependencies {
			facadeNames[i] = convertBindingToFacade(dependencies[i])
		}
		ctx.Info(fmt.Sprintf("%s depends on %s, they will be installed simultaneously", name, strings.Join(facadeNames, ", ")))
	}

	dependencies = append(dependencies, bindingName)
	for _, facade := range dependencies {
		setup := r.facades[facade].PkgPath + "/setup"

		if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "run", setup, "install")); err != nil {
			ctx.Error(fmt.Sprintf("Failed to install facade %s, error: %s", convertBindingToFacade(facade), err.Error()))

			return nil
		}

		ctx.Success(fmt.Sprintf("Facade %s installed successfully", convertBindingToFacade(facade)))
	}

	return nil
}

func (r *PackageInstallCommand) getDependenciesThatNeedInstall(name string) (needInstall []string) {
	for _, dep := range getFacadeDependencies(name, r.facades) {
		if !slices.Contains(r.installedFacades, dep) {
			needInstall = append(needInstall, dep)
		}
	}

	return
}

func convertBindingToFacade(b string) string {
	return str.Of(b).After("goravel.").Studly().WhenIs("Db", func(s *str.String) *str.String {
		return s.Upper()
	}).String()
}

func convertFacadeToBinding(f string) string {
	return "goravel." + str.Of(f).Snake().String()
}

func getAvailableFacades(facades map[string]binding.FacadeInfo) []string {
	var result []string
	for name, info := range facades {
		if !info.IsBase {
			result = append(result, convertBindingToFacade(name))
		}
	}

	slices.Sort(result)

	return result
}

func getFacadeDependencies(name string, facades map[string]binding.FacadeInfo) []string {
	var deps []string
	for _, dep := range facades[name].Dependencies {
		if info, ok := facades[dep]; ok && !info.IsBase {
			deps = append(deps, dep)
			deps = append(deps, getFacadeDependencies(dep, facades)...)
		}
	}

	return collect.Unique(deps)
}

func isPackage(pkg string) bool {
	return strings.Contains(pkg, "/")
}
