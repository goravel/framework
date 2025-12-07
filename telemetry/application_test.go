package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	configmocks "github.com/goravel/framework/mocks/config"
)

func TestNewApplication(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(*configmocks.Config)
		expectError  bool
		expectNilApp bool
		expectTracer bool
	}{
		{
			name: "no exporter returns app without tracer provider",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configPropagators.String()).Return("")
				cfg.EXPECT().GetString(configTracesExporter.String()).Return("")
			},
			expectTracer: false,
		},
		{
			name: "none exporter returns app without tracer provider",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configPropagators.String()).Return("")
				cfg.EXPECT().GetString(configTracesExporter.String()).Return("none")
			},
			expectTracer: false,
		},
		{
			name: "console exporter initializes tracer provider",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configPropagators.String()).Return("")
				cfg.EXPECT().GetString(configTracesExporter.String()).Return("console")

				cfg.EXPECT().GetString(configServiceName.String(), "goravel").Return("test-service")
				cfg.EXPECT().GetString(configServiceVersion.String()).Return("1.0.0")
				cfg.EXPECT().GetString(configEnvironment.String()).Return("test")

				cfg.EXPECT().Get(configTracesSamplerRatio.String(), defaultRatio).Return(1.0)
				cfg.EXPECT().GetString(configTracesSamplerType.String(), "always_on").Return("always_on")
				cfg.EXPECT().GetBool(configTracesSamplerParent.String(), true).Return(true)

				cfg.EXPECT().GetString(configExporterDriver.With("console"), "console").Return("console")
			},
			expectTracer: true,
		},
		{
			name: "otlp exporter initializes tracer provider",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configPropagators.String()).Return("tracecontext")
				cfg.EXPECT().GetString(configTracesExporter.String()).Return("otlp")

				cfg.EXPECT().GetString(configServiceName.String(), "goravel").Return("test-service")
				cfg.EXPECT().GetString(configServiceVersion.String()).Return("1.0.0")
				cfg.EXPECT().GetString(configEnvironment.String()).Return("test")

				cfg.EXPECT().Get(configTracesSamplerRatio.String(), defaultRatio).Return(0.5)
				cfg.EXPECT().GetString(configTracesSamplerType.String(), "always_on").Return("traceidratio")
				cfg.EXPECT().GetBool(configTracesSamplerParent.String(), true).Return(true)

				cfg.EXPECT().GetString(configExporterDriver.With("otlp"), "otlp").Return("otlp")
				cfg.EXPECT().GetString(configExporterTracesProtocol.With("otlp"), "").Return("")
				cfg.EXPECT().GetString(configExporterProtocol.With("otlp"), protocolHTTPProtobuf).Return(protocolHTTPProtobuf)
				cfg.EXPECT().GetInt(configExporterTracesTimeout.With("otlp"), mock.AnythingOfType("int")).Return(5000)
				cfg.EXPECT().GetInt(configExporterTimeout.With("otlp"), defaultTimeout).Return(10000)
				cfg.EXPECT().GetString(configExporterTracesHeaders.With("otlp")).Return("")
				cfg.EXPECT().GetString(configExporterEndpoint.With("otlp")).Return("localhost:4318")
				cfg.EXPECT().GetBool(configExporterInsecure.With("otlp")).Return(true)
			},
			expectTracer: true,
		},
		{
			name: "zipkin exporter initializes tracer provider",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configPropagators.String()).Return("")
				cfg.EXPECT().GetString(configTracesExporter.String()).Return("zipkin")

				cfg.EXPECT().GetString(configServiceName.String(), "goravel").Return("goravel")
				cfg.EXPECT().GetString(configServiceVersion.String()).Return("")
				cfg.EXPECT().GetString(configEnvironment.String()).Return("")

				cfg.EXPECT().Get(configTracesSamplerRatio.String(), defaultRatio).Return(defaultRatio)
				cfg.EXPECT().GetString(configTracesSamplerType.String(), "always_on").Return("always_on")
				cfg.EXPECT().GetBool(configTracesSamplerParent.String(), true).Return(true)

				cfg.EXPECT().GetString(configExporterDriver.With("zipkin"), "zipkin").Return("zipkin")
				cfg.EXPECT().GetString(configExporterEndpoint.With("zipkin")).Return("http://localhost:9411/api/v2/spans")
			},
			expectTracer: true,
		},
		{
			name: "unknown exporter returns error",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configPropagators.String()).Return("")
				cfg.EXPECT().GetString(configTracesExporter.String()).Return("unknown")

				cfg.EXPECT().GetString(configServiceName.String(), "goravel").Return("goravel")
				cfg.EXPECT().GetString(configServiceVersion.String()).Return("")
				cfg.EXPECT().GetString(configEnvironment.String()).Return("")

				cfg.EXPECT().GetString(configExporterDriver.With("unknown"), "unknown").Return("unknown")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := configmocks.NewConfig(t)
			tt.setupMock(mockConfig)

			app, err := NewApplication(mockConfig)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.expectNilApp {
				assert.Nil(t, app)
				return
			}

			assert.NotNil(t, app)

			if tt.expectTracer {
				_, ok := app.tracerProvider.(*sdktrace.TracerProvider)
				assert.True(t, ok, "expected SDK tracer provider")
			} else {
				assert.Nil(t, app.tracerProvider)
			}
		})
	}
}

func TestApplication_Tracer(t *testing.T) {
	t.Run("returns tracer from noop provider when no tracer provider set", func(t *testing.T) {
		app := &Application{}

		tracer := app.Tracer("test-tracer")

		assert.NotNil(t, tracer)
	})

	t.Run("returns tracer from provider when tracer provider is set", func(t *testing.T) {
		tp := sdktrace.NewTracerProvider()
		app := &Application{
			tracerProvider: tp,
		}

		tracer := app.Tracer("test-tracer")

		assert.NotNil(t, tracer)
	})
}

func TestApplication_TracerProvider(t *testing.T) {
	t.Run("returns noop provider when no tracer provider set", func(t *testing.T) {
		app := &Application{}

		provider := app.TracerProvider()

		assert.NotNil(t, provider)
		_, ok := provider.(tracenoop.TracerProvider)
		assert.True(t, ok, "expected noop tracer provider")
	})

	t.Run("returns set tracer provider", func(t *testing.T) {
		tp := sdktrace.NewTracerProvider()
		app := &Application{
			tracerProvider: tp,
		}

		provider := app.TracerProvider()

		assert.Equal(t, tp, provider)
	})
}

func TestApplication_Propagator(t *testing.T) {
	t.Run("returns default composite propagator when not set", func(t *testing.T) {
		app := &Application{}

		propagator := app.Propagator()

		assert.NotNil(t, propagator)
		assert.Equal(t, defaultCompositePropagator, propagator)
	})

	t.Run("returns set propagator", func(t *testing.T) {
		customPropagator := newCompositeTextMapPropagator("b3")
		app := &Application{
			propagator: customPropagator,
		}

		propagator := app.Propagator()

		assert.Equal(t, customPropagator, propagator)
	})
}

func TestApplication_Shutdown(t *testing.T) {
	t.Run("returns nil when tracer provider is nil", func(t *testing.T) {
		app := &Application{}

		err := app.Shutdown(context.Background())

		assert.NoError(t, err)
	})

	t.Run("returns nil when tracer provider is not SDK provider", func(t *testing.T) {
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

func TestApplication_resourceConfig(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*configmocks.Config)
		expectService string
		expectVersion string
		expectEnv     string
	}{
		{
			name: "returns default service name",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configServiceName.String(), "goravel").Return("goravel")
				cfg.EXPECT().GetString(configServiceVersion.String()).Return("")
				cfg.EXPECT().GetString(configEnvironment.String()).Return("")
			},
			expectService: "goravel",
		},
		{
			name: "returns custom service name and version",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configServiceName.String(), "goravel").Return("my-service")
				cfg.EXPECT().GetString(configServiceVersion.String()).Return("2.0.0")
				cfg.EXPECT().GetString(configEnvironment.String()).Return("production")
			},
			expectService: "my-service",
			expectVersion: "2.0.0",
			expectEnv:     "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := configmocks.NewConfig(t)
			tt.setupMock(mockConfig)

			app := &Application{config: mockConfig}
			cfg := app.resourceConfig()

			assert.Equal(t, tt.expectService, cfg.serviceName)
			assert.Equal(t, tt.expectVersion, cfg.serviceVersion)
			assert.Equal(t, tt.expectEnv, cfg.environment)
		})
	}
}

func TestApplication_samplerConfig(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(*configmocks.Config)
		expectType   string
		expectParent bool
		expectRatio  float64
	}{
		{
			name: "returns default sampler config",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().Get(configTracesSamplerRatio.String(), defaultRatio).Return(defaultRatio)
				cfg.EXPECT().GetString(configTracesSamplerType.String(), "always_on").Return("always_on")
				cfg.EXPECT().GetBool(configTracesSamplerParent.String(), true).Return(true)
			},
			expectType:   "always_on",
			expectParent: true,
			expectRatio:  defaultRatio,
		},
		{
			name: "returns custom sampler config",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().Get(configTracesSamplerRatio.String(), defaultRatio).Return(0.75)
				cfg.EXPECT().GetString(configTracesSamplerType.String(), "always_on").Return("traceidratio")
				cfg.EXPECT().GetBool(configTracesSamplerParent.String(), true).Return(false)
			},
			expectType:   "traceidratio",
			expectParent: false,
			expectRatio:  0.75,
		},
		{
			name: "handles non-float64 ratio value",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().Get(configTracesSamplerRatio.String(), defaultRatio).Return("invalid")
				cfg.EXPECT().GetString(configTracesSamplerType.String(), "always_on").Return("always_off")
				cfg.EXPECT().GetBool(configTracesSamplerParent.String(), true).Return(true)
			},
			expectType:   "always_off",
			expectParent: true,
			expectRatio:  defaultRatio,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := configmocks.NewConfig(t)
			tt.setupMock(mockConfig)

			app := &Application{config: mockConfig}
			cfg := app.samplerConfig()

			assert.Equal(t, tt.expectType, cfg.samplerType)
			assert.Equal(t, tt.expectParent, cfg.parentBased)
			assert.Equal(t, tt.expectRatio, cfg.ratio)
		})
	}
}

func TestApplication_createExporter(t *testing.T) {
	tests := []struct {
		name         string
		exporterName string
		setupMock    func(*configmocks.Config)
		expectNil    bool
		expectError  bool
	}{
		{
			name:         "creates console exporter",
			exporterName: "console",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configExporterDriver.With("console"), "console").Return("console")
			},
			expectNil: false,
		},
		{
			name:         "creates otlp exporter",
			exporterName: "otlp",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configExporterDriver.With("otlp"), "otlp").Return("otlp")
				cfg.EXPECT().GetString(configExporterTracesProtocol.With("otlp"), "").Return("")
				cfg.EXPECT().GetString(configExporterProtocol.With("otlp"), protocolHTTPProtobuf).Return(protocolHTTPProtobuf)
				cfg.EXPECT().GetInt(configExporterTracesTimeout.With("otlp"), mock.AnythingOfType("int")).Return(5000)
				cfg.EXPECT().GetInt(configExporterTimeout.With("otlp"), defaultTimeout).Return(10000)
				cfg.EXPECT().GetString(configExporterTracesHeaders.With("otlp")).Return("X-Api-Key=test")
				cfg.EXPECT().GetString(configExporterEndpoint.With("otlp")).Return("localhost:4318")
				cfg.EXPECT().GetBool(configExporterInsecure.With("otlp")).Return(true)
			},
			expectNil: false,
		},
		{
			name:         "creates zipkin exporter",
			exporterName: "zipkin",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configExporterDriver.With("zipkin"), "zipkin").Return("zipkin")
				cfg.EXPECT().GetString(configExporterEndpoint.With("zipkin")).Return("http://localhost:9411/api/v2/spans")
			},
			expectNil: false,
		},
		{
			name:         "returns error for unknown exporter",
			exporterName: "unknown",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configExporterDriver.With("unknown"), "unknown").Return("unknown")
			},
			expectError: true,
		},
		{
			name:         "uses custom driver from config",
			exporterName: "custom",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configExporterDriver.With("custom"), "custom").Return("console")
			},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := configmocks.NewConfig(t)
			tt.setupMock(mockConfig)

			app := &Application{config: mockConfig}
			ctx := context.Background()

			exp, err := app.createExporter(ctx, tt.exporterName)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.expectNil {
				assert.Nil(t, exp)
			} else {
				assert.NotNil(t, exp)
			}
		})
	}
}

func TestApplication_otlpConfig(t *testing.T) {
	tests := []struct {
		name           string
		exporterName   string
		setupMock      func(*configmocks.Config)
		expectEndpoint string
		expectProtocol string
		expectInsecure bool
		expectTimeout  int
		expectHeaders  map[string]string
	}{
		{
			name:         "returns default otlp config",
			exporterName: "otlp",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configExporterTracesProtocol.With("otlp"), "").Return("")
				cfg.EXPECT().GetString(configExporterProtocol.With("otlp"), protocolHTTPProtobuf).Return(protocolHTTPProtobuf)
				cfg.EXPECT().GetInt(configExporterTracesTimeout.With("otlp"), mock.AnythingOfType("int")).Return(10000)
				cfg.EXPECT().GetInt(configExporterTimeout.With("otlp"), defaultTimeout).Return(10000)
				cfg.EXPECT().GetString(configExporterTracesHeaders.With("otlp")).Return("")
				cfg.EXPECT().GetString(configExporterEndpoint.With("otlp")).Return("localhost:4318")
				cfg.EXPECT().GetBool(configExporterInsecure.With("otlp")).Return(false)
			},
			expectEndpoint: "localhost:4318",
			expectProtocol: protocolHTTPProtobuf,
			expectInsecure: false,
			expectTimeout:  10000,
			expectHeaders:  map[string]string{},
		},
		{
			name:         "returns custom otlp config with grpc protocol",
			exporterName: "custom-otlp",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configExporterTracesProtocol.With("custom-otlp"), "").Return(protocolGRPC)
				cfg.EXPECT().GetInt(configExporterTracesTimeout.With("custom-otlp"), mock.AnythingOfType("int")).Return(5000)
				cfg.EXPECT().GetInt(configExporterTimeout.With("custom-otlp"), defaultTimeout).Return(10000)
				cfg.EXPECT().GetString(configExporterTracesHeaders.With("custom-otlp")).Return("Authorization=Bearer token,X-Tenant=tenant1")
				cfg.EXPECT().GetString(configExporterEndpoint.With("custom-otlp")).Return("otel-collector:4317")
				cfg.EXPECT().GetBool(configExporterInsecure.With("custom-otlp")).Return(true)
			},
			expectEndpoint: "otel-collector:4317",
			expectProtocol: protocolGRPC,
			expectInsecure: true,
			expectTimeout:  5000,
			expectHeaders:  map[string]string{"Authorization": "Bearer token", "X-Tenant": "tenant1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := configmocks.NewConfig(t)
			tt.setupMock(mockConfig)

			app := &Application{config: mockConfig}
			cfg := app.otlpConfig(tt.exporterName)

			assert.Equal(t, tt.expectEndpoint, cfg.endpoint)
			assert.Equal(t, tt.expectProtocol, cfg.protocol)
			assert.Equal(t, tt.expectInsecure, cfg.insecure)
			assert.Equal(t, tt.expectTimeout, cfg.timeout)
			assert.Equal(t, tt.expectHeaders, cfg.headers)
		})
	}
}

func TestApplication_zipkinConfig(t *testing.T) {
	tests := []struct {
		name           string
		exporterName   string
		setupMock      func(*configmocks.Config)
		expectEndpoint string
	}{
		{
			name:         "returns zipkin config with custom endpoint",
			exporterName: "zipkin",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configExporterEndpoint.With("zipkin")).Return("http://zipkin:9411/api/v2/spans")
			},
			expectEndpoint: "http://zipkin:9411/api/v2/spans",
		},
		{
			name:         "returns zipkin config with empty endpoint",
			exporterName: "custom-zipkin",
			setupMock: func(cfg *configmocks.Config) {
				cfg.EXPECT().GetString(configExporterEndpoint.With("custom-zipkin")).Return("")
			},
			expectEndpoint: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := configmocks.NewConfig(t)
			tt.setupMock(mockConfig)

			app := &Application{config: mockConfig}
			cfg := app.zipkinConfig(tt.exporterName)

			assert.Equal(t, tt.expectEndpoint, cfg.endpoint)
		})
	}
}
