package http

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/goravel/framework/telemetry"
)

func NewTransport(base http.RoundTripper) http.RoundTripper {
	return nil
}

func getOptions() []otelhttp.Option {
	return []otelhttp.Option{
		otelhttp.WithTracerProvider(telemetry.TelemetryFacade.TracerProvider()),
		otelhttp.WithMeterProvider(telemetry.TelemetryFacade.MeterProvider()),
		otelhttp.WithPropagators(telemetry.TelemetryFacade.Propagator()),
	}
}
