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
	processFacadePath := path.Facades("process.go")
	modulePath := packages.GetModulePath()
	processServiceProvider := "&process.ServiceProvider{}"

	packages.Setup(os.Args).
		Install(
			// Add the process service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, processServiceProvider),

			// Add the Process facade
			modify.WhenFacade(facades.Process, modify.File(processFacadePath).Overwrite(stubs.ProcessFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Process},
				// Remove the process service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, processServiceProvider),
			),

			// Remove the Process facade
			modify.WhenFacade(facades.Process, modify.File(processFacadePath).Remove()),
		).
		Execute()
}
