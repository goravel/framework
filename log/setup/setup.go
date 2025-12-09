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
	logFacadePath := path.Facades("log.go")
	loggingConfigPath := path.Config("logging.go")
	packageName := setup.PackageName()
	logServiceProvider := "&log.ServiceProvider{}"
	envPath := path.Base(".env")
	envExamplePath := path.Base(".env.example")
	env := `
LOG_CHANNEL=stack
LOG_LEVEL=debug
`

	setup.Install(
		// Add the log service provider to the providers array in bootstrap/providers.go
		modify.AddProviderApply(setup.ModulePath(), logServiceProvider),

		// Create config/logging.go
		modify.File(loggingConfigPath).Overwrite(stubs.Config(packageName)),

		// Add the Log facade
		modify.File(logFacadePath).Overwrite(stubs.LogFacade()),

		// Add configurations to the .env and .env.example files
		modify.WhenFileNotContains(envPath, "LOG_CHANNEL", modify.File(envPath).Append(env)),
		modify.WhenFileNotContains(envExamplePath, "LOG_CHANNEL", modify.File(envExamplePath).Append(env)),
	).Uninstall(
		// Remove config/logging.go
		modify.File(loggingConfigPath).Remove(),

		// Remove the log service provider from the providers array in bootstrap/providers.go
		modify.RemoveProviderApply(setup.ModulePath(), logServiceProvider),

		// Remove the Log facade
		modify.File(logFacadePath).Remove(),
	).Execute()
}
