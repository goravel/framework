package console

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/convert"
	"github.com/goravel/framework/support/file"
)

type PackageUninstallCommand struct {
	app               foundation.Application
	bindings          map[string]binding.Info
	installedBindings []any
}

func NewPackageUninstallCommand(
	app foundation.Application,
	bindings map[string]binding.Info,
	installedBindings []any,
) *PackageUninstallCommand {
	return &PackageUninstallCommand{
		app:               app,
		bindings:          bindings,
		installedBindings: installedBindings,
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
	binding := convert.FacadeToBinding(name)
	if r.bindings[binding].IsBase {
		ctx.Warning(fmt.Sprintf("Facade %s is a base facade, cannot be uninstalled", name))
		return nil
	}

	bindingInfo, exists := r.bindings[binding]
	if !exists {
		ctx.Warning(errors.PackageFacadeNotFound.Args(name).Error())
		ctx.Info(fmt.Sprintf("Available facades: %s", strings.Join(getAvailableFacades(r.bindings), ", ")))
		return nil
	}

	var bindingAny any = binding
	if !slices.Contains(r.installedBindings, bindingAny) {
		ctx.Warning(fmt.Sprintf("Facade %s is not installed", name))
		return nil
	}

	existingUpperDependencyFacades := r.getExistingUpperDependencyFacades(binding)

	if len(existingUpperDependencyFacades) > 0 {
		ctx.Error(fmt.Sprintf("Facade %s is depended on %s facades, cannot be uninstalled", name, strings.Join(existingUpperDependencyFacades, ", ")))
		return nil
	}

	force := ctx.OptionBool("force")
	setup := bindingInfo.PkgPath + "/setup"
	facade := convert.BindingToFacade(binding)

	uninstall := exec.Command("go", "run", setup, "uninstall")
	uninstall.Args = append(uninstall.Args, "--facade="+facade)

	if force {
		uninstall.Args = append(uninstall.Args, "--force")
	}

	if err := supportconsole.ExecuteCommand(ctx, uninstall); err != nil {
		ctx.Error(fmt.Sprintf("Failed to uninstall facade %s, error: %s", facade, err.Error()))

		return nil
	}

	ctx.Success(fmt.Sprintf("Facade %s uninstalled successfully", facade))

	if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "mod", "tidy")); err != nil {
		ctx.Error(fmt.Sprintf("failed to tidy go.mod file: %s", err))

		return nil
	}

	return nil
}

func (r *PackageUninstallCommand) getBindingsThatNeedUninstall(binding string) []string {
	var facadeDependentCount = make(map[string]int)
	for _, installedBinding := range r.installedBindings {
		installedBindingStr, ok := installedBinding.(string)
		if !ok {
			continue
		}

		for _, dependency := range getDependencyBindings(installedBindingStr, r.bindings) {
			facadeDependentCount[dependency]++
		}
	}

	var needUninstallBindings []string
	for _, dependency := range getDependencyBindings(binding, r.bindings) {
		if facadeDependentCount[dependency] == 1 {
			needUninstallBindings = append(needUninstallBindings, dependency)
		}
	}

	if facadeDependentCount[binding] == 0 {
		needUninstallBindings = append(needUninstallBindings, binding)
	}

	return needUninstallBindings
}

func (r *PackageUninstallCommand) getExistingUpperDependencyFacades(binding string) []string {
	var facades []string
	for _, installedBinding := range r.installedBindings {
		installedBindingStr, ok := installedBinding.(string)
		if !ok {
			continue
		}

		for _, dependency := range getDependencyBindings(installedBindingStr, r.bindings) {
			facade := convert.BindingToFacade(installedBindingStr)

			if dependency == binding && file.Exists(r.app.FacadesPath(fmt.Sprintf("%s.go", strings.ToLower(facade)))) {
				facades = append(facades, convert.BindingToFacade(installedBindingStr))
			}
		}
	}
	return facades
}
