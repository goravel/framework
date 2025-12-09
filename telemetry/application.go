package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/goravel/framework/contracts/telemetry"
	"github.com/goravel/framework/errors"
)

var _ telemetry.Telemetry = (*Application)(nil)

type Application struct {
	meterProvider  otelmetric.MeterProvider
	propagator     propagation.TextMapPropagator
	tracerProvider oteltrace.TracerProvider
}

func NewApplication(cfg Config) (*Application, error) {
	propagator, err := newCompositeTextMapPropagator(cfg.Propagators)
	if err != nil {
		return nil, err
	}

	otel.SetTextMapPropagator(propagator)

	ctx := context.Background()

	traceProvider, err := createTraceProvider(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Application{
		tracerProvider: traceProvider,
		propagator:     propagator,
	}, nil
}

func (r *Application) Meter(name string, opts ...otelmetric.MeterOption) otelmetric.Meter {
	panic("not implemented")
}

func (r *Application) MeterProvider() otelmetric.MeterProvider {
	panic("not implemented")
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

func (r *Application) Tracer(name string, opts ...oteltrace.TracerOption) oteltrace.Tracer {
	return r.tracerProvider.Tracer(name, opts...)
}

func (r *Application) TracerProvider() oteltrace.TracerProvider {
	return r.tracerProvider
}

func createTraceProvider(ctx context.Context, cfg Config) (oteltrace.TracerProvider, error) {
	exporterName := cfg.Traces.Exporter
	if exporterName == "" {
		return tracenoop.NewTracerProvider(), nil
	}

	exporterCfg, ok := cfg.GetExporter(exporterName)
	if !ok {
		return nil, errors.TelemetryExporterNotFound
	}

	var exporter sdktrace.SpanExporter
	var err error
	switch exporterCfg.Driver {
	case ExporterTraceDriverOTLP:
		exporter, err = newOTLPTraceExporter(ctx, exporterCfg)
	case ExporterTraceDriverZipkin:
		exporter, err = newZipkinTraceExporter(exporterCfg)
	case ExporterTraceDriverConsole:
		exporter, err = newConsoleTraceExporter()
	default:
		err = errors.TelemetryUnsupportedDriver.Args(string(exporterCfg.Driver))
	}
	if err != nil {
		return nil, err
	}

	res, err := newResource(ctx, cfg.Service)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(newTraceSampler(cfg.Traces.Sampler)),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}
