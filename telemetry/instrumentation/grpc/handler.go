package grpc

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/stats"

	"github.com/goravel/framework/telemetry"
)

// ServerStatsHandler creates an OTel stats handler for the server.
func ServerStatsHandler(opts ...Option) stats.Handler {
	if telemetry.TelemetryFacade == nil {
		return nil
	}

	finalOpts := append(getCommonOptions(), opts...)

	return otelgrpc.NewServerHandler(finalOpts...)
}

// ClientStatsHandler creates an OTel stats handler for the client.
func ClientStatsHandler(opts ...Option) stats.Handler {
	if telemetry.TelemetryFacade == nil {
		return nil
	}

	finalOpts := append(getCommonOptions(), opts...)

	return otelgrpc.NewClientHandler(finalOpts...)
}

func getCommonOptions() []otelgrpc.Option {
	return []otelgrpc.Option{
		otelgrpc.WithTracerProvider(telemetry.TelemetryFacade.TracerProvider()),
		otelgrpc.WithMeterProvider(telemetry.TelemetryFacade.MeterProvider()),
		otelgrpc.WithPropagators(telemetry.TelemetryFacade.Propagator()),
		otelgrpc.WithMessageEvents(otelgrpc.ReceivedEvents, otelgrpc.SentEvents),
	}
}
