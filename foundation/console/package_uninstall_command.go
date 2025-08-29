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
)

type PackageUninstallCommand struct {
	facades          map[string]binding.FacadeInfo
	installedFacades []string
}

func NewPackageUninstallCommand(
	facades map[string]binding.FacadeInfo,
	installedFacades []string,
) *PackageUninstallCommand {
	return &PackageUninstallCommand{
		facades:          facades,
		installedFacades: installedFacades,
	}
}

// Signature The name and signature of the console command.
func (r *PackageUninstallCommand) Signature() string {
	return "package:uninstall"
}

// Description The console command description.
func (r *PackageUninstallCommand) Description() string {
	return "Uninstall a package or a facade"
}

// Extend The console command extend.
func (r *PackageUninstallCommand) Extend() command.Extend {
	return command.Extend{
		ArgsUsage: " <package@version> or <facade>",
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
	names := ctx.Arguments()
	if len(names) == 0 {
		var err error
		name, err := ctx.Ask("Enter the package name to uninstall", console.AskOption{
			Placeholder: " E.g example.com/pkg",
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
			if err := r.uninstallPackage(ctx, name); err != nil {
				return err
			}
		} else {
			if err := r.uninstallFacade(ctx, name); err != nil {
				return err
			}
		}
	}

	return nil
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
		ctx.Error(fmt.Sprintf("failed to uninstall package: %s", err))

		return nil
	}

	// tidy go.mod file
	if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "mod", "tidy")); err != nil {
		ctx.Error(fmt.Sprintf("failed to tidy go.mod file: %s", err))

		return nil
	}

	ctx.Success(fmt.Sprintf("Package %s uninstalled successfully", pkg))

	return nil
}

func (r *PackageUninstallCommand) uninstallFacade(ctx console.Context, name string) error {
	bindingName := convertFacadeToBinding(name)
	if r.facades[bindingName].IsBase {
		ctx.Warning(fmt.Sprintf("Facade %s is a base facade, cannot be uninstalled", name))
		return nil
	}

	_, exists := r.facades[bindingName]
	if !exists {
		ctx.Warning(errors.PackageFacadeNotFound.Args(name).Error())
		ctx.Info(fmt.Sprintf("Available facades: %s", strings.Join(getAvailableFacades(r.facades), ", ")))
		return nil
	}

	if !slices.Contains(r.installedFacades, bindingName) {
		ctx.Warning(fmt.Sprintf("Facade %s is not installed", name))
		return nil
	}

	facadesThatNeedUninstall := r.getFacadesThatNeedUninstall(bindingName)
	if !slices.Contains(facadesThatNeedUninstall, bindingName) {
		ctx.Error(fmt.Sprintf("Facade %s is depended on by other facades, cannot be uninstalled", name))
		return nil
	}

	dependenciesThatNeedUninstall := collect.Filter(facadesThatNeedUninstall, func(facade string, _ int) bool {
		return facade != bindingName
	})

	if len(dependenciesThatNeedUninstall) > 0 {
		needUninstallFacadeNames := make([]string, len(dependenciesThatNeedUninstall))
		for i := range dependenciesThatNeedUninstall {
			needUninstallFacadeNames[i] = convertBindingToFacade(dependenciesThatNeedUninstall[i])
		}
		if !ctx.Confirm(fmt.Sprintf("Do you want to remove the dependency facades as well: %s?", strings.Join(needUninstallFacadeNames, ", "))) {
			facadesThatNeedUninstall = []string{bindingName}
		}

	}

	for _, facade := range facadesThatNeedUninstall {
		setup := r.facades[facade].PkgPath + "/setup"

		uninstall := exec.Command("go", "run", setup, "uninstall")
		uninstall.Args = append(uninstall.Args, "--facade="+facade)
		if ctx.OptionBool("force") {
			uninstall.Args = append(uninstall.Args, "--force")
		}

		if err := supportconsole.ExecuteCommand(ctx, uninstall); err != nil {
			ctx.Error(fmt.Sprintf("Failed to uninstall facade %s, error: %s", convertBindingToFacade(facade), err.Error()))

			if ctx.OptionBool("force") {
				continue
			}

			return nil
		}

		ctx.Success(fmt.Sprintf("Facade %s uninstalled successfully", convertBindingToFacade(facade)))
	}

	return nil
}

func (r *PackageUninstallCommand) getFacadesThatNeedUninstall(facade string) []string {
	var facadeDependentCount = make(map[string]int)
	for _, installedFacade := range r.installedFacades {
		for _, dependency := range getFacadeDependencies(installedFacade, r.facades) {
			facadeDependentCount[dependency]++
		}
	}

	var needUninstallFacades []string
	for _, dependency := range getFacadeDependencies(facade, r.facades) {
		if facadeDependentCount[dependency] == 1 {
			needUninstallFacades = append(needUninstallFacades, dependency)
		}
	}

	if facadeDependentCount[facade] == 0 {
		needUninstallFacades = append(needUninstallFacades, facade)
	}

	return needUninstallFacades
}
