package grpc

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/stats"

	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/telemetry"
)

// NewServerStatsHandler creates an OTel stats handler for the server.
func NewServerStatsHandler(opts ...Option) stats.Handler {
	if telemetry.ConfigFacade == nil || !telemetry.ConfigFacade.GetBool("telemetry.instrumentation.grpc_server", true) {
		return nil
	}

	if telemetry.TelemetryFacade == nil {
		color.Warningln("[Telemetry] Facade not initialized. gRPC server stats instrumentation is disabled.")
		return nil
	}

	finalOpts := append(getCommonOptions(), opts...)

	return otelgrpc.NewServerHandler(finalOpts...)
}

// NewClientStatsHandler creates an OTel stats handler for the client.
func NewClientStatsHandler(opts ...Option) stats.Handler {
	if telemetry.ConfigFacade == nil || !telemetry.ConfigFacade.GetBool("telemetry.instrumentation.grpc_client", true) {
		return nil
	}

	if telemetry.TelemetryFacade == nil {
		color.Warningln("[Telemetry] Facade not initialized. gRPC client stats instrumentation is disabled.")
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
