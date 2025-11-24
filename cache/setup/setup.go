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
	cacheConfigPath := path.Config("cache.go")
	cacheFacadePath := path.Facades("cache.go")
	cacheServiceProvider := "&cache.ServiceProvider{}"
	modulePath := packages.GetModulePath()

	packages.Setup(os.Args).
		Install(
			// Add the cache service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, cacheServiceProvider),

			// Create config/cache.go
			modify.File(cacheConfigPath).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),

			// Add the Cache facade
			modify.WhenFacade(facades.Cache, modify.File(cacheFacadePath).Overwrite(stubs.CacheFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Cache},
				// Remove config/cache.go
				modify.File(cacheConfigPath).Remove(),

				// Remove the cache service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, cacheServiceProvider),
			),

			// Remove the Cache facade
			modify.WhenFacade(facades.Cache, modify.File(cacheFacadePath).Remove()),
		).
		Execute()
}
