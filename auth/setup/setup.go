package main

import (
	"os"

	"github.com/goravel/framework/contracts/facades"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	stubs := Stubs{}
	authConfigPath := path.Config("auth.go")
	authFacadePath := path.Facade("auth.go")
	gateFacadePath := path.Facade("gate.go")
	modulePath := setup.Paths().Module().Import()
	authServiceProvider := "&auth.ServiceProvider{}"
	facadesPackage := setup.Paths().Facades().Package()

	setup.Install(
		// Add the auth service provider to the providers array in bootstrap/providers.go
		modify.WhenFileNotContains(path.Bootstrap("providers.go"), authServiceProvider, modify.RegisterProvider(modulePath, authServiceProvider)),

		// Create config/auth.go
		modify.WhenFileNotExists(authConfigPath, modify.File(authConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Facades().Import(), facadesPackage))),

		// Add the Auth and Gate facades
		modify.WhenFacade(facades.Auth, modify.File(authFacadePath).Overwrite(stubs.AuthFacade(facadesPackage))),
		modify.WhenFacade(facades.Gate, modify.File(gateFacadePath).Overwrite(stubs.GateFacade(facadesPackage))),
	).Uninstall(
		modify.WhenNoFacades([]string{facades.Auth, facades.Gate},
			// Remove config/auth.go
			modify.File(authConfigPath).Remove(),

			// Remove the auth service provider from the providers array in bootstrap/providers.go
			modify.UnregisterProvider(modulePath, authServiceProvider),
		),

		// Remove the Auth and Gate facades
		modify.WhenFacade(facades.Auth, modify.File(authFacadePath).Remove()),
		modify.WhenFacade(facades.Gate, modify.File(gateFacadePath).Remove()),
	).Execute()
}
