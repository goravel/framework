package grpc

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/stats"

	"github.com/goravel/framework/telemetry"
)

func ServerHandler() stats.Handler {
	if telemetry.TelemetryFacade == nil {
		return nil
	}

	return otelgrpc.NewServerHandler(getOptions()...)
}

func ClientHandler() stats.Handler {
	if telemetry.TelemetryFacade == nil {
		return nil
	}

	return otelgrpc.NewClientHandler(getOptions()...)
}

func getOptions() []otelgrpc.Option {
	return []otelgrpc.Option{
		otelgrpc.WithTracerProvider(telemetry.TelemetryFacade.TracerProvider()),
		otelgrpc.WithMeterProvider(telemetry.TelemetryFacade.MeterProvider()),
		otelgrpc.WithPropagators(telemetry.TelemetryFacade.Propagator()),
	}
}
