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
	storageConfigPath := path.Config("filesystems.go")
	storageFacadePath := path.Facades("storage.go")
	filesystemServiceProvider := "&filesystem.ServiceProvider{}"
	modulePath := packages.GetModulePath()

	packages.Setup(os.Args).
		Install(
			// Add the filesystem service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, filesystemServiceProvider),

			// Create config/filesystems.go
			modify.File(storageConfigPath).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),

			// Add the Storage facade
			modify.WhenFacade(facades.Storage, modify.File(storageFacadePath).Overwrite(stubs.StorageFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Storage},
				// Remove config/filesystems.go
				modify.File(storageConfigPath).Remove(),

				// Remove the filesystem service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, filesystemServiceProvider),
			),

			// Remove the Storage facade
			modify.WhenFacade(facades.Storage, modify.File(storageFacadePath).Remove()),
		).
		Execute()
}
