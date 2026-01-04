package log

import (
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/telemetry"
)

const defaultInstrumentationName = "github.com/goravel/framework/telemetry/instrumentation/log"

type TelemetryChannel struct{}

func NewTelemetryChannel() *TelemetryChannel {
	return &TelemetryChannel{}
}

func (r *TelemetryChannel) Handle(channelPath string) (log.Handler, error) {
	config := telemetry.ConfigFacade
	if config == nil {
		return nil, errors.ConfigFacadeNotSet
	}

	if !config.GetBool("telemetry.instrumentation.log", true) {
		return &handler{enabled: false}, nil
	}

	if telemetry.TelemetryFacade == nil {
		return nil, errors.TelemetryFacadeNotSet
	}

	instrumentName := config.GetString(channelPath+".instrument_name", defaultInstrumentationName)
	return &handler{
		enabled: true,
		logger:  telemetry.TelemetryFacade.Logger(instrumentName),
	}, nil
}
