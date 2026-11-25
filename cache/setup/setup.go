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
	cacheConfigPath := path.Config("cache.go")
	cacheFacadePath := path.Facade("cache.go")
	cacheServiceProvider := "&cache.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	facadesPackage := setup.Paths().Facades().Package()
	envPath := path.Base(".env")
	envExamplePath := path.Base(".env.example")

	return []contractsmodify.Apply{
		// Avoid duplicate installation when installing drivers
		modify.WhenFacade(facades.Cache,
			// Add the cache service provider to the providers array in bootstrap/providers.go
			modify.RegisterProvider(moduleImport, cacheServiceProvider),

			// Create config/cache.go
			modify.File(cacheConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Facades().Import(), facadesPackage)),

			// Add the Cache facade
			modify.File(cacheFacadePath).Overwrite(stubs.CacheFacade(facadesPackage)),

			// Add the cache store to the .env and .env.example files
			modify.WhenFileExists(envPath, modify.EnvWhenMissing(envPath, "CACHE_STORE", "memory")),
			modify.WhenFileExists(envExamplePath, modify.EnvWhenMissing(envExamplePath, "CACHE_STORE", "memory")),
		),

		// Selecting the memory driver overwrites the selector, so a stale value cannot survive.
		modify.WhenDriver("memory",
			modify.WhenFileExists(envPath, modify.Env(envPath, "CACHE_STORE", "memory")),
			modify.WhenFileExists(envExamplePath, modify.Env(envExamplePath, "CACHE_STORE", "memory")),
		),
	}
}

func uninstallActions(setup contractspackages.Setup) []contractsmodify.Apply {
	cacheConfigPath := path.Config("cache.go")
	cacheFacadePath := path.Facade("cache.go")
	cacheServiceProvider := "&cache.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()

	return []contractsmodify.Apply{
		modify.WhenFacade(facades.Cache,
			// Remove config/cache.go
			modify.File(cacheConfigPath).Remove(),

			// Remove the cache service provider from the providers array in bootstrap/providers.go
			modify.UnregisterProvider(moduleImport, cacheServiceProvider),

			// Remove the Cache facade
			modify.File(cacheFacadePath).Remove(),
		),
	}
}
