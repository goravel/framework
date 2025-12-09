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
	modulePath := setup.ModulePath()
	hashServiceProvider := "&hash.ServiceProvider{}"
	configPath := path.Config("hashing.go")
	hashFacadePath := path.Facades("hash.go")

	setup.Install(
		// Add the hash service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(modulePath, hashServiceProvider),

		// Create config/hashing.go
		modify.File(configPath).Overwrite(stubs.Config(setup.PackageName())),

		// Add the Hash facade
		modify.File(hashFacadePath).Overwrite(stubs.HashFacade()),
	).Uninstall(
		// Remove config/hashing.go
		modify.File(configPath).Remove(),

		// Remove the hash service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(modulePath, hashServiceProvider),

		// Remove the Hash facade
		modify.File(hashFacadePath).Remove(),
	).Execute()
}
