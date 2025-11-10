package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	testingFacade := "Testing"
	providersBootstrapPath := path.Bootstrap("providers.go")
	testingServiceProvider := "&testing.ServiceProvider{}"
	testCasePath := path.Base("tests", "test_case.go")
	exampleTestPath := path.Base("tests", "feature", "example_test.go")
	testingFacadePath := path.Facades("testing.go")

	packages.Setup(os.Args).
		Install(
			modify.GoFile(providersBootstrapPath).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register(testingServiceProvider)),
			modify.File(testCasePath).Overwrite(stubs.TestCase()),
			modify.File(exampleTestPath).Overwrite(stubs.ExampleTest()),
			modify.WhenFacade(testingFacade, modify.File(testingFacadePath).Overwrite(stubs.TestingFacade())),
		).
		Uninstall(
			modify.File(exampleTestPath).Remove(),
			modify.File(testCasePath).Remove(),
			modify.WhenNoFacades([]string{testingFacade},
				modify.GoFile(providersBootstrapPath).
					Find(match.Providers()).Modify(modify.Unregister(testingServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
			),
			modify.WhenFacade(testingFacade, modify.File(testingFacadePath).Remove()),
		).
		Execute()
}
