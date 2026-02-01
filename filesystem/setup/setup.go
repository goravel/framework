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
	storageFacadePath := path.Facade("storage.go")
	filesystemServiceProvider := "&filesystem.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	facadesPackage := setup.Paths().Facades().Package()

	setup.Install(
		// Add the filesystem service provider to the providers array in bootstrap/providers.go
		modify.RegisterProvider(moduleImport, filesystemServiceProvider),

		// Create config/filesystems.go
		modify.File(storageConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Facades().Import(), facadesPackage)),

		// Add the Storage facade
		modify.File(storageFacadePath).Overwrite(stubs.StorageFacade(facadesPackage)),
	).Uninstall(
		// Remove config/filesystems.go
		modify.File(storageConfigPath).Remove(),

		// Remove the filesystem service provider from the providers array in bootstrap/providers.go
		modify.UnregisterProvider(moduleImport, filesystemServiceProvider),

		// Remove the Storage facade
		modify.File(storageFacadePath).Remove(),
	).Execute()
}
