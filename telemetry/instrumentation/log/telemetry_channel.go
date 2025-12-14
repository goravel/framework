package log

import (
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/telemetry"
)

const (
	defaultInstrumentationName = "github.com/goravel/framework/telemetry/instrumentation/log"
	configKeyEnabled           = "telemetry.instrumentation.log.enabled"
	configKeyName              = "telemetry.instrumentation.log.name"
)

type TelemetryChannel struct{}

func NewTelemetryChannel() *TelemetryChannel {
	return &TelemetryChannel{}
}

func (r *TelemetryChannel) Handle(_ string) (log.Hook, error) {
	if telemetry.TelemetryFacade == nil {
		return nil, errors.TelemetryFacadeNotSet
	}

	config := telemetry.ConfigFacade
	if config == nil {
		return nil, errors.ConfigFacadeNotSet
	}

	if !config.GetBool(configKeyEnabled) {
		return &hook{enabled: false}, nil
	}

	instrumentName := config.GetString(configKeyName, defaultInstrumentationName)
	return &hook{
		enabled: true,
		logger:  telemetry.TelemetryFacade.Logger(instrumentName),
	}, nil
}
