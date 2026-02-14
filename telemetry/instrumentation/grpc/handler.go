package grpc

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/stats"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/telemetry"
)

// NewServerStatsHandler creates an OTel stats handler for the server.
func NewServerStatsHandler(opts ...Option) stats.Handler {
	if telemetry.Facade == nil {
		color.Warningln(errors.TelemetryGrpcServerStatsHandlerDisabled.Error())
		return nil
	}

	if telemetry.ConfigFacade == nil || !telemetry.ConfigFacade.GetBool("telemetry.instrumentation.grpc_server.enabled") {
		return nil
	}

	finalOpts := append(getCommonOptions(), opts...)

	return otelgrpc.NewServerHandler(finalOpts...)
}

// NewClientStatsHandler creates an OTel stats handler for the client.
func NewClientStatsHandler(opts ...Option) stats.Handler {
	if telemetry.Facade == nil {
		color.Warningln(errors.TelemetryGrpcClientStatsHandlerDisabled.Error())
		return nil
	}

	if telemetry.ConfigFacade == nil || !telemetry.ConfigFacade.GetBool("telemetry.instrumentation.grpc_client.enabled") {
		return nil
	}

	finalOpts := append(getCommonOptions(), opts...)

	return otelgrpc.NewClientHandler(finalOpts...)
}

func getCommonOptions() []otelgrpc.Option {
	return []otelgrpc.Option{
		otelgrpc.WithTracerProvider(telemetry.Facade.TracerProvider()),
		otelgrpc.WithMeterProvider(telemetry.Facade.MeterProvider()),
		otelgrpc.WithPropagators(telemetry.Facade.Propagator()),
		otelgrpc.WithMessageEvents(otelgrpc.ReceivedEvents, otelgrpc.SentEvents),
	}
}
