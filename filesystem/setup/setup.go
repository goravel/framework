package main

import (
	"os"

	"github.com/goravel/framework/contracts/facades"
	contractspackages "github.com/goravel/framework/contracts/packages"
	contractsmodify "github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)

	setup.Install(installActions(setup)...).Uninstall(uninstallActions(setup)...).Execute()
}

func installActions(setup contractspackages.Setup) []contractsmodify.Apply {
	stubs := Stubs{}
	storageConfigPath := path.Config("filesystems.go")
	storageFacadePath := path.Facade("storage.go")
	filesystemServiceProvider := "&filesystem.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	facadesPackage := setup.Paths().Facades().Package()
	envPath := path.Base(".env")
	envExamplePath := path.Base(".env.example")

	return []contractsmodify.Apply{
		// Avoid duplicate installation when installing drivers
		modify.WhenFacade(facades.Storage,
			// Add the filesystem service provider to the providers array in bootstrap/providers.go
			modify.RegisterProvider(moduleImport, filesystemServiceProvider),

			// Create config/filesystems.go
			modify.File(storageConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Facades().Import(), facadesPackage)),

			// Add the Storage facade
			modify.File(storageFacadePath).Overwrite(stubs.StorageFacade(facadesPackage)),

			// Add the filesystem disk to the .env and .env.example files
			modify.WhenFileExists(envPath, modify.EnvWhenMissing(envPath, "FILESYSTEM_DISK", "local")),
			modify.WhenFileExists(envExamplePath, modify.EnvWhenMissing(envExamplePath, "FILESYSTEM_DISK", "local")),
		),

		// Selecting the local driver overwrites the selector, so a stale value cannot survive.
		modify.WhenDriver("local",
			modify.WhenFileExists(envPath, modify.Env(envPath, "FILESYSTEM_DISK", "local")),
			modify.WhenFileExists(envExamplePath, modify.Env(envExamplePath, "FILESYSTEM_DISK", "local")),
		),
	}
}

func uninstallActions(setup contractspackages.Setup) []contractsmodify.Apply {
	storageConfigPath := path.Config("filesystems.go")
	storageFacadePath := path.Facade("storage.go")
	filesystemServiceProvider := "&filesystem.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()

	return []contractsmodify.Apply{
		modify.WhenFacade(facades.Storage,
			// Remove config/filesystems.go
			modify.File(storageConfigPath).Remove(),

			// Remove the filesystem service provider from the providers array in bootstrap/providers.go
			modify.UnregisterProvider(moduleImport, filesystemServiceProvider),

			// Remove the Storage facade
			modify.File(storageFacadePath).Remove(),
		),
	}
}
