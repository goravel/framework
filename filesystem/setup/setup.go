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
	storageConfigPath := path.Config("filesystems.go")
	storageFacadePath := path.Facades("storage.go")
	filesystemServiceProvider := "&filesystem.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()

	setup.Install(
		// Add the filesystem service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(moduleImport, filesystemServiceProvider),

		// Create config/filesystems.go
		modify.File(storageConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Main().Package())),

		// Add the Storage facade
		modify.File(storageFacadePath).Overwrite(stubs.StorageFacade(setup.Paths().Facades().Package())),
	).Uninstall(
		// Remove config/filesystems.go
		modify.File(storageConfigPath).Remove(),

		// Remove the filesystem service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(moduleImport, filesystemServiceProvider),

		// Remove the Storage facade
		modify.File(storageFacadePath).Remove(),
	).Execute()
}
