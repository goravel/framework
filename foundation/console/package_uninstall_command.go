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
	"github.com/goravel/framework/support"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/convert"
	"github.com/goravel/framework/support/env"
)

type PackageUninstallCommand struct {
	bindings map[string]binding.Info
	paths    string
}

func NewPackageUninstallCommand(
	bindings map[string]binding.Info,
	json foundation.Json,
) *PackageUninstallCommand {
	paths, err := json.MarshalString(support.Config.Paths)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal paths: %s", err))
	}

	return &PackageUninstallCommand{
		bindings: bindings,
		paths:    paths,
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
	uninstall := exec.Command("go", "run", setup, "uninstall", "--main-path="+env.MainPath(), "--paths="+r.paths)
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
	bindingInfo, exists := r.bindings[binding]
	if !exists {
		ctx.Warning(errors.PackageFacadeNotFound.Args(name).Error())
		ctx.Info(fmt.Sprintf("Available facades: %s", strings.Join(getAvailableFacades(r.bindings), ", ")))
		return nil
	}

	if r.bindings[binding].IsBase {
		ctx.Warning(fmt.Sprintf("Facade %s is a base facade, cannot be uninstalled", name))
		return nil
	}

	if !doesFacadeExist(name) {
		ctx.Warning(fmt.Sprintf("Facade %s is not installed", name))
		return nil
	}

	existingUpperDependencyFacades := r.getExistingUpperDependencyFacades(name)
	if len(existingUpperDependencyFacades) > 0 {
		ctx.Error(fmt.Sprintf("Facade %s is depended on %s facades, cannot be uninstalled", name, strings.Join(existingUpperDependencyFacades, ", ")))
		return nil
	}

	force := ctx.OptionBool("force")
	setup := bindingInfo.PkgPath + "/setup"
	facade := convert.BindingToFacade(binding)
	uninstall := exec.Command("go", "run", setup, "uninstall", "--facade="+facade, "--main-path="+env.MainPath(), "--paths="+r.paths)

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

func (r *PackageUninstallCommand) getExistingUpperDependencyFacades(facade string) []string {
	var facades []string
	binding := convert.FacadeToBinding(facade)
	for bindingItem, info := range r.bindings {
		facadeItem := convert.BindingToFacade(bindingItem)
		if slices.Contains(info.Dependencies, binding) && doesFacadeExist(facadeItem) {
			facades = append(facades, facadeItem)
		}
	}

	return facades
}
