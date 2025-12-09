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
	telemetryConfigPath := path.Config("telemetry.go")
	telemetryFacadePath := path.Facades("telemetry.go")
	telemetryServiceProvider := "&telemetry.ServiceProvider{}"
	packageName := setup.PackageName()
	modulePath := setup.ModulePath()
	configPackage := support.PathPackage(support.Config.Paths.Config, packageName)
	facadePackage := support.PathPackage(support.Config.Paths.Facades, packageName)

	setup.Install(
		modify.AddProviderApply(modulePath, telemetryServiceProvider),
		modify.File(telemetryConfigPath).Overwrite(stubs.Config(configPackage, packageName)),
		modify.File(telemetryFacadePath).Overwrite(stubs.TelemetryFacade(facadePackage)),
	).Uninstall(
		modify.File(telemetryConfigPath).Remove(),
		modify.File(telemetryFacadePath).Remove(),
		modify.RemoveProviderApply(modulePath, telemetryServiceProvider),
	).Execute()
}
