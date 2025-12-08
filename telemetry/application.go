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
	config         config.Config
	tracerProvider trace.TracerProvider
	propagator     propagation.TextMapPropagator
}

func NewApplication(cfg config.Config) (*Application, error) {
	r := &Application{config: cfg}
	if err := r.init(); err != nil {
		return nil, err
	}
	return r, nil
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
	return r.TracerProvider().Tracer(name, opts...)
}

func (r *Application) TracerProvider() trace.TracerProvider {
	if r.tracerProvider == nil {
		return tracenoop.NewTracerProvider()
	}
	return r.tracerProvider
}

func (r *Application) init() error {
	propagator, err := newCompositeTextMapPropagator(r.config.GetString(configPropagators.String()))
	if err != nil {
		return err
	}

	r.propagator = propagator
	otel.SetTextMapPropagator(r.propagator)

	exporterName := r.config.GetString(configTracesExporter.String())
	if exporterName == "" || exporterName == "none" {
		return nil
	}

	ctx := context.Background()

	res, err := newResource(ctx, r.resourceConfig())
	if err != nil {
		return err
	}

	exp, err := r.createExporter(ctx, exporterName)
	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(newTraceSampler(r.samplerConfig())),
	)

	r.tracerProvider = tp
	otel.SetTracerProvider(tp)

	return nil
}

func (r *Application) resourceConfig() resourceConfig {
	return resourceConfig{
		serviceName:    r.config.GetString(configServiceName.String(), "goravel"),
		serviceVersion: r.config.GetString(configServiceVersion.String()),
		environment:    r.config.GetString(configEnvironment.String()),
	}
}

func (r *Application) samplerConfig() samplerConfig {
	ratio := defaultRatio
	if v, ok := r.config.Get(configTracesSamplerRatio.String(), defaultRatio).(float64); ok {
		ratio = v
	}

	return samplerConfig{
		samplerType: r.config.GetString(configTracesSamplerType.String(), "always_on"),
		parentBased: r.config.GetBool(configTracesSamplerParent.String(), true),
		ratio:       ratio,
	}
}

func (r *Application) createExporter(ctx context.Context, exporterName string) (sdktrace.SpanExporter, error) {
	driver := r.config.GetString(configExporterDriver.With(exporterName), exporterName)

	switch driver {
	case exporterOTLP:
		return newOTLPTraceExporter(ctx, r.otlpConfig(exporterName))
	case exporterZipkin:
		return newZipkinTraceExporter(r.zipkinConfig(exporterName))
	case exporterConsole:
		return newConsoleTraceExporter(consoleExporterConfig{prettyPrint: true})
	default:
		return nil, errors.TelemetryUnsupportedDriver.Args(driver)
	}
}

func (r *Application) otlpConfig(exporterName string) otlpExporterConfig {
	protocol := r.config.GetString(configExporterTracesProtocol.With(exporterName), "")
	if protocol == "" {
		protocol = r.config.GetString(configExporterProtocol.With(exporterName), protocolHTTPProtobuf)
	}

	timeout := r.config.GetInt(
		configExporterTracesTimeout.With(exporterName),
		r.config.GetInt(configExporterTimeout.With(exporterName), defaultTimeout),
	)
	headers := parseHeaders(r.config.GetString(configExporterTracesHeaders.With(exporterName)))

	return otlpExporterConfig{
		endpoint: r.config.GetString(configExporterEndpoint.With(exporterName)),
		protocol: protocol,
		insecure: r.config.GetBool(configExporterInsecure.With(exporterName)),
		timeout:  timeout,
		headers:  headers,
	}
}

func (r *Application) zipkinConfig(exporterName string) zipkinExporterConfig {
	return zipkinExporterConfig{
		endpoint: r.config.GetString(configExporterEndpoint.With(exporterName)),
	}
}
