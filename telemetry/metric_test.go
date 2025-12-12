package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/goravel/framework/errors"
)

func TestNewMeterProvider(t *testing.T) {
	manualReader := sdkmetric.NewManualReader()
	mockFactory := func(ctx context.Context) (sdkmetric.Reader, error) {
		return manualReader, nil
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
				Metrics: MetricsConfig{Exporter: ""},
			},
		},
		{
			name: "Success: Console Driver",
			config: Config{
				Metrics: MetricsConfig{Exporter: "console"},
				Exporters: map[string]ExporterEntry{
					"console": {Driver: MetricsExporterDriverConsole, PrettyPrint: true},
				},
			},
		},
		{
			name: "Success: OTLP HTTP Driver",
			config: Config{
				Metrics: MetricsConfig{Exporter: "otlp"},
				Exporters: map[string]ExporterEntry{
					"otlp": {
						Driver:   MetricsExporterDriverOTLP,
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
				Metrics: MetricsConfig{Exporter: "otlp_grpc"},
				Exporters: map[string]ExporterEntry{
					"otlp_grpc": {
						Driver:   MetricsExporterDriverOTLP,
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
				Metrics: MetricsConfig{Exporter: "custom_instance"},
				Exporters: map[string]ExporterEntry{
					"custom_instance": {
						Driver: MetricsExporterDriverCustom,
						Via:    manualReader,
					},
				},
			},
		},
		{
			name: "Success: Custom Driver (Via Factory)",
			config: Config{
				Metrics: MetricsConfig{Exporter: "custom_factory"},
				Exporters: map[string]ExporterEntry{
					"custom_factory": {
						Driver: MetricsExporterDriverCustom,
						Via:    mockFactory,
					},
				},
			},
		},
		{
			name: "Error: Exporter Not Found",
			config: Config{
				Metrics: MetricsConfig{Exporter: "missing_exporter"},
				Exporters: map[string]ExporterEntry{
					"other": {Driver: MetricsExporterDriverConsole},
				},
			},
			expectError: errors.TelemetryExporterNotFound,
		},
		{
			name: "Error: Unsupported Driver",
			config: Config{
				Metrics: MetricsConfig{Exporter: "alien_tech"},
				Exporters: map[string]ExporterEntry{
					"alien_tech": {Driver: "alien_driver"},
				},
			},
			expectError: errors.TelemetryUnsupportedDriver.Args("alien_driver"),
		},
		{
			name: "Error: Custom Driver Missing Via",
			config: Config{
				Metrics: MetricsConfig{Exporter: "custom_invalid"},
				Exporters: map[string]ExporterEntry{
					"custom_invalid": {
						Driver: MetricsExporterDriverCustom,
						Via:    nil,
					},
				},
			},
			expectError: errors.TelemetryViaRequired,
		},
		{
			name: "Error: Custom Driver Wrong Type",
			config: Config{
				Metrics: MetricsConfig{Exporter: "custom_wrong_type"},
				Exporters: map[string]ExporterEntry{
					"custom_wrong_type": {
						Driver: MetricsExporterDriverCustom,
						Via:    "i-am-a-string",
					},
				},
			},
			expectError: errors.TelemetryMetricViaTypeMismatch.Args("string"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			provider, shutdown, err := NewMeterProvider(ctx, tt.config)

			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err, tt.description)
				assert.Nil(t, provider)
				if shutdown != nil {
					_ = shutdown(ctx)
				}
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, provider)
				assert.NotNil(t, shutdown)
				_ = shutdown(ctx)
			}
		})
	}
}

func TestNewMetricReader_Config(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		readerCfg MetricsReaderConfig
		cfg       ExporterEntry
	}{
		{
			name:      "Defaults",
			readerCfg: MetricsReaderConfig{}, // Should trigger defaults
			cfg:       ExporterEntry{Driver: MetricsExporterDriverConsole},
		},
		{
			name: "Custom Values",
			readerCfg: MetricsReaderConfig{
				Interval: 10 * time.Second,
				Timeout:  5 * time.Second,
			},
			cfg: ExporterEntry{Driver: MetricsExporterDriverConsole},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := newMetricReader(ctx, tt.cfg, tt.readerCfg)
			assert.NoError(t, err)
			assert.NotNil(t, reader)
		})
	}
}

func TestGetTemporalitySelector(t *testing.T) {
	tests := []struct {
		name        string
		temporality MetricTemporality
		expected    metricdata.Temporality
	}{
		{
			name:        "Default (Cumulative)",
			temporality: "",
			expected:    metricdata.CumulativeTemporality,
		},
		{
			name:        "Explicit Cumulative",
			temporality: TemporalityCumulative,
			expected:    metricdata.CumulativeTemporality,
		},
		{
			name:        "Explicit Delta",
			temporality: TemporalityDelta,
			expected:    metricdata.DeltaTemporality,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := getTemporalitySelector(tt.temporality)
			// Check Counter kind (usually what we care about for Delta vs Cumulative)
			result := selector(sdkmetric.InstrumentKindCounter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewConsoleMetricExporter(t *testing.T) {
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
			exp, err := newConsoleMetricExporter(tt.cfg)
			assert.NoError(t, err)
			assert.NotNil(t, exp)
		})
	}
}

func TestNewOTLPMetricExporter(t *testing.T) {
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
			name: "With Temporality Delta",
			cfg: ExporterEntry{
				Endpoint:          "localhost:4318",
				Protocol:          ProtocolHTTPProtobuf,
				MetricTemporality: TemporalityDelta,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			exp, err := newOTLPMetricExporter(ctx, tt.cfg)

			assert.NoError(t, err)
			assert.NotNil(t, exp)
		})
	}
}
