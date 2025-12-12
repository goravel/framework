package telemetry

import (
	"context"

	"github.com/goravel/framework/contracts/telemetry"
	"go.opentelemetry.io/otel"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/goravel/framework/errors"
)

var _ telemetry.Telemetry = (*Application)(nil)

type Application struct {
	meterProvider  otelmetric.MeterProvider
	propagator     propagation.TextMapPropagator
	tracerProvider oteltrace.TracerProvider
	shutdownFuncs  []ShutdownFunc
}

func NewApplication(cfg Config) (*Application, error) {
	propagator, err := newCompositeTextMapPropagator(cfg.Propagators)
	if err != nil {
		return nil, err
	}

	otel.SetTextMapPropagator(propagator)

	ctx := context.Background()
	resource, err := newResource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	traceProvider, traceShutdown, err := NewTracerProvider(ctx, cfg, sdktrace.WithResource(resource))
	if err != nil {
		return nil, err
	}

	meterProvider, metricShutdown, err := NewMeterProvider(ctx, cfg, sdkmetric.WithResource(resource))
	if err != nil {
		_ = traceShutdown(ctx) // Ensure tracer provider is shut down to avoid resource leak
		return nil, err
	}

	return &Application{
		meterProvider:  meterProvider,
		tracerProvider: traceProvider,
		propagator:     propagator,
		shutdownFuncs: []ShutdownFunc{
			traceShutdown,
			metricShutdown,
		},
	}, nil
}

func (r *Application) Meter(name string, opts ...otelmetric.MeterOption) otelmetric.Meter {
	return r.meterProvider.Meter(name, opts...)
}

func (r *Application) Propagator() propagation.TextMapPropagator {
	return r.propagator
}

func (r *Application) Shutdown(ctx context.Context) error {
	var err error

	for _, fn := range r.shutdownFuncs {
		if fn == nil {
			continue
		}
		if e := fn(ctx); e != nil {
			err = errors.Join(err, e)
		}
	}

	return err
}

func (r *Application) Tracer(name string, opts ...oteltrace.TracerOption) oteltrace.Tracer {
	return r.tracerProvider.Tracer(name, opts...)
}
