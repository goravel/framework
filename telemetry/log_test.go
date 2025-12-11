package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	"github.com/goravel/framework/errors"
)

func TestNewLoggerProvider(t *testing.T) {
	mockFactory := func(ctx context.Context) (sdklog.Exporter, error) {
		return &MockLogExporter{}, nil
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
				Logs: LogsConfig{Exporter: ""},
			},
		},
		{
			name: "Success: Console Driver",
			config: Config{
				Logs: LogsConfig{
					Exporter: "console",
					Processor: LogsProcessorConfig{
						Interval: 5000,
						Timeout:  2000,
					},
				},
				Exporters: map[string]ExporterEntry{
					"console": {Driver: LogExporterDriverConsole},
				},
			},
		},
		{
			name: "Success: OTLP HTTP Driver",
			config: Config{
				Logs: LogsConfig{Exporter: "otlp"},
				Exporters: map[string]ExporterEntry{
					"otlp": {
						Driver:   LogExporterDriverOTLP,
						Endpoint: "localhost:4318",
						Protocol: ProtocolHTTPProtobuf,
						Insecure: true,
					},
				},
			},
		},
		{
			name: "Success: OTLP gRPC Driver",
			config: Config{
				Logs: LogsConfig{Exporter: "otlp_grpc"},
				Exporters: map[string]ExporterEntry{
					"otlp_grpc": {
						Driver:   LogExporterDriverOTLP,
						Endpoint: "localhost:4317",
						Protocol: ProtocolGRPC,
						Insecure: true,
					},
				},
			},
		},
		{
			name: "Success: Custom Driver (Via Instance)",
			config: Config{
				Logs: LogsConfig{Exporter: "custom_instance"},
				Exporters: map[string]ExporterEntry{
					"custom_instance": {
						Driver: LogExporterDriverCustom,
						Via:    &MockLogExporter{},
					},
				},
			},
		},
		{
			name: "Success: Custom Driver (Via Factory)",
			config: Config{
				Logs: LogsConfig{Exporter: "custom_factory"},
				Exporters: map[string]ExporterEntry{
					"custom_factory": {
						Driver: LogExporterDriverCustom,
						Via:    mockFactory,
					},
				},
			},
		},
		{
			name: "Error: Exporter Not Found",
			config: Config{
				Logs: LogsConfig{Exporter: "missing_exporter"},
				Exporters: map[string]ExporterEntry{
					"other": {Driver: LogExporterDriverConsole},
				},
			},
			expectError: errors.TelemetryExporterNotFound,
		},
		{
			name: "Error: Unsupported Driver",
			config: Config{
				Logs: LogsConfig{Exporter: "unknown"},
				Exporters: map[string]ExporterEntry{
					"unknown": {Driver: "alien_technology"},
				},
			},
			expectError: errors.TelemetryUnsupportedDriver.Args("alien_technology"),
		},
		{
			name: "Error: Custom Driver Missing Via",
			config: Config{
				Logs: LogsConfig{Exporter: "custom_invalid"},
				Exporters: map[string]ExporterEntry{
					"custom_invalid": {
						Driver: LogExporterDriverCustom,
						Via:    nil,
					},
				},
			},
			expectError: errors.TelemetryViaRequired,
		},
		{
			name: "Error: Custom Driver Wrong Type",
			config: Config{
				Logs: LogsConfig{Exporter: "custom_wrong_type"},
				Exporters: map[string]ExporterEntry{
					"custom_wrong_type": {
						Driver: LogExporterDriverCustom,
						Via:    "i-am-a-string",
					},
				},
			},
			expectError: errors.TelemetryLogViaTypeMismatch.Args("string"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			provider, shutdown, err := NewLoggerProvider(ctx, tt.config)

			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err, tt.description)
				assert.Nil(t, provider)

				if shutdown != nil {
					assert.NoError(t, shutdown(ctx))
				}
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, provider)
				assert.NotNil(t, shutdown)

				assert.NoError(t, shutdown(ctx))
			}
		})
	}
}

func TestNewOTLPLogExporter(t *testing.T) {
	tests := []struct {
		name        string
		cfg         ExporterEntry
		expectError bool
	}{
		{
			name: "HTTP Protobuf (Default)",
			cfg: ExporterEntry{
				Driver:   LogExporterDriverOTLP,
				Endpoint: "localhost:4318",
				Insecure: true,
			},
		},
		{
			name: "gRPC",
			cfg: ExporterEntry{
				Driver:   LogExporterDriverOTLP,
				Protocol: ProtocolGRPC,
				Endpoint: "localhost:4317",
				Insecure: true,
			},
		},
		{
			name: "With Headers and Timeout",
			cfg: ExporterEntry{
				Driver:   LogExporterDriverOTLP,
				Endpoint: "localhost:4318",
				Insecure: true,
				Timeout:  5000,
				Headers:  map[string]string{"Authorization": "Bearer token"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			exp, err := newOTLPLogExporter(ctx, tt.cfg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, exp)
			}
		})
	}
}

func TestNewConsoleLogExporter(t *testing.T) {
	exp, err := newConsoleLogExporter()
	assert.NoError(t, err)
	assert.NotNil(t, exp)
}

func TestBuildOTLPLogOptions(t *testing.T) {
	cfg := ExporterEntry{
		Endpoint: "http://example.com",
		Insecure: true,
		Timeout:  1234 * time.Millisecond,
		Headers:  map[string]string{"key": "val"},
	}

	type Option int
	const (
		OptEndpoint Option = iota
		OptInsecure
		OptTimeout
		OptHeaders
	)

	opts := buildOTLPLogOptions[Option](
		cfg,
		func(e string) Option {
			assert.Equal(t, "example.com", e)
			return OptEndpoint
		},
		func() Option {
			return OptInsecure
		},
		func(d time.Duration) Option {
			assert.Equal(t, 1234*time.Millisecond, d)
			return OptTimeout
		},
		func(h map[string]string) Option {
			assert.Equal(t, "val", h["key"])
			return OptHeaders
		},
	)

	assert.Len(t, opts, 4)
}

type MockLogExporter struct{}

func (m *MockLogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	return nil
}

func (m *MockLogExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (m *MockLogExporter) ForceFlush(ctx context.Context) error {
	return nil
}
