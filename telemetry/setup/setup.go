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
	telemetryServiceProvider := "&telemetry.ServiceProvider{}"
	modulePath := setup.Paths().Module().Import()

	setup.Install(
		modify.AddProviderApply(modulePath, telemetryServiceProvider),
		modify.File(telemetryConfigPath).Overwrite(stubs.Config(setup.Paths().Config().Package(), setup.Paths().Facades().Import())),
		modify.File(telemetryFacadePath).Overwrite(stubs.TelemetryFacade(setup.Paths().Facades().Package())),
	).Uninstall(
		modify.File(telemetryConfigPath).Remove(),
		modify.File(telemetryFacadePath).Remove(),
		modify.RemoveProviderApply(modulePath, telemetryServiceProvider),
	).Execute()
}
