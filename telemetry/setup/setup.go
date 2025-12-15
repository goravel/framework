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
		importLog           = "github.com/goravel/framework/telemetry/instrumentation/log"
		aliasLog            = "telemetrylog"
		keyLoggingChannels  = "logging.channels"
		keyOtel             = "otel"
		configSnippetOtel   = `map[string]any{
		"driver": "custom",
		"via":    telemetrylog.NewTelemetryChannel(),
	}`
	)

	setup := packages.Setup(os.Args)
	stubs := Stubs{}
	paths := setup.Paths()

	pathConfigTelemetry := path.Config(fileTelemetryConfig)
	pathFacadesTelemetry := path.Facades(fileTelemetryConfig)
	pathConfigLogging := path.Config(fileLoggingConfig)

	importModule := paths.Module().Import()
	packageFacades := paths.Facades().Package()

	matchLoggingChannels := match.Config(keyLoggingChannels)
	matchImports := match.Imports()

	setup.Install(
		modify.AddProviderApply(importModule, providerTelemetry),
		modify.File(pathConfigTelemetry).Overwrite(stubs.Config(paths.Config().Package(), paths.Facades().Import(), packageFacades)),
		modify.File(pathFacadesTelemetry).Overwrite(stubs.TelemetryFacade(packageFacades)),
		modify.GoFile(pathConfigLogging).
			Find(matchImports).
			Modify(modify.AddImport(importLog, aliasLog)).
			Find(matchLoggingChannels).
			Modify(modify.AddConfig(keyOtel, configSnippetOtel)),
	).Uninstall(
		modify.File(pathConfigTelemetry).Remove(),
		modify.File(pathFacadesTelemetry).Remove(),
		modify.RemoveProviderApply(importModule, providerTelemetry),
		modify.GoFile(pathConfigLogging).
			Find(matchLoggingChannels).
			Modify(modify.RemoveConfig(keyOtel)).
			Find(matchImports).
			Modify(modify.RemoveImport(importLog, aliasLog)),
	).Execute()
}
