package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	stubs := Stubs{}
	viteServiceProvider := "&vite.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	configPath := path.Config("vite.go")
	viteFacadePath := path.Facade("vite.go")
	facadesPackage := setup.Paths().Facades().Package()

	setup.Install(
		// Add the vite service provider to the providers array in bootstrap/providers.go
		modify.RegisterProvider(moduleImport, viteServiceProvider),

		// Create config/vite.go
		modify.File(configPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Facades().Import(), facadesPackage)),

		// Add the Vite facade
		modify.File(viteFacadePath).Overwrite(stubs.ViteFacade(facadesPackage)),
	).Uninstall(
		// Remove config/vite.go
		modify.File(configPath).Remove(),

		// Remove the vite service provider from the providers array in bootstrap/providers.go
		modify.UnregisterProvider(moduleImport, viteServiceProvider),

		// Remove the Vite facade
		modify.File(viteFacadePath).Remove(),
	).Execute()
}
