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
	langFacadePath := path.Facade("lang.go")
	translationServiceProvider := "&translation.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()

	setup.Install(
		// Add the translation service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(moduleImport, translationServiceProvider),

		// Add the Lang facade
		modify.File(langFacadePath).Overwrite(stubs.LangFacade(setup.Paths().Facades().Package())),
	).Uninstall(
		// Remove the translation service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(moduleImport, translationServiceProvider),

		// Remove the Lang facade
		modify.File(langFacadePath).Remove(),
	).Execute()
}
