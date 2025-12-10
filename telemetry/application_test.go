package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	}{
		{
			name:        "empty propagator returns error",
			config:      Config{Propagators: ""},
			expectError: errors.TelemetryPropagatorRequired,
		},
		{
			name:        "invalid propagator returns error",
			config:      Config{Propagators: "invalid"},
			expectError: errors.TelemetryUnsupportedPropagator.Args("invalid"),
		},
		{
			name: "empty exporter returns app with noop tracer provider",
			config: Config{
				Service: ServiceConfig{
					Name: "goravel",
				},
				Propagators: "tracecontext",
			},
			expectSDKTracer: false,
		},
		{
			name: "console exporter initializes SDK tracer provider",
			config: Config{
				Service:     ServiceConfig{Name: "test-service"},
				Propagators: "tracecontext,baggage",
				Traces:      TracesConfig{Exporter: "console"},
				Exporters: map[string]ExporterEntry{
					"console": {Driver: TraceExporterDriverConsole},
				},
			},
			expectSDKTracer: true,
		},
		{
			name: "otlp exporter initializes SDK tracer provider",
			config: Config{
				Service:     ServiceConfig{Name: "test-service", Version: "1.0.0", Environment: "test"},
				Propagators: "tracecontext",
				Traces: TracesConfig{
					Exporter: "otlp",
					Sampler:  SamplerConfig{Type: "traceidratio", Ratio: 0.5, Parent: true},
				},
				Exporters: map[string]ExporterEntry{
					"otlp": {
						Driver:   TraceExporterDriverOTLP,
						Endpoint: "localhost:4318",
						Protocol: ProtocolHTTPProtobuf,
						Insecure: true,
						Timeout:  5000 * time.Millisecond,
					},
				},
			},
			expectSDKTracer: true,
		},
		{
			name: "zipkin exporter initializes SDK tracer provider",
			config: Config{
				Service: ServiceConfig{
					Name: "goravel",
				},
				Propagators: "b3",
				Traces:      TracesConfig{Exporter: "zipkin"},
				Exporters: map[string]ExporterEntry{
					"zipkin": {
						Driver:   TraceExporterDriverZipkin,
						Endpoint: "http://localhost:9411/api/v2/spans",
					},
				},
			},
			expectSDKTracer: true,
		},
		{
			name: "unknown exporter returns error",
			config: Config{
				Service: ServiceConfig{
					Name: "goravel",
				},
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
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, app)
			assert.NotNil(t, app.tracerProvider)

			if tt.expectSDKTracer {
				_, ok := app.tracerProvider.(*sdktrace.TracerProvider)
				assert.True(t, ok, "expected SDK tracer provider")
			} else {
				_, ok := app.tracerProvider.(tracenoop.TracerProvider)
				assert.True(t, ok, "expected noop tracer provider")
			}
		})
	}
}

func TestApplication_Tracer(t *testing.T) {
	t.Run("returns tracer from noop provider", func(t *testing.T) {
		app := &Application{
			tracerProvider: tracenoop.NewTracerProvider(),
		}

		tracer := app.Tracer("test-tracer")

		assert.NotNil(t, tracer)
	})

	t.Run("returns tracer from SDK provider", func(t *testing.T) {
		tp := sdktrace.NewTracerProvider()
		app := &Application{
			tracerProvider: tp,
		}

		tracer := app.Tracer("test-tracer")

		assert.NotNil(t, tracer)
	})
}

func TestApplication_TracerProvider(t *testing.T) {
	t.Run("returns noop provider", func(t *testing.T) {
		noopTP := tracenoop.NewTracerProvider()
		app := &Application{
			tracerProvider: noopTP,
		}

		provider := app.TracerProvider()

		assert.Equal(t, noopTP, provider)
	})

	t.Run("returns SDK provider", func(t *testing.T) {
		tp := sdktrace.NewTracerProvider()
		app := &Application{
			tracerProvider: tp,
		}

		provider := app.TracerProvider()

		assert.Equal(t, tp, provider)
	})
}

func TestApplication_Propagator(t *testing.T) {
	t.Run("returns nil when not set", func(t *testing.T) {
		app := &Application{}

		propagator := app.Propagator()

		assert.Nil(t, propagator)
	})

	t.Run("returns set propagator", func(t *testing.T) {
		customPropagator, err := newCompositeTextMapPropagator("b3")
		assert.NoError(t, err)

		app := &Application{
			propagator: customPropagator,
		}

		propagator := app.Propagator()

		assert.Equal(t, customPropagator, propagator)
	})
}

func TestApplication_Shutdown(t *testing.T) {
	t.Run("returns nil when tracer provider is noop", func(t *testing.T) {
		app := &Application{
			tracerProvider: tracenoop.NewTracerProvider(),
		}

		err := app.Shutdown(context.Background())

		assert.NoError(t, err)
	})

	t.Run("shuts down SDK tracer provider", func(t *testing.T) {
		tp := sdktrace.NewTracerProvider()
		app := &Application{
			tracerProvider: tp,
		}

		err := app.Shutdown(context.Background())

		assert.NoError(t, err)
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
			name: "returns default for non-existent exporter",
			config: Config{
				Exporters: map[string]ExporterEntry{},
			},
			exporterName: "unknown",
			expectDriver: "unknown",
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
