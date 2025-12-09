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
	telemetryConfigPath := path.Config("telemetry.go")
	telemetryFacadePath := path.Facades("telemetry.go")
	packageName := setup.PackageName()
	modulePath := setup.ModulePath()
	telemetryServiceProvider := "&telemetry.ServiceProvider{}"

	packages.Setup(os.Args).
		Install(
			modify.AddProviderApply(modulePath, telemetryServiceProvider),
			modify.File(telemetryConfigPath).Overwrite(stubs.Config(packageName)),
			modify.File(telemetryFacadePath).Overwrite(stubs.TelemetryFacade()),
		).
		Uninstall(
			modify.File(telemetryConfigPath).Remove(),
			modify.File(telemetryFacadePath).Remove(),
			modify.RemoveProviderApply(modulePath, telemetryServiceProvider),
		).
		Execute()
}
