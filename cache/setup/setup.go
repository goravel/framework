package main

import (
	"os"

	"github.com/goravel/framework/contracts/facades"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	providersBootstrapPath := path.Bootstrap("providers.go")
	cacheConfigPath := path.Config("cache.go")
	cacheFacadePath := path.Facades("cache.go")
	cacheServiceProvider := "&cache.ServiceProvider{}"

	packages.Setup(os.Args).
		Install(
			modify.GoFile(providersBootstrapPath).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register(cacheServiceProvider)),
			modify.File(cacheConfigPath).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.WhenFacade(facades.Cache, modify.File(cacheFacadePath).Overwrite(stubs.CacheFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Cache},
				modify.GoFile(providersBootstrapPath).
					Find(match.Providers()).Modify(modify.Unregister(cacheServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.File(cacheConfigPath).Remove(),
			),
			modify.WhenFacade(facades.Cache, modify.File(cacheFacadePath).Remove()),
		).
		Execute()
}
