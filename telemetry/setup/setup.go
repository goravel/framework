package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	const (
		fileTelemetryConfig = "telemetry.go"
		fileLoggingConfig   = "logging.go"
		providerTelemetry   = "&telemetry.ServiceProvider{}"
		keyLoggingChannels  = "logging.channels"
		keyOtel             = "otel"
		configSnippetOtel   = `map[string]any{
		"driver":          "otel",
		"instrument_name": config.GetString("APP_NAME", "goravel/log"),
	}`
	)

	setup := packages.Setup(os.Args)
	stubs := Stubs{}
	paths := setup.Paths()

	pathConfigTelemetry := path.Config(fileTelemetryConfig)
	pathFacadesTelemetry := path.Facade(fileTelemetryConfig)
	pathConfigLogging := path.Config(fileLoggingConfig)

	moduleImport := paths.Module().Import()
	facadesPackage := paths.Facades().Package()

	matchLoggingChannels := match.Config(keyLoggingChannels)

	setup.Install(
		// Add Telemetry Service Provider
		modify.AddProviderApply(moduleImport, providerTelemetry),

		// Add Telemetry Config and Facade
		modify.File(pathConfigTelemetry).Overwrite(stubs.Config(paths.Config().Package(), paths.Facades().Import(), facadesPackage)),
		modify.File(pathFacadesTelemetry).Overwrite(stubs.TelemetryFacade(facadesPackage)),

		// Add Otel Channel to Logging Config
		modify.GoFile(pathConfigLogging).
			Find(matchLoggingChannels).
			Modify(modify.AddConfig(keyOtel, configSnippetOtel)),
	).Uninstall(
		// Remove Telemetry Config and Facade
		modify.File(pathConfigTelemetry).Remove(),
		modify.File(pathFacadesTelemetry).Remove(),

		// Remove Telemetry Service Provider
		modify.RemoveProviderApply(moduleImport, providerTelemetry),

		// Remove Otel Channel from Logging Config
		modify.GoFile(pathConfigLogging).
			Find(matchLoggingChannels).
			Modify(modify.RemoveConfig(keyOtel)),
	).Execute()
}