package telemetry

import (
	"context"

	oteltrace "go.opentelemetry.io/otel/trace"
)

type Telemetry interface {
	// Tracer returns a tracer for the given instrumentation name.
	Tracer(name string, opts ...oteltrace.TracerOption) oteltrace.Tracer

	// TracerProvider returns the underlying trace provider.
	TracerProvider() oteltrace.TracerProvider

	// Shutdown flushes remaining spans and releases resources.
	Shutdown(ctx context.Context) error
}
