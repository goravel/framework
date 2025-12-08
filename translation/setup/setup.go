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
	langFacadePath := path.Facades("lang.go")
	translationServiceProvider := "&translation.ServiceProvider{}"
	modulePath := setup.ModulePath()

	setup.Install(
		// Add the translation service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(modulePath, translationServiceProvider),

		// Add the Lang facade
		modify.File(langFacadePath).Overwrite(stubs.LangFacade()),
	).Uninstall(
		// Remove the translation service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(modulePath, translationServiceProvider),

		// Remove the Lang facade
		modify.File(langFacadePath).Remove(),
	).Execute()
}
