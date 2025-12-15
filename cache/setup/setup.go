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
	cacheConfigPath := path.Config("cache.go")
	cacheFacadePath := path.Facades("cache.go")
	cacheServiceProvider := "&cache.ServiceProvider{}"
	moduleImport := setup.Paths().Module().Import()
	facadesPackage := setup.Paths().Facades().Package()

	setup.Install(
		// Add the cache service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(moduleImport, cacheServiceProvider),

		// Create config/cache.go
		modify.File(cacheConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Facades().Import(), facadesPackage)),

		// Add the Cache facade
		modify.File(cacheFacadePath).Overwrite(stubs.CacheFacade(facadesPackage)),
	).Uninstall(
		// Remove config/cache.go
		modify.File(cacheConfigPath).Remove(),

		// Remove the cache service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(moduleImport, cacheServiceProvider),

		// Remove the Cache facade
		modify.File(cacheFacadePath).Remove(),
	).Execute()
}
