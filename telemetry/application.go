package telemetry

import (
	"context"

	"github.com/goravel/framework/contracts/telemetry"
	"go.opentelemetry.io/otel"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
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
	traceProvider, err := NewTracerProvider(ctx, cfg)
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
