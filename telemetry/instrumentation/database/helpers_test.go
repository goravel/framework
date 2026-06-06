package database

import (
	"context"
	"sync"
	"testing"

	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/telemetry"
)

type recordingSpanExporter struct {
	mu    sync.Mutex
	spans []sdktrace.ReadOnlySpan
}

func (r *recordingSpanExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.spans = append(r.spans, spans...)
	return nil
}

func (r *recordingSpanExporter) Shutdown(_ context.Context) error { return nil }

func setupRecordingTelemetry(t *testing.T) *recordingSpanExporter {
	t.Helper()

	exporter := &recordingSpanExporter{}
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	mockTelemetry := mockstelemetry.NewTelemetry(t)
	mockTelemetry.EXPECT().Tracer(instrumentationName).Return(provider.Tracer(instrumentationName)).Maybe()
	mockTelemetry.EXPECT().Meter(instrumentationName).Return(metricnoop.NewMeterProvider().Meter(instrumentationName)).Maybe()

	original := telemetry.Facade
	telemetry.Facade = mockTelemetry
	t.Cleanup(func() { telemetry.Facade = original })

	return exporter
}
