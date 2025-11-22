package main

import (
	"os"

	"github.com/goravel/framework/contracts/facades"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	authConfigPath := path.Config("auth.go")
	authFacadePath := path.Facades("auth.go")
	gateFacadePath := path.Facades("gate.go")
	modulePath := packages.GetModulePath()
	authServiceProvider := "&auth.ServiceProvider{}"

	packages.Setup(os.Args).
		Install(
			// Add the auth service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, authServiceProvider),

			// Create config/auth.go
			modify.File(authConfigPath).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),

			// Add the Auth and Gate facades
			modify.WhenFacade(facades.Auth, modify.File(authFacadePath).Overwrite(stubs.AuthFacade())),
			modify.WhenFacade(facades.Gate, modify.File(gateFacadePath).Overwrite(stubs.GateFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Auth, facades.Gate},
				// Remove config/auth.go
				modify.File(authConfigPath).Remove(),

				// Remove the auth service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, authServiceProvider),
			),

			// Remove the Auth and Gate facades
			modify.WhenFacade(facades.Auth, modify.File(authFacadePath).Remove()),
			modify.WhenFacade(facades.Gate, modify.File(gateFacadePath).Remove()),
		).
		Execute()
}
