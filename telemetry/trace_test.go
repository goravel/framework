package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/goravel/framework/errors"
)

func TestNewTracerProvider(t *testing.T) {
	mockFactory := func(ctx context.Context) (sdktrace.SpanExporter, error) {
		return &MockSpanExporter{}, nil
	}
	tests := []struct {
		name         string
		config       Config
		exporterName string
		expectError  error
	}{
		{
			name: "creates console exporter",
			config: Config{
				Traces: TracesConfig{
					Exporter: "console",
				},
				Exporters: map[string]ExporterEntry{
					"console": {Driver: TraceExporterDriverConsole},
				},
			},
		},
		{
			name: "creates otlp exporter",
			config: Config{
				Traces: TracesConfig{
					Exporter: "otlp",
				},
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
			name: "creates zipkin exporter",
			config: Config{
				Traces: TracesConfig{
					Exporter: "zipkin",
				},
				Exporters: map[string]ExporterEntry{
					"zipkin": {
						Driver:   TraceExporterDriverZipkin,
						Endpoint: "http://localhost:9411/api/v2/spans",
					},
				},
			},
		},
		{
			name: "creates custom exporter (via struct instance)",
			config: Config{
				Traces: TracesConfig{
					Exporter: "custom_instance",
				},
				Exporters: map[string]ExporterEntry{
					"custom_instance": {
						Driver: TraceExporterDriverCustom,
						Via:    &MockSpanExporter{},
					},
				},
			},
		},
		{
			name: "creates custom exporter (via factory function)",
			config: Config{
				Traces: TracesConfig{
					Exporter: "custom_factory",
				},
				Exporters: map[string]ExporterEntry{
					"custom_factory": {
						Driver: TraceExporterDriverCustom,
						Via:    mockFactory,
					},
				},
			},
		},
		{
			name: "fails custom exporter when Via is missing",
			config: Config{
				Traces: TracesConfig{
					Exporter: "custom_invalid",
				},
				Exporters: map[string]ExporterEntry{
					"custom_invalid": {
						Driver: TraceExporterDriverCustom,
						Via:    nil,
					},
				},
			},
			expectError: errors.TelemetryViaRequired,
		},
		{
			name: "fails custom exporter when Via is wrong type",
			config: Config{
				Traces: TracesConfig{
					Exporter: "custom_wrong_type",
				},
				Exporters: map[string]ExporterEntry{
					"custom_wrong_type": {
						Driver: TraceExporterDriverCustom,
						Via:    "i-am-a-string-not-an-exporter",
					},
				},
			},
			expectError: errors.TelemetryTraceViaTypeMismatch.Args("string"),
		},
		{
			name: "returns error for unknown exporter",
			config: Config{
				Traces: TracesConfig{
					Exporter: "unknown",
				},
				Exporters: map[string]ExporterEntry{},
			},
			expectError: errors.TelemetryExporterNotFound,
		},
		{
			name: "uses custom driver from config",
			config: Config{
				Traces: TracesConfig{
					Exporter: "custom",
				},
				Exporters: map[string]ExporterEntry{
					"custom": {Driver: TraceExporterDriverConsole},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			exp, _, err := NewTracerProvider(ctx, tt.config)

			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, exp)
		})
	}
}

func TestNewConsoleTraceExporter(t *testing.T) {
	exp, err := newConsoleTraceExporter()

	assert.NoError(t, err)
	assert.NotNil(t, exp)
}

func TestNewZipkinTraceExporter(t *testing.T) {
	tests := []struct {
		name        string
		cfg         ExporterEntry
		expectError error
	}{
		{
			name:        "empty endpoint",
			cfg:         ExporterEntry{},
			expectError: errors.TelemetryZipkinEndpointRequired,
		},
		{
			name: "custom endpoint",
			cfg:  ExporterEntry{Endpoint: "http://zipkin:9411/api/v2/spans"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, err := newZipkinTraceExporter(tt.cfg)
			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
				return
			}
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
			name: "default protocol (http/protobuf)",
			cfg:  ExporterEntry{Endpoint: "localhost:4318", Insecure: true},
		},
		{
			name: "grpc protocol",
			cfg: ExporterEntry{
				Endpoint: "localhost:4317",
				Protocol: ProtocolGRPC,
				Insecure: true,
			},
		},
		{
			name: "with headers and timeout",
			cfg: ExporterEntry{
				Endpoint: "localhost:4318",
				Protocol: ProtocolHTTPProtobuf,
				Insecure: true,
				Timeout:  5000,
				Headers: map[string]string{
					"Authorization": "Bearer token",
				},
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
