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
	processFacadePath := path.Facades("process.go")
	moduleImport := setup.Paths().Module().Import()
	processServiceProvider := "&process.ServiceProvider{}"

	setup.Install(
		// Add the process service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(moduleImport, processServiceProvider),

		// Add the Process facade
		modify.File(processFacadePath).Overwrite(stubs.ProcessFacade(setup.Paths().Facades().Package())),
	).Uninstall(
		// Remove the process service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(moduleImport, processServiceProvider),

		// Remove the Process facade
		modify.File(processFacadePath).Remove(),
	).Execute()
}
