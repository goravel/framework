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
	moduleImport := setup.Paths().Module().Import()
	hashServiceProvider := "&hash.ServiceProvider{}"
	configPath := path.Config("hashing.go")
	hashFacadePath := path.Facades("hash.go")

	setup.Install(
		// Add the hash service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(moduleImport, hashServiceProvider),

		// Create config/hashing.go
		modify.File(configPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Main().Package())),

		// Add the Hash facade
		modify.File(hashFacadePath).Overwrite(stubs.HashFacade(setup.Paths().Facades().Package())),
	).Uninstall(
		// Remove config/hashing.go
		modify.File(configPath).Remove(),

		// Remove the hash service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(moduleImport, hashServiceProvider),

		// Remove the Hash facade
		modify.File(hashFacadePath).Remove(),
	).Execute()
}
