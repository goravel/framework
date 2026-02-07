package log

import (
	contractsconfig "github.com/goravel/framework/contracts/config"
	contractslog "github.com/goravel/framework/contracts/log"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
)

const DefaultInstrumentationName = "github.com/goravel/framework/telemetry/instrumentation/log"

type TelemetryChannel struct {
	config   contractsconfig.Config
	resolver contractstelemetry.Resolver
}

func NewTelemetryChannel(config contractsconfig.Config, telemetry contractstelemetry.Telemetry) *TelemetryChannel {
	return NewLazyTelemetryChannel(config, func() contractstelemetry.Telemetry {
		return telemetry
	})
}

func NewLazyTelemetryChannel(config contractsconfig.Config, resolver contractstelemetry.Resolver) *TelemetryChannel {
	return &TelemetryChannel{
		config:   config,
		resolver: resolver,
	}
}

func (r *TelemetryChannel) Handle(channelPath string) (contractslog.Handler, error) {
	if r.config == nil || !r.config.GetBool("telemetry.instrumentation.log", false) {
		return &handler{enabled: false}, nil
	}

	instrumentName := r.config.GetString(channelPath+".instrument_name", DefaultInstrumentationName)
	return &handler{
		resolver:       r.resolver,
		enabled:        true,
		instrumentName: instrumentName,
	}, nil
}
