package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/goravel/framework/errors"
)

func TestNewApplication(t *testing.T) {
	tests := []struct {
		name            string
		config          Config
		expectError     error
		expectSDKTracer bool
		expectSDKMeter  bool
	}{
		{
			name:        "Error: Empty propagator",
			config:      Config{Propagators: ""},
			expectError: errors.TelemetryPropagatorRequired,
		},
		{
			name:        "Error: Invalid propagator",
			config:      Config{Propagators: "invalid"},
			expectError: errors.TelemetryUnsupportedPropagator.Args("invalid"),
		},
		{
			name: "Success: Defaults (No Exporters) returns Noop providers",
			config: Config{
				Service:     ServiceConfig{Name: "goravel"},
				Propagators: "tracecontext",
			},
			expectSDKTracer: false,
			expectSDKMeter:  false,
		},
		{
			name: "Success: Console Exporters initialize SDKs",
			config: Config{
				Service:     ServiceConfig{Name: "test-service"},
				Propagators: "tracecontext",
				Traces:      TracesConfig{Exporter: "console"},
				Metrics:     MetricsConfig{Exporter: "console"},
				Exporters: map[string]ExporterEntry{
					"console": {Driver: TraceExporterDriverConsole},
				},
			},
			expectSDKTracer: true,
			expectSDKMeter:  true,
		},
		{
			name: "Success: OTLP Exporters initialize SDKs",
			config: Config{
				Service:     ServiceConfig{Name: "test-service"},
				Propagators: "tracecontext",
				Traces:      TracesConfig{Exporter: "otlp"},
				Metrics:     MetricsConfig{Exporter: "otlp"},
				Exporters: map[string]ExporterEntry{
					"otlp": {
						Driver:   TraceExporterDriverOTLP,
						Endpoint: "localhost:4318",
						Protocol: ProtocolHTTPProtobuf,
						Insecure: true,
						Timeout:  5 * time.Second,
					},
				},
			},
			expectSDKTracer: true,
			expectSDKMeter:  true,
		},
		{
			name: "Error: Unknown Exporter",
			config: Config{
				Service:     ServiceConfig{Name: "goravel"},
				Propagators: "tracecontext",
				Traces:      TracesConfig{Exporter: "unknown"},
				Exporters:   map[string]ExporterEntry{},
			},
			expectError: errors.TelemetryExporterNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := NewApplication(tt.config)

			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
				assert.Nil(t, app)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, app)
			assert.NotNil(t, app.tracerProvider)
			assert.NotNil(t, app.meterProvider)

			if tt.expectSDKTracer {
				_, ok := app.tracerProvider.(*sdktrace.TracerProvider)
				assert.True(t, ok, "expected SDK tracer provider")
			} else {
				_, ok := app.tracerProvider.(tracenoop.TracerProvider)
				assert.True(t, ok, "expected noop tracer provider")
			}

			if tt.expectSDKMeter {
				_, ok := app.meterProvider.(*sdkmetric.MeterProvider)
				assert.True(t, ok, "expected SDK meter provider")
			} else {
				_, ok := app.meterProvider.(metricnoop.MeterProvider)
				assert.True(t, ok, "expected noop meter provider")
			}
		})
	}
}

func TestApplication_Tracer(t *testing.T) {
	app := &Application{
		tracerProvider: tracenoop.NewTracerProvider(),
	}
	tracer := app.Tracer("test-tracer")
	assert.NotNil(t, tracer)
}

func TestApplication_Meter(t *testing.T) {
	app := &Application{
		meterProvider: metricnoop.NewMeterProvider(),
	}
	meter := app.Meter("test-meter")
	assert.NotNil(t, meter)
}

func TestApplication_Propagator(t *testing.T) {
	t.Run("returns nil when not set", func(t *testing.T) {
		app := &Application{}
		assert.Nil(t, app.Propagator())
	})

	t.Run("returns set propagator", func(t *testing.T) {
		customPropagator, err := newCompositeTextMapPropagator("tracecontext")
		assert.NoError(t, err)
		app := &Application{propagator: customPropagator}
		assert.Equal(t, customPropagator, app.Propagator())
	})
}

func TestApplication_Shutdown(t *testing.T) {
	t.Run("executes all shutdown functions", func(t *testing.T) {
		shutdownCallCount := 0

		mockShutdown := func(ctx context.Context) error {
			shutdownCallCount++
			return nil
		}

		app := &Application{
			shutdownFuncs: []ShutdownFunc{mockShutdown, mockShutdown},
		}

		err := app.Shutdown(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 2, shutdownCallCount, "Shutdown should have been called twice")
	})

	t.Run("aggregates errors", func(t *testing.T) {
		app := &Application{
			shutdownFuncs: []ShutdownFunc{
				func(ctx context.Context) error { return errors.New("error 1") },
				func(ctx context.Context) error { return nil },
				func(ctx context.Context) error { return errors.New("error 2") },
			},
		}

		err := app.Shutdown(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error 1")
		assert.Contains(t, err.Error(), "error 2")
	})
}

func TestConfig_GetExporter(t *testing.T) {
	tests := []struct {
		name         string
		config       Config
		exporterName string
		expectDriver ExporterDriver
		expectFound  bool
	}{
		{
			name: "returns existing exporter",
			config: Config{
				Exporters: map[string]ExporterEntry{
					"otlp": {Driver: TraceExporterDriverOTLP, Endpoint: "localhost:4318"},
				},
			},
			exporterName: "otlp",
			expectDriver: TraceExporterDriverOTLP,
			expectFound:  true,
		},
		{
			name: "returns zero value for non-existent exporter",
			config: Config{
				Exporters: map[string]ExporterEntry{},
			},
			exporterName: "unknown",
			// When key is not found, Driver string is empty "", NOT "unknown"
			expectDriver: "",
			expectFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, found := tt.config.GetExporter(tt.exporterName)
			assert.Equal(t, tt.expectFound, found)
			assert.Equal(t, tt.expectDriver, exp.Driver)
		})
	}
}
