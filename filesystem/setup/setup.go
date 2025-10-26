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
	storageConfigPath := path.Config("filesystems.go")
	storageFacadePath := path.Facades("storage.go")
	storageServiceProvider := "&filesystem.ServiceProvider{}"

	packages.Setup(os.Args).
		Install(
			modify.GoFile(providersBootstrapPath).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register(storageServiceProvider)),
			modify.File(storageConfigPath).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.WhenFacade(facades.Storage, modify.File(storageFacadePath).Overwrite(stubs.StorageFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Storage},
				modify.GoFile(providersBootstrapPath).
					Find(match.Providers()).Modify(modify.Unregister(storageServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.File(storageConfigPath).Remove(),
			),
			modify.WhenFacade(facades.Storage, modify.File(storageFacadePath).Remove()),
		).
		Execute()
}
