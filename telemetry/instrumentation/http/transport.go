package http

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/telemetry"
)

func NewTransport(base http.RoundTripper) http.RoundTripper {
	if telemetry.TelemetryFacade == nil {
		color.Warningln("[Telemetry] Facade not initialized. HTTP client instrumentation is disabled.")
		if base == nil {
			return http.DefaultTransport
		}
		return base
	}

	return otelhttp.NewTransport(
		base,
		otelhttp.WithTracerProvider(telemetry.TelemetryFacade.TracerProvider()),
		otelhttp.WithMeterProvider(telemetry.TelemetryFacade.MeterProvider()),
		otelhttp.WithPropagators(telemetry.TelemetryFacade.Propagator()),
	)
}
