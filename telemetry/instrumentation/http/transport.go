package http

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/telemetry"
)

// NewTransport returns an http.RoundTripper instrumented with OpenTelemetry.
// It wraps the provided base RoundTripper with otelhttp using the configured
// telemetry facade's tracer provider, meter provider, and propagator.
//
// If telemetry.TelemetryFacade is nil, a warning is logged and no
// instrumentation is applied. In that case, http.DefaultTransport is returned
// when base is nil; otherwise the provided base RoundTripper is returned.
func NewTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	if telemetry.TelemetryFacade == nil {
		color.Warningln("[Telemetry] Facade not initialized. HTTP client instrumentation is disabled.")
		return base
	}

	return otelhttp.NewTransport(
		base,
		otelhttp.WithTracerProvider(telemetry.TelemetryFacade.TracerProvider()),
		otelhttp.WithMeterProvider(telemetry.TelemetryFacade.MeterProvider()),
		otelhttp.WithPropagators(telemetry.TelemetryFacade.Propagator()),
	)
}
