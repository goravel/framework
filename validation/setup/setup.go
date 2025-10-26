package main

import (
	"os"

	"github.com/goravel/framework/contracts/facades"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	providersBootstrapPath := path.Bootstrap("providers.go")
	validationFacadePath := path.Facades("validation.go")
	validationServiceProvider := "&validation.ServiceProvider{}"

	packages.Setup(os.Args).
		Install(
			modify.GoFile(providersBootstrapPath).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register(validationServiceProvider)),
			modify.WhenFacade(facades.Validation, modify.File(validationFacadePath).Overwrite(stubs.ValidationFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Validation},
				modify.GoFile(providersBootstrapPath).
					Find(match.Providers()).Modify(modify.Unregister(validationServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			),
			modify.WhenFacade(facades.Validation, modify.File(validationFacadePath).Remove()),
		).
		Execute()
}
