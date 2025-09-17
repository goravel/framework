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
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/support/collect"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/convert"
)

type PackageInstallCommand struct {
	bindings                            map[string]binding.Info
	installedBindings                   []any
	installedFacadesInTheCurrentCommand []string
}

func NewPackageInstallCommand(bindings map[string]binding.Info, installedBindings []any) *PackageInstallCommand {
	return &PackageInstallCommand{
		bindings:          bindings,
		installedBindings: installedBindings,
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
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "all-facades",
				Usage:   "Install all facades",
				Aliases: []string{"a"},
				Value:   false,
			},
		},
	}
}

// Handle Execute the console command.
func (r *PackageInstallCommand) Handle(ctx console.Context) error {
	names := ctx.Arguments()

	if len(names) == 0 {
		if ctx.OptionBool("all-facades") {
			names = getAvailableFacades(r.bindings)
		} else {
			var options []console.Choice
			for _, facade := range getAvailableFacades(r.bindings) {
				options = append(options, console.Choice{
					Key:   facade,
					Value: facade,
				})
			}

			name, err := ctx.MultiSelect("Select the facades to install", options, console.MultiSelectOption{
				Description: "If you want to install a package, please run the command with the package name.\nIf you want to install all facades, you can run the command with --all-facades or click ctrl+a to select all.",
				Filterable:  true,
			})
			if err != nil {
				ctx.Error(err.Error())
				return nil
			}

			if len(name) > 0 {
				names = append(names, name...)
			}
		}
	}

	for _, name := range names {
		if isPackage(name) {
			if err := r.installPackage(ctx, name); err != nil {
				return err
			}
		} else {
			if slices.Contains(r.installedFacadesInTheCurrentCommand, name) {
				continue
			}

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
		ctx.Error(fmt.Sprintf("Failed to get package: %s", err))

		return nil
	}

	// install package
	if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "run", setup, "install")); err != nil {
		ctx.Error(fmt.Sprintf("Failed to install package: %s", err))

		return nil
	}

	// tidy go.mod file
	if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "mod", "tidy")); err != nil {
		ctx.Error(fmt.Sprintf("Failed to tidy go.mod file: %s", err))

		return nil
	}

	ctx.Success(fmt.Sprintf("Package %s installed successfully", pkg))

	return nil
}

func (r *PackageInstallCommand) installFacade(ctx console.Context, name string) error {
	binding := convert.FacadeToBinding(name)
	if _, exists := r.bindings[binding]; !exists {
		ctx.Warning(errors.PackageFacadeNotFound.Args(name).Error())
		ctx.Info(fmt.Sprintf("Available facades: %s", strings.Join(getAvailableFacades(r.bindings), ", ")))
		return nil
	}

	dependencies := r.getDependenciesThatNeedInstall(binding)
	if len(dependencies) > 0 && !ctx.OptionBool("all-facades") {
		facades := make([]string, len(dependencies))
		for i := range dependencies {
			facades[i] = convert.BindingToFacade(dependencies[i])
		}
		ctx.Info(fmt.Sprintf("%s depends on %s, they will be installed simultaneously", name, strings.Join(facades, ", ")))
	}

	dependencies = append(dependencies, binding)
	for _, binding := range dependencies {
		setup := r.bindings[binding].PkgPath + "/setup"
		facade := convert.BindingToFacade(binding)

		if slices.Contains(r.installedFacadesInTheCurrentCommand, facade) {
			continue
		}

		if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "run", setup, "install", "--facade="+facade, "--module="+packages.GetModuleName())); err != nil {
			ctx.Error(fmt.Sprintf("Failed to install facade %s: %s", facade, err.Error()))

			return nil
		}

		r.installedFacadesInTheCurrentCommand = append(r.installedFacadesInTheCurrentCommand, facade)

		ctx.Success(fmt.Sprintf("Facade %s installed successfully", facade))
	}

	if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "mod", "tidy")); err != nil {
		ctx.Error(fmt.Sprintf("Failed to tidy go.mod file: %s", err))
	}

	return nil
}

func (r *PackageInstallCommand) getDependenciesThatNeedInstall(binding string) (needInstall []string) {
	for _, dependencyBinding := range getDependencyBindings(binding, r.bindings) {
		var binding any = dependencyBinding
		if !slices.Contains(r.installedBindings, binding) {
			needInstall = append(needInstall, dependencyBinding)
		}
	}

	return
}

func getAvailableFacades(bindings map[string]binding.Info) []string {
	var result []string
	for binding, info := range bindings {
		if !info.IsBase {
			result = append(result, convert.BindingToFacade(binding))
		}
	}

	slices.Sort(result)

	return result
}

func getDependencyBindings(binding string, bindings map[string]binding.Info) []string {
	var deps []string
	for _, dep := range bindings[binding].Dependencies {
		if info, ok := bindings[dep]; ok && !info.IsBase {
			deps = append(deps, dep)
			deps = append(deps, getDependencyBindings(dep, bindings)...)
		}
	}

	return collect.Unique(deps)
}

func isPackage(pkg string) bool {
	return strings.Contains(pkg, "/")
}
