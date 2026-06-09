package telemetry

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/goravel/framework/errors"
)

func TestNewTracerProvider(t *testing.T) {
	mockFactory := func(ctx context.Context) (sdktrace.SpanExporter, error) {
		return &MockSpanExporter{}, nil
	}

	tests := []struct {
		name        string
		config      Config
		expectError error
		description string
	}{
		{
			name: "Success: Disabled (Empty Exporter)",
			config: Config{
				Traces: TracesConfig{Exporter: ""},
			},
		},
		{
			name: "Success: Console Exporter",
			config: Config{
				Traces: TracesConfig{Exporter: "console"},
				Exporters: map[string]ExporterEntry{
					"console": {Driver: TraceExporterDriverConsole, PrettyPrint: true},
				},
			},
		},
		{
			name: "Success: OTLP HTTP Exporter",
			config: Config{
				Traces: TracesConfig{Exporter: "otlp"},
				Exporters: map[string]ExporterEntry{
					"otlp": {
						Driver:   TraceExporterDriverOTLP,
						Endpoint: "localhost:4318",
						Protocol: ProtocolHTTPProtobuf,
						Insecure: true,
						Timeout:  5000,
					},
				},
			},
		},
		{
			name: "Success: OTLP gRPC Exporter",
			config: Config{
				Traces: TracesConfig{Exporter: "otlp_grpc"},
				Exporters: map[string]ExporterEntry{
					"otlp_grpc": {
						Driver:   TraceExporterDriverOTLP,
						Endpoint: "localhost:4317",
						Protocol: ProtocolGRPC,
						Insecure: true,
					},
				},
			},
		},
		{
			name: "Success: Custom Exporter (Instance)",
			config: Config{
				Traces: TracesConfig{Exporter: "custom_inst"},
				Exporters: map[string]ExporterEntry{
					"custom_inst": {
						Driver: TraceExporterDriverCustom,
						Via:    &MockSpanExporter{},
					},
				},
			},
		},
		{
			name: "Success: Custom Exporter (Factory)",
			config: Config{
				Traces: TracesConfig{Exporter: "custom_fact"},
				Exporters: map[string]ExporterEntry{
					"custom_fact": {
						Driver: TraceExporterDriverCustom,
						Via:    mockFactory,
					},
				},
			},
		},
		{
			name: "Error: Exporter Not Found",
			config: Config{
				Traces: TracesConfig{Exporter: "missing"},
				Exporters: map[string]ExporterEntry{
					"other": {Driver: TraceExporterDriverConsole},
				},
			},
			expectError: errors.TelemetryExporterNotFound,
		},
		{
			name: "Error: Unsupported Driver",
			config: Config{
				Traces: TracesConfig{Exporter: "unknown"},
				Exporters: map[string]ExporterEntry{
					"unknown": {Driver: "alien_tech"},
				},
			},
			expectError: errors.TelemetryUnsupportedDriver.Args("alien_tech"),
		},
		{
			name: "Error: Custom Driver Missing Via",
			config: Config{
				Traces: TracesConfig{Exporter: "custom_bad"},
				Exporters: map[string]ExporterEntry{
					"custom_bad": {Driver: TraceExporterDriverCustom, Via: nil},
				},
			},
			expectError: errors.TelemetryViaRequired,
		},
		{
			name: "Error: Custom Driver Wrong Type",
			config: Config{
				Traces: TracesConfig{Exporter: "custom_type"},
				Exporters: map[string]ExporterEntry{
					"custom_type": {
						Driver: TraceExporterDriverCustom,
						Via:    "invalid-string",
					},
				},
			},
			expectError: errors.TelemetryTraceViaTypeMismatch.Args("string"),
		},
		{
			name: "Error: Unsupported Protocol",
			config: Config{
				Traces: TracesConfig{Exporter: "otlp"},
				Exporters: map[string]ExporterEntry{
					"otlp": {
						Driver:   TraceExporterDriverOTLP,
						Endpoint: "localhost:4318",
						Protocol: "http/json",
					},
				},
			},
			expectError: errors.TelemetryUnsupportedProtocol.Args("http/json"),
		},
		{
			name: "Success: OTLP Endpoint With URL Path",
			config: Config{
				Traces: TracesConfig{Exporter: "otlp"},
				Exporters: map[string]ExporterEntry{
					"otlp": {
						Driver:   TraceExporterDriverOTLP,
						Endpoint: "https://collector.example.com/otel",
					},
				},
			},
		},
		{
			name: "Success: OTLP With Compression And Retry",
			config: Config{
				Traces: TracesConfig{Exporter: "otlp"},
				Exporters: map[string]ExporterEntry{
					"otlp": {
						Driver:      TraceExporterDriverOTLP,
						Endpoint:    "localhost:4318",
						Insecure:    true,
						Compression: "gzip",
						Retry:       RetryConfig{MaxElapsedTime: 10 * time.Second},
					},
				},
			},
		},
		{
			name: "Error: Unsupported Compression",
			config: Config{
				Traces: TracesConfig{Exporter: "otlp"},
				Exporters: map[string]ExporterEntry{
					"otlp": {
						Driver:      TraceExporterDriverOTLP,
						Endpoint:    "localhost:4318",
						Compression: "zstd",
					},
				},
			},
			expectError: errors.TelemetryUnsupportedCompression.Args("zstd"),
		},
		{
			name: "Error: Unsupported Processor",
			config: Config{
				Traces: TracesConfig{Exporter: "console", Processor: ProcessorConfig{Type: "alien"}},
				Exporters: map[string]ExporterEntry{
					"console": {Driver: TraceExporterDriverConsole},
				},
			},
			expectError: errors.TelemetryUnsupportedProcessor.Args("alien"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			provider, shutdown, flush, err := NewTracerProvider(ctx, tt.config)

			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
				assert.Nil(t, provider)

				if shutdown != nil {
					assert.NoError(t, shutdown(ctx))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.NotNil(t, shutdown)
				assert.NotNil(t, flush)
				assert.NoError(t, shutdown(ctx))
			}
		})
	}
}

func TestNewConsoleTraceExporter(t *testing.T) {
	tests := []struct {
		name string
		cfg  ExporterEntry
	}{
		{
			name: "Default (No Pretty Print)",
			cfg:  ExporterEntry{},
		},
		{
			name: "With Pretty Print",
			cfg:  ExporterEntry{PrettyPrint: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, err := newConsoleTraceExporter(tt.cfg)
			assert.NoError(t, err)
			assert.NotNil(t, exp)
		})
	}
}

func TestNewOTLPTraceExporter(t *testing.T) {
	tests := []struct {
		name string
		cfg  ExporterEntry
	}{
		{
			name: "Default Protocol (HTTP)",
			cfg:  ExporterEntry{Endpoint: "localhost:4318", Insecure: true},
		},
		{
			name: "gRPC Protocol",
			cfg: ExporterEntry{
				Endpoint: "localhost:4317",
				Protocol: ProtocolGRPC,
				Insecure: true,
			},
		},
		{
			name: "HTTP With Compression Retry And TLS",
			cfg: ExporterEntry{
				Endpoint:    "https://otel.com",
				Compression: "gzip",
				TLS:         TLSConfig{CA: testCAFile(t)},
				Retry:       RetryConfig{MaxElapsedTime: 5 * time.Second},
			},
		},
		{
			name: "GRPC With Compression Retry And TLS",
			cfg: ExporterEntry{
				Endpoint:    "otel.com:4317",
				Protocol:    ProtocolGRPC,
				Compression: "gzip",
				TLS:         TLSConfig{CA: testCAFile(t)},
				Retry:       RetryConfig{MaxElapsedTime: 5 * time.Second},
			},
		},
		{
			name: "Complex Configuration",
			cfg: ExporterEntry{
				Endpoint: "https://otel.com",
				Protocol: ProtocolHTTPProtobuf,
				Insecure: false, // TLS enabled
				Timeout:  5 * time.Second,
				Headers:  map[string]string{"Authorization": "Bearer token"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			exp, err := newOTLPTraceExporter(ctx, tt.cfg)

			assert.NoError(t, err)
			assert.NotNil(t, exp)
		})
	}
}

type MockSpanExporter struct{}

func (m *MockSpanExporter) ExportSpans(ctx context.Context, ss []sdktrace.ReadOnlySpan) error {
	return nil
}

func (m *MockSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}

type recordingSpanExporter struct {
	mu    sync.Mutex
	spans []sdktrace.ReadOnlySpan
}

func (r *recordingSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.spans = append(r.spans, spans...)
	return nil
}

func (r *recordingSpanExporter) Shutdown(ctx context.Context) error { return nil }

func (r *recordingSpanExporter) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.spans)
}

func TestNewTracerProvider_ProcessorTypes(t *testing.T) {
	ctx := context.Background()

	newConfig := func(exporter sdktrace.SpanExporter, processor ProcessorConfig) Config {
		return Config{
			Traces: TracesConfig{Exporter: "custom", Processor: processor},
			Exporters: map[string]ExporterEntry{
				"custom": {Driver: TraceExporterDriverCustom, Via: exporter},
			},
		}
	}

	t.Run("simple exports on span end", func(t *testing.T) {
		exporter := &recordingSpanExporter{}
		provider, shutdown, _, err := NewTracerProvider(ctx, newConfig(exporter, ProcessorConfig{Type: ProcessorSimple}))
		assert.NoError(t, err)

		_, span := provider.Tracer("test").Start(ctx, "operation")
		span.End()

		assert.Equal(t, 1, exporter.count())
		assert.NoError(t, shutdown(ctx))
	})

	t.Run("batch defers export until shutdown", func(t *testing.T) {
		exporter := &recordingSpanExporter{}
		provider, shutdown, _, err := NewTracerProvider(ctx, newConfig(exporter, ProcessorConfig{Type: ProcessorBatch, Interval: time.Hour}))
		assert.NoError(t, err)

		_, span := provider.Tracer("test").Start(ctx, "operation")
		span.End()

		assert.Equal(t, 0, exporter.count())
		assert.NoError(t, shutdown(ctx))
		assert.Equal(t, 1, exporter.count())
	})
}
