package log

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/log"
)

const defaultInstrumentationName = "github.com/goravel/framework/telemetry/instrumentation/log"

type TelemetryChannel struct {
	config config.Config
}

func NewTelemetryChannel(config config.Config) *TelemetryChannel {
	return &TelemetryChannel{
		config: config,
	}
}

func (r *TelemetryChannel) Handle(channelPath string) (log.Handler, error) {
	if !r.config.GetBool("telemetry.instrumentation.log", true) {
		return &handler{enabled: false}, nil
	}

	instrumentName := r.config.GetString(channelPath+".instrument_name", defaultInstrumentationName)
	return &handler{
		enabled:        true,
		instrumentName: instrumentName,
	}, nil
}
