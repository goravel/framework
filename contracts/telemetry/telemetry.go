package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Telemetry interface {
	// Propagator returns the configured text map propagator for context propagation.
	Propagator() propagation.TextMapPropagator

	// Shutdown flushes remaining spans and releases resources.
	Shutdown(ctx context.Context) error

	// Tracer returns a tracer for the given instrumentation name.
	Tracer(name string, opts ...oteltrace.TracerOption) oteltrace.Tracer

	// TracerProvider returns the underlying trace provider.
	TracerProvider() oteltrace.TracerProvider
}
