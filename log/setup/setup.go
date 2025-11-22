package main

import (
	"os"

	"github.com/goravel/framework/contracts/facades"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	logFacadePath := path.Facades("log.go")
	loggingConfigPath := path.Config("logging.go")
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	logServiceProvider := "&log.ServiceProvider{}"
	envPath := path.Base(".env")
	envExamplePath := path.Base(".env.example")
	env := `
LOG_CHANNEL=stack
LOG_LEVEL=debug
`

	packages.Setup(os.Args).
		Install(
			// Add the log service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(packages.GetModulePath(), logServiceProvider),

			// Create config/logging.go
			modify.File(loggingConfigPath).Overwrite(stubs.Config(moduleName)),

			// Add the Log facade
			modify.WhenFacade(facades.Log, modify.File(logFacadePath).Overwrite(stubs.LogFacade())),

			// Add configurations to the .env and .env.example files
			modify.WhenFileNotContains(envPath, "LOG_CHANNEL", modify.File(envPath).Append(env)),
			modify.WhenFileNotContains(envExamplePath, "LOG_CHANNEL", modify.File(envExamplePath).Append(env)),
		).
		Uninstall(
			modify.WhenNoFacades([]string{facades.Log},
				// Remove config/logging.go
				modify.File(loggingConfigPath).Remove(),

				// Remove the log service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(packages.GetModulePath(), logServiceProvider),
			),

			// Remove the Log facade
			modify.WhenFacade(facades.Log, modify.File(logFacadePath).Remove()),
		).
		Execute()
}
