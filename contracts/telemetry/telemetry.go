package telemetry

import (
	"context"

	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Telemetry interface {
	// Meter returns a metric.Meter instance for recording metrics under the given instrumentation name.
	// The optional metric.MeterOption parameters allow further customization.
	Meter(name string, opts ...otelmetric.MeterOption) otelmetric.Meter

	// Propagator returns the configured TextMapPropagator used to inject and extract
	// context across service boundaries for distributed tracing.
	Propagator() propagation.TextMapPropagator

	// Shutdown flushes any pending telemetry data and releases associated resources.
	// This should typically be called during application shutdown.
	Shutdown(ctx context.Context) error

	// Tracer returns a trace.Tracer instance for the given instrumentation name.
	// Optional trace.TracerOption parameters allow customization of tracer behavior.
	Tracer(name string, opts ...oteltrace.TracerOption) oteltrace.Tracer
}
