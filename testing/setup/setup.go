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
	testingServiceProvider := "&testing.ServiceProvider{}"
	testCasePath := path.Base("tests", "test_case.go")
	exampleTestPath := path.Base("tests", "feature", "example_test.go")
	testingFacadePath := path.Facades("testing.go")
	modulePath := setup.ModulePath()

	setup.Install(
		// Add the testing service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(modulePath, testingServiceProvider),

		// Create tests/test_case.go
		modify.File(testCasePath).Overwrite(stubs.TestCase()),

		// Create tests/feature/example_test.go
		modify.File(exampleTestPath).Overwrite(stubs.ExampleTest()),

		// Add the Testing facade
		modify.File(testingFacadePath).Overwrite(stubs.TestingFacade()),
	).Uninstall(
		// Remove tests/feature/example_test.go
		modify.File(exampleTestPath).Remove(),

		// Remove tests/test_case.go
		modify.File(testCasePath).Remove(),

		// Remove the testing service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(modulePath, testingServiceProvider),

		// Remove the Testing facade
		modify.File(testingFacadePath).Remove(),
	).Execute()
}
