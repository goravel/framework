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
	modulePath := setup.ModulePath()
	processServiceProvider := "&process.ServiceProvider{}"

	setup.Install(
		// Add the process service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(modulePath, processServiceProvider),

		// Add the Process facade
		modify.File(processFacadePath).Overwrite(stubs.ProcessFacade()),
	).Uninstall(
		// Remove the process service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(modulePath, processServiceProvider),

		// Remove the Process facade
		modify.File(processFacadePath).Remove(),
	).Execute()
}
