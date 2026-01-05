package http

import (
	"net/http"
	"sync"

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

	return &TransportProxy{
		base: base,
	}
}

type TransportProxy struct {
	base          http.RoundTripper
	otelTransport http.RoundTripper
	once          sync.Once
}

func (t *TransportProxy) RoundTrip(req *http.Request) (*http.Response, error) {
	t.once.Do(func() {
		if telemetry.ConfigFacade == nil || !telemetry.ConfigFacade.GetBool("telemetry.instrumentation.http_client", true) {
			return
		}

		if telemetry.TelemetryFacade == nil {
			color.Warningln("[Telemetry] Facade not initialized. HTTP client instrumentation is disabled.")
			return
		}

		t.otelTransport = otelhttp.NewTransport(
			t.base,
			otelhttp.WithTracerProvider(telemetry.TelemetryFacade.TracerProvider()),
			otelhttp.WithMeterProvider(telemetry.TelemetryFacade.MeterProvider()),
			otelhttp.WithPropagators(telemetry.TelemetryFacade.Propagator()),
		)
	})

	if t.otelTransport != nil {
		return t.otelTransport.RoundTrip(req)
	}

	return t.base.RoundTrip(req)
}
