package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	stubs := Stubs{}
	telemetryConfigPath := path.Config("telemetry.go")
	telemetryFacadePath := path.Facades("telemetry.go")
	telemetryServiceProvider := "&telemetry.ServiceProvider{}"
	modulePath := packages.GetModulePath()

	packages.Setup(os.Args).
		Install(
			modify.AddProviderApply(modulePath, telemetryServiceProvider),
			modify.File(telemetryConfigPath).Overwrite(stubs.Config(packages.GetModuleNameFromArgs(os.Args))),
			modify.File(telemetryFacadePath).Overwrite(stubs.TelemetryFacade()),
		).
		Uninstall(
			modify.File(telemetryConfigPath).Remove(),
			modify.File(telemetryFacadePath).Remove(),
			modify.RemoveProviderApply(modulePath, telemetryServiceProvider),
		).
		Execute()
}
