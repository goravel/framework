package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	modulePath := packages.GetModulePath()
	hashServiceProvider := "&hash.ServiceProvider{}"
	configPath := path.Config("hashing.go")
	hashFacade := "Hash"
	hashFacadePath := path.Facades("hash.go")

	packages.Setup(os.Args).
		Install(
			// Add the hash service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, hashServiceProvider),

			// Create config/hashing.go
			modify.File(configPath).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),

			// Add the Hash facade
			modify.WhenFacade(hashFacade, modify.File(hashFacadePath).Overwrite(stubs.HashFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{hashFacade},
				// Remove config/hashing.go
				modify.File(configPath).Remove(),

				// Remove the hash service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, hashServiceProvider),
			),

			// Remove the Hash facade
			modify.WhenFacade(hashFacade, modify.File(hashFacadePath).Remove()),
		).
		Execute()
}
