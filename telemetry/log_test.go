package telemetry

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	otellog "go.opentelemetry.io/otel/log"
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
					Exporter:  "console",
					Processor: ProcessorConfig{Type: ProcessorBatch},
				},
				Exporters: map[string]ExporterEntry{
					"console": {Driver: LogExporterDriverConsole, PrettyPrint: true},
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
		{
			name: "Error: Unsupported Protocol",
			config: Config{
				Logs: LogsConfig{Exporter: "otlp"},
				Exporters: map[string]ExporterEntry{
					"otlp": {
						Driver:   LogExporterDriverOTLP,
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
				Logs: LogsConfig{Exporter: "otlp"},
				Exporters: map[string]ExporterEntry{
					"otlp": {
						Driver:   LogExporterDriverOTLP,
						Endpoint: "https://collector.example.com/otel",
					},
				},
			},
		},
		{
			name: "Success: OTLP With Compression And Retry",
			config: Config{
				Logs: LogsConfig{Exporter: "otlp"},
				Exporters: map[string]ExporterEntry{
					"otlp": {
						Driver:      LogExporterDriverOTLP,
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
				Logs: LogsConfig{Exporter: "otlp"},
				Exporters: map[string]ExporterEntry{
					"otlp": {
						Driver:      LogExporterDriverOTLP,
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
				Logs: LogsConfig{Exporter: "console", Processor: ProcessorConfig{Type: "alien"}},
				Exporters: map[string]ExporterEntry{
					"console": {Driver: LogExporterDriverConsole},
				},
			},
			expectError: errors.TelemetryUnsupportedProcessor.Args("alien"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			provider, shutdown, flush, err := NewLoggerProvider(ctx, tt.config)

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
				assert.NotNil(t, flush)
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
			name: "HTTP With Compression Retry And TLS",
			cfg: ExporterEntry{
				Driver:      LogExporterDriverOTLP,
				Endpoint:    "https://otel.com",
				Compression: "gzip",
				TLS:         TLSConfig{CA: testCAFile(t)},
				Retry:       RetryConfig{MaxElapsedTime: 5 * time.Second},
			},
		},
		{
			name: "GRPC With Compression Retry And TLS",
			cfg: ExporterEntry{
				Driver:      LogExporterDriverOTLP,
				Endpoint:    "otel.com:4317",
				Protocol:    ProtocolGRPC,
				Compression: "gzip",
				TLS:         TLSConfig{CA: testCAFile(t)},
				Retry:       RetryConfig{MaxElapsedTime: 5 * time.Second},
			},
		},
		{
			name: "With Headers and Timeout",
			cfg: ExporterEntry{
				Driver:   LogExporterDriverOTLP,
				Endpoint: "localhost:4318",
				Insecure: true,
				Timeout:  5000 * time.Millisecond,
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
			exp, err := newConsoleLogExporter(tt.cfg)
			assert.NoError(t, err)
			assert.NotNil(t, exp)
		})
	}
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

type recordingLogExporter struct {
	mu      sync.Mutex
	records []sdklog.Record
}

func (r *recordingLogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, records...)
	return nil
}

func (r *recordingLogExporter) Shutdown(ctx context.Context) error   { return nil }
func (r *recordingLogExporter) ForceFlush(ctx context.Context) error { return nil }

func (r *recordingLogExporter) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.records)
}

func TestNewLoggerProvider_ProcessorTypes(t *testing.T) {
	ctx := context.Background()

	newConfig := func(exporter sdklog.Exporter, processor ProcessorConfig) Config {
		return Config{
			Logs: LogsConfig{Exporter: "custom", Processor: processor},
			Exporters: map[string]ExporterEntry{
				"custom": {Driver: LogExporterDriverCustom, Via: exporter},
			},
		}
	}

	emit := func(provider otellog.LoggerProvider) {
		var record otellog.Record
		record.SetBody(otellog.StringValue("message"))
		provider.Logger("test").Emit(ctx, record)
	}

	t.Run("simple exports on emit", func(t *testing.T) {
		exporter := &recordingLogExporter{}
		provider, shutdown, _, err := NewLoggerProvider(ctx, newConfig(exporter, ProcessorConfig{Type: ProcessorSimple}))
		assert.NoError(t, err)

		emit(provider)

		assert.Equal(t, 1, exporter.count())
		assert.NoError(t, shutdown(ctx))
	})

	t.Run("batch defers export until shutdown", func(t *testing.T) {
		exporter := &recordingLogExporter{}
		provider, shutdown, _, err := NewLoggerProvider(ctx, newConfig(exporter, ProcessorConfig{Type: ProcessorBatch, Interval: time.Hour}))
		assert.NoError(t, err)

		emit(provider)

		assert.Equal(t, 0, exporter.count())
		assert.NoError(t, shutdown(ctx))
		assert.Equal(t, 1, exporter.count())
	})
}
