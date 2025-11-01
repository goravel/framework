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
	authConfigPath := path.Config("auth.go")
	authFacadePath := path.Facades("auth.go")
	gateFacadePath := path.Facades("gate.go")
	modulePath := packages.GetModulePath()
	authServiceProvider := "&auth.ServiceProvider{}"

	packages.Setup(os.Args).
		Install(
			modify.GoFile(providersBootstrapPath).
				Find(match.Imports()).Modify(modify.AddImport(modulePath)).
				Find(match.Providers()).Modify(modify.Register(authServiceProvider)),
			modify.File(authConfigPath).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.WhenFacade(facades.Auth, modify.File(authFacadePath).Overwrite(stubs.AuthFacade())),
			modify.WhenFacade(facades.Gate, modify.File(gateFacadePath).Overwrite(stubs.GateFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Auth, facades.Gate},
				modify.GoFile(providersBootstrapPath).
					Find(match.Providers()).Modify(modify.Unregister(authServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(modulePath)),
				modify.File(authConfigPath).Remove(),
			),
			modify.WhenFacade(facades.Auth, modify.File(authFacadePath).Remove()),
			modify.WhenFacade(facades.Gate, modify.File(gateFacadePath).Remove()),
		).
		Execute()
}
