package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/goravel/framework/contracts/telemetry"
	"github.com/goravel/framework/errors"
)

var _ telemetry.Telemetry = (*Application)(nil)

type Application struct {
	tracerProvider trace.TracerProvider
	propagator     propagation.TextMapPropagator
}

func NewApplication(cfg Config) (*Application, error) {
	propagator, err := newCompositeTextMapPropagator(cfg.Propagators)
	if err != nil {
		return nil, err
	}

	otel.SetTextMapPropagator(propagator)

	exporterName := cfg.Traces.Exporter
	if exporterName == "" {
		return &Application{
			tracerProvider: tracenoop.NewTracerProvider(),
			propagator:     propagator,
		}, nil
	}

	ctx := context.Background()

	res, err := newResource(ctx, cfg.Service)
	if err != nil {
		return nil, err
	}

	exp, err := createTraceExporter(ctx, cfg)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(newTraceSampler(cfg.Traces.Sampler)),
	)

	otel.SetTracerProvider(tp)

	return &Application{
		tracerProvider: tp,
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

func (r *Application) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return r.tracerProvider.Tracer(name, opts...)
}

func (r *Application) TracerProvider() trace.TracerProvider {
	return r.tracerProvider
}

func createTraceExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	exporterName := cfg.Traces.Exporter
	exporterCfg, ok := cfg.GetExporter(exporterName)
	if !ok {
		return nil, errors.TelemetryExporterNotFound
	}

	switch exporterCfg.Driver {
	case ExporterDriverOTLP:
		return newOTLPTraceExporter(ctx, exporterCfg)
	case ExporterDriverZipkin:
		return newZipkinTraceExporter(exporterCfg)
	case ExporterDriverConsole:
		return newConsoleTraceExporter()
	default:
		return nil, errors.TelemetryUnsupportedDriver.Args(string(exporterCfg.Driver))
	}
}
