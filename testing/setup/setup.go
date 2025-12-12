package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	stubs := Stubs{}
	testingServiceProvider := "&testing.ServiceProvider{}"
	testCasePath := path.Base(support.Config.Paths.Tests, "test_case.go")
	exampleTestPath := path.Base(support.Config.Paths.Tests, "feature", "example_test.go")
	testingFacadePath := path.Facades("testing.go")
	moduleImport := setup.Paths().Module().Import()

	setup.Install(
		// Add the testing service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(moduleImport, testingServiceProvider),

		// Create tests/test_case.go
		modify.File(testCasePath).Overwrite(stubs.TestCase(setup.Paths().Tests().Package(), setup.Paths().Bootstrap().Import())),

		// Create tests/feature/example_test.go
		modify.File(exampleTestPath).Overwrite(stubs.ExampleTest(setup.Paths().Tests().Import(), setup.Paths().Tests().Package())),

		// Add the Testing facade
		modify.File(testingFacadePath).Overwrite(stubs.TestingFacade(setup.Paths().Facades().Package())),
	).Uninstall(
		// Remove tests/feature/example_test.go
		modify.File(exampleTestPath).Remove(),

		// Remove tests/test_case.go
		modify.File(testCasePath).Remove(),

		// Remove the testing service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(moduleImport, testingServiceProvider),

		// Remove the Testing facade
		modify.File(testingFacadePath).Remove(),
	).Execute()
}
