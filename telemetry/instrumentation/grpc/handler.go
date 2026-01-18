package grpc

import (
	contractsconfig "github.com/goravel/framework/contracts/config"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/stats"
)

// NewServerStatsHandler creates an OTel stats handler for the server.
func NewServerStatsHandler(config contractsconfig.Config, telemetry contractstelemetry.Telemetry, opts ...Option) stats.Handler {
	if config == nil || !config.GetBool("telemetry.instrumentation.grpc_server", true) {
		return nil
	}

	if telemetry == nil {
		return nil
	}

	finalOpts := append(getCommonOptions(telemetry), opts...)

	return otelgrpc.NewServerHandler(finalOpts...)
}

// NewClientStatsHandler creates an OTel stats handler for the client.
func NewClientStatsHandler(config contractsconfig.Config, telemetry contractstelemetry.Telemetry, opts ...Option) stats.Handler {
	if config == nil || !config.GetBool("telemetry.instrumentation.grpc_client", true) {
		return nil
	}

	if telemetry == nil {
		return nil
	}

	finalOpts := append(getCommonOptions(telemetry), opts...)

	return otelgrpc.NewClientHandler(finalOpts...)
}

func getCommonOptions(telemetry contractstelemetry.Telemetry) []otelgrpc.Option {
	return []otelgrpc.Option{
		otelgrpc.WithTracerProvider(telemetry.TracerProvider()),
		otelgrpc.WithMeterProvider(telemetry.MeterProvider()),
		otelgrpc.WithPropagators(telemetry.Propagator()),
		otelgrpc.WithMessageEvents(otelgrpc.ReceivedEvents, otelgrpc.SentEvents),
	}
}
