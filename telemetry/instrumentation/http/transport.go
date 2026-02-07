package http

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	contractsconfig "github.com/goravel/framework/contracts/config"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
)

// NewTransport returns an http.RoundTripper instrumented with OpenTelemetry.
// It wraps the provided base RoundTripper with otelhttp using the configured
// telemetry facade's tracer provider, meter provider, and propagator.
//
// If telemetry is nil, no instrumentation is applied. In that case,
// http.DefaultTransport is returned when base is nil; otherwise the provided
// base RoundTripper is returned.
func NewTransport(config contractsconfig.Config, telemetry contractstelemetry.Telemetry, base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	if config == nil || telemetry == nil {
		return base
	}

	if !config.GetBool("telemetry.instrumentation.http_client", true) {
		return base
	}

	return otelhttp.NewTransport(
		base,
		otelhttp.WithTracerProvider(telemetry.TracerProvider()),
		otelhttp.WithMeterProvider(telemetry.MeterProvider()),
		otelhttp.WithPropagators(telemetry.Propagator()),
	)
}
