package log

import (
	contractsconfig "github.com/goravel/framework/contracts/config"
	contractslog "github.com/goravel/framework/contracts/log"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
)

const defaultInstrumentationName = "github.com/goravel/framework/telemetry/instrumentation/log"

type TelemetryChannel struct {
	config    contractsconfig.Config
	telemetry contractstelemetry.Telemetry
}

func NewTelemetryChannel(config contractsconfig.Config, telemetry contractstelemetry.Telemetry) *TelemetryChannel {
	return &TelemetryChannel{
		config:    config,
		telemetry: telemetry,
	}
}

func (r *TelemetryChannel) Handle(channelPath string) (contractslog.Handler, error) {
	if !r.config.GetBool("telemetry.instrumentation.log", true) {
		return &handler{enabled: false}, nil
	}

	instrumentName := r.config.GetString(channelPath+".instrument_name", defaultInstrumentationName)
	return &handler{
		telemetry:      r.telemetry,
		enabled:        true,
		instrumentName: instrumentName,
	}, nil
}
