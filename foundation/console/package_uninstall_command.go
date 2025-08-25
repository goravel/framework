package console

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/maps"
)

type PackageUninstallCommand struct {
	baseFacades        []string
	facadeDependencies map[string][]string
	facadeToPath       map[string]string
	installedFacades   []string
}

func NewPackageUninstallCommand(
	facadeDependencies map[string][]string,
	facadeToPath map[string]string,
	baseFacades []string,
	installedFacades []string,
) *PackageUninstallCommand {
	return &PackageUninstallCommand{
		facadeDependencies: facadeDependencies,
		facadeToPath:       facadeToPath,
		baseFacades:        baseFacades,
		installedFacades:   installedFacades,
	}
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
	names := ctx.Arguments()
	if len(names) == 0 {
		var err error
		name, err := ctx.Ask("Enter the package name to uninstall", console.AskOption{
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

func (r *PackageUninstallCommand) uninstallFacade(ctx console.Context, name string) error {
	if slices.Contains(r.baseFacades, name) {
		ctx.Warning(fmt.Sprintf("Facade %s is a base facade and cannot be uninstalled", name))
		return nil
	}

	_, exists := r.facadeDependencies[name]
	if !exists {
		ctx.Warning(errors.PackageFacadeNotFound.Args(name).Error())
		ctx.Info(fmt.Sprintf("Available facades: %s", strings.Join(filterBaseFacades(maps.Keys(r.facadeDependencies), r.baseFacades), ", ")))
		return nil
	}

	facadesThatNeedUninstall := []string{name}
	dependenciesThatNeedUninstall := r.getDependenciesThatNeedUninstall(name)

	if len(dependenciesThatNeedUninstall) > 0 && ctx.Confirm(fmt.Sprintf("Do you want to remove the dependency facades as well: %s?", strings.Join(dependenciesThatNeedUninstall, ", "))) {
		facadesThatNeedUninstall = append(facadesThatNeedUninstall, dependenciesThatNeedUninstall...)
	}

	for _, facade := range facadesThatNeedUninstall {
		if slices.Contains(r.baseFacades, facade) {
			continue
		}

		setup := r.facadeToPath[facade] + "/setup"

		if err := supportconsole.ExecuteCommand(ctx, exec.Command("go", "run", setup, "uninstall")); err != nil {
			ctx.Error(fmt.Sprintf("Failed to uninstall facade %s, error: %s", facade, err.Error()))

			return nil
		}

		color.Successf("Facade %s uninstalled successfully\n", facade)
	}

	return nil
}

func (r *PackageUninstallCommand) getDependenciesThatNeedUninstall(facade string) []string {
	dependencies := r.facadeDependencies[facade]
	if len(dependencies) == 0 {
		return nil
	}

	facadeToNumber := make(map[string]int)
	for _, installedFacade := range r.installedFacades {
		facadeToNumber[installedFacade]++

		for _, dependency := range r.facadeDependencies[installedFacade] {
			facadeToNumber[dependency]++
		}
	}

	var needUninstallFacades []string
	for _, dependency := range dependencies {
		if facadeToNumber[dependency] == 1 {
			needUninstallFacades = append(needUninstallFacades, dependency)
		}
	}

	return filterBaseFacades(needUninstallFacades, r.baseFacades)
}
