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
	aiConfigPath := path.Config("ai.go")
	aiFacadePath := path.Facade("ai.go")
	modulePath := setup.Paths().Module().Import()
	aiServiceProvider := "&ai.ServiceProvider{}"
	facadesPackage := setup.Paths().Facades().Package()

	setup.Install(
		// Add the ai service provider to the providers array in bootstrap/providers.go
		modify.RegisterProvider(modulePath, aiServiceProvider),

		// Create config/ai.go
		modify.File(aiConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Facades().Import(), facadesPackage)),

		// Add the AI facade
		modify.File(aiFacadePath).Overwrite(stubs.AIFacade(facadesPackage)),
	).Uninstall(
		// Remove the AI facade
		modify.File(aiFacadePath).Remove(),

		// Remove config/ai.go
		modify.File(aiConfigPath).Remove(),

		// Remove the ai service provider from the providers array in bootstrap/providers.go
		modify.UnregisterProvider(modulePath, aiServiceProvider),
	).Execute()
}
