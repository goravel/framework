package main

import (
	"os"

	"github.com/goravel/framework/contracts/facades"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	stubs := Stubs{}
	cacheConfigPath := path.Config("cache.go")
	cacheFacadePath := path.Facade("cache.go")
	cacheServiceProvider := "&cache.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	facadesPackage := setup.Paths().Facades().Package()

	setup.Install(
		// Avoid duplicate installation when installing drivers
		modify.WhenFacade(facades.Cache,
			// Add the cache service provider to the providers array in bootstrap/providers.go
			modify.RegisterProvider(moduleImport, cacheServiceProvider),

			// Create config/cache.go
			modify.File(cacheConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Facades().Import(), facadesPackage)),

			// Add the Cache facade
			modify.File(cacheFacadePath).Overwrite(stubs.CacheFacade(facadesPackage)),
		),
	).Uninstall(
		modify.WhenFacade(facades.Cache,
			// Remove config/cache.go
			modify.File(cacheConfigPath).Remove(),

			// Remove the cache service provider from the providers array in bootstrap/providers.go
			modify.UnregisterProvider(moduleImport, cacheServiceProvider),

			// Remove the Cache facade
			modify.File(cacheFacadePath).Remove(),
		),
	).Execute()
}
