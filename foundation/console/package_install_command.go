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
	"github.com/goravel/framework/support/color"
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
			var err error

			options := []console.Choice{
				{Key: "All facades", Value: "all"},
				{Key: "Select facades", Value: "select"},
				{Key: "Third package", Value: "third"},
			}

			choice, err := ctx.Choice("Which facades or package do you want to install?", options)
			if err != nil {
				ctx.Error(err.Error())
				return nil
			}

			if choice == "all" {
				names = getAvailableFacades(r.bindings)
			}

			if choice == "select" {
				names, err = r.selectFacades(ctx)
			}

			if choice == "third" {
				var name string
				name, err = r.inputThirdPackage(ctx)
				if name != "" {
					names = []string{name}
				}
			}

			if err != nil {
				ctx.Error(err.Error())
				return nil
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

func (r *PackageInstallCommand) selectFacades(ctx console.Context) ([]string, error) {
	var facadeOptions []console.Choice
	for _, facade := range getAvailableFacades(r.bindings) {
		key := facade
		description := getFacadeDescription(facade, r.bindings)
		if description != "" {
			key = fmt.Sprintf("%-11s", facade) + color.Gray().Sprintf(" - %s", description)
		}
		facadeOptions = append(facadeOptions, console.Choice{
			Key:   key,
			Value: facade,
		})
	}

	return ctx.MultiSelect("Select the facades to install", facadeOptions, console.MultiSelectOption{
		Filterable: true,
	})
}

func (r *PackageInstallCommand) inputThirdPackage(ctx console.Context) (string, error) {
	return ctx.Ask("Enter the package", console.AskOption{
		Description: "E.g.: github.com/goravel/framework or github.com/goravel/framework@master",
	})
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

	bindingsToInstall := r.getBindingsToInstall(binding)
	if len(bindingsToInstall) > 0 && !ctx.OptionBool("all-facades") {
		facades := make([]string, len(bindingsToInstall))
		for i := range bindingsToInstall {
			facades[i] = convert.BindingToFacade(bindingsToInstall[i])
		}
		ctx.Info(fmt.Sprintf("%s depends on %s, they will be installed simultaneously", name, strings.Join(facades, ", ")))
	}

	bindingsToInstall = append(bindingsToInstall, binding)
	for _, binding := range bindingsToInstall {
		bindingInfo := r.bindings[binding]
		setup := bindingInfo.PkgPath + "/setup"
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

		if err := r.installDriver(ctx, facade, bindingInfo); err != nil {
			return err
		}
	}

	if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "mod", "tidy")); err != nil {
		ctx.Error(fmt.Sprintf("Failed to tidy go.mod file: %s", err))
	}

	return nil
}

func (r *PackageInstallCommand) installDriver(ctx console.Context, facade string, bindingInfo binding.Info) error {
	if len(bindingInfo.Drivers) == 0 {
		return nil
	}

	var options []console.Choice
	for _, driver := range bindingInfo.Drivers {
		key := driver.Name
		if driver.Description != "" {
			key += color.Gray().Sprintf(" - %s", driver.Description)
		}

		options = append(options, console.Choice{
			Key:   key,
			Value: driver.Package,
		})
	}

	options = append(options, console.Choice{
		Key:   "Custom",
		Value: "Custom",
	})

	driver, err := ctx.Choice(fmt.Sprintf("Select the %s driver to install", facade), options, console.ChoiceOption{
		Description: fmt.Sprintf("A driver is required for %s, please select one to install.", facade),
	})
	if err != nil {
		return err
	}

	if driver == "Custom" {
		driver, err = ctx.Ask(fmt.Sprintf("Please enter the %s driver package", facade))
		if err != nil {
			return err
		}
	}

	if driver == "" {
		return r.installDriver(ctx, facade, bindingInfo)
	} else {
		return r.installPackage(ctx, driver)
	}
}

func (r *PackageInstallCommand) getBindingsToInstall(binding string) (bindingsToInstall []string) {
	for _, dependencyBinding := range getDependencyBindings(binding, r.bindings) {
		var binding any = dependencyBinding
		if !slices.Contains(r.installedBindings, binding) {
			bindingsToInstall = append(bindingsToInstall, dependencyBinding)
		}
	}

	InstallTogether := r.bindings[binding].InstallTogether
	for _, installTogetherBinding := range InstallTogether {
		var binding any = installTogetherBinding
		if !slices.Contains(r.installedBindings, binding) && !slices.Contains(bindingsToInstall, installTogetherBinding) {
			bindingsToInstall = append(bindingsToInstall, installTogetherBinding)
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

func getFacadeDescription(facade string, bindings map[string]binding.Info) string {
	binding := convert.FacadeToBinding(facade)
	if info, exists := bindings[binding]; exists {
		return info.Description
	}

	return ""
}

func isPackage(pkg string) bool {
	return strings.Contains(pkg, "/")
}
