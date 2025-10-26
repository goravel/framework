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
	langFacadePath := path.Facades("lang.go")
	langServiceProvider := "&translation.ServiceProvider{}"

	packages.Setup(os.Args).
		Install(
			modify.GoFile(providersBootstrapPath).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register(langServiceProvider)),
			modify.WhenFacade(facades.Lang, modify.File(path.Facades(langFacadePath)).Overwrite(stubs.LangFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Lang},
				modify.GoFile(providersBootstrapPath).
					Find(match.Providers()).Modify(modify.Unregister(langServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			),
			modify.WhenFacade(facades.Lang, modify.File(path.Facades(langFacadePath)).Remove()),
		).
		Execute()
}
