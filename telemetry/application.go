package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/telemetry"
	"github.com/goravel/framework/errors"
)

var _ telemetry.Telemetry = (*Application)(nil)

type Application struct {
	tracerProvider trace.TracerProvider
	propagator     propagation.TextMapPropagator
}

func NewApplication(cfg config.Config) (*Application, error) {
	propagator, err := newCompositeTextMapPropagator(cfg.GetString(configPropagators.String()))
	if err != nil {
		return nil, err
	}

	otel.SetTextMapPropagator(propagator)

	exporterName := cfg.GetString(configTracesExporter.String())
	if exporterName == "" {
		return &Application{
			tracerProvider: tracenoop.NewTracerProvider(),
			propagator:     propagator,
		}, nil
	}

	ctx := context.Background()

	res, err := newResource(ctx, getResourceConfig(cfg))
	if err != nil {
		return nil, err
	}

	exp, err := createExporter(ctx, cfg, exporterName)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(newTraceSampler(getSamplerConfig(cfg))),
	)

	otel.SetTracerProvider(tp)

	return &Application{
		tracerProvider: tp,
		propagator:     propagator,
	}, nil
}

func (r *Application) Propagator() propagation.TextMapPropagator {
	return r.propagator
}

func (r *Application) Shutdown(ctx context.Context) error {
	if tp, ok := r.tracerProvider.(*sdktrace.TracerProvider); ok {
		return tp.Shutdown(ctx)
	}
	return nil
}

func (r *Application) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return r.tracerProvider.Tracer(name, opts...)
}

func (r *Application) TracerProvider() trace.TracerProvider {
	return r.tracerProvider
}

func getResourceConfig(cfg config.Config) resourceConfig {
	return resourceConfig{
		serviceName:    cfg.GetString(configServiceName.String(), "goravel"),
		serviceVersion: cfg.GetString(configServiceVersion.String()),
		environment:    cfg.GetString(configEnvironment.String()),
	}
}

func getSamplerConfig(cfg config.Config) samplerConfig {
	ratio := defaultRatio
	if v, ok := cfg.Get(configTracesSamplerRatio.String(), defaultRatio).(float64); ok {
		ratio = v
	}

	return samplerConfig{
		samplerType: cfg.GetString(configTracesSamplerType.String(), "always_on"),
		parentBased: cfg.GetBool(configTracesSamplerParent.String(), true),
		ratio:       ratio,
	}
}

func createExporter(ctx context.Context, cfg config.Config, exporterName string) (sdktrace.SpanExporter, error) {
	driver := cfg.GetString(configExporterDriver.With(exporterName), exporterName)

	switch driver {
	case exporterOTLP:
		return newOTLPTraceExporter(ctx, getOTLPConfig(cfg, exporterName))
	case exporterZipkin:
		return newZipkinTraceExporter(getZipkinConfig(cfg, exporterName))
	case exporterConsole:
		return newConsoleTraceExporter(consoleExporterConfig{prettyPrint: true})
	default:
		return nil, errors.TelemetryUnsupportedDriver.Args(driver)
	}
}

func getOTLPConfig(cfg config.Config, exporterName string) otlpExporterConfig {
	protocol := cfg.GetString(configExporterTracesProtocol.With(exporterName), "")
	if protocol == "" {
		protocol = cfg.GetString(configExporterProtocol.With(exporterName), protocolHTTPProtobuf)
	}

	timeout := cfg.GetInt(
		configExporterTracesTimeout.With(exporterName),
		cfg.GetInt(configExporterTimeout.With(exporterName), defaultTimeout),
	)
	headers := parseHeaders(cfg.GetString(configExporterTracesHeaders.With(exporterName)))

	return otlpExporterConfig{
		endpoint: cfg.GetString(configExporterEndpoint.With(exporterName)),
		protocol: protocol,
		insecure: cfg.GetBool(configExporterInsecure.With(exporterName)),
		timeout:  timeout,
		headers:  headers,
	}
}

func getZipkinConfig(cfg config.Config, exporterName string) zipkinExporterConfig {
	return zipkinExporterConfig{
		endpoint: cfg.GetString(configExporterEndpoint.With(exporterName)),
	}
}
