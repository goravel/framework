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
	testingServiceProvider := "&testing.ServiceProvider{}"
	testCasePath := path.Test("test_case.go")
	exampleTestPath := path.Test("feature", "example_test.go")
	testingFacadePath := path.Facade("testing.go")
	fakeFacadePath := path.Facade("fake.go")
	moduleImport := setup.Paths().Module().Import()
	facadesPackage := setup.Paths().Facades().Package()

	setup.Install(
		// Add the testing service provider to the providers array in bootstrap/providers.go
		modify.WhenFileNotContains(path.Bootstrap("providers.go"), testingServiceProvider, modify.RegisterProvider(moduleImport, testingServiceProvider)),

		modify.WhenFacade(facades.Testing,
			// Create tests/test_case.go
			modify.File(testCasePath).Overwrite(stubs.TestCase(setup.Paths().Tests().Package(), setup.Paths().Bootstrap().Import(), setup.Paths().Bootstrap().Package())),

			// Create tests/feature/example_test.go
			modify.File(exampleTestPath).Overwrite(stubs.ExampleTest(setup.Paths().Tests().Import(), setup.Paths().Tests().Package())),

			// Add the Testing facade
			modify.File(testingFacadePath).Overwrite(stubs.TestingFacade(facadesPackage)),
		),

		// Add the Fake facade
		modify.WhenFacade(facades.Fake, modify.File(fakeFacadePath).Overwrite(stubs.FakeFacade(facadesPackage))),
	).Uninstall(
		modify.WhenFacade(facades.Testing,
			// Remove tests/feature/example_test.go
			modify.File(exampleTestPath).Remove(),

			// Remove tests/test_case.go
			modify.File(testCasePath).Remove(),

			// Remove the Testing facade
			modify.File(testingFacadePath).Remove(),
		),

		// Remove the Fake facade
		modify.WhenFacade(facades.Fake, modify.File(fakeFacadePath).Remove()),

		modify.WhenNoFacades([]string{facades.Testing, facades.Fake},
			// Remove the testing service provider from the providers array in bootstrap/providers.go
			modify.UnregisterProvider(moduleImport, testingServiceProvider),
		),
	).Execute()
}
