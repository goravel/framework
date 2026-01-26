package grpc

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/stats"

	contractsconfig "github.com/goravel/framework/contracts/config"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
)

// NewServerStatsHandler creates an OTel stats handler for the server.
func NewServerStatsHandler(config contractsconfig.Config, telemetry contractstelemetry.Telemetry, opts ...Option) stats.Handler {
	if config == nil || !config.GetBool("telemetry.instrumentation.grpc_server") {
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
	if config == nil || !config.GetBool("telemetry.instrumentation.grpc_client") {
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
