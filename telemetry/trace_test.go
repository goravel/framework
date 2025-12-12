package telemetry

import (
	"context"
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
			name: "Success: Zipkin Exporter",
			config: Config{
				Traces: TracesConfig{Exporter: "zipkin"},
				Exporters: map[string]ExporterEntry{
					"zipkin": {
						Driver:   TraceExporterDriverZipkin,
						Endpoint: "http://localhost:9411/api/v2/spans",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			provider, shutdown, err := NewTracerProvider(ctx, tt.config)

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

func TestNewZipkinTraceExporter(t *testing.T) {
	tests := []struct {
		name        string
		cfg         ExporterEntry
		expectError error
	}{
		{
			name:        "Error: Empty Endpoint",
			cfg:         ExporterEntry{},
			expectError: errors.TelemetryZipkinEndpointRequired,
		},
		{
			name: "Success: Valid Endpoint",
			cfg:  ExporterEntry{Endpoint: "http://zipkin:9411/api/v2/spans"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, err := newZipkinTraceExporter(tt.cfg)
			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, exp)
			}
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
