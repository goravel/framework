package telemetry

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	otellog "go.opentelemetry.io/otel/log"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/goravel/framework/contracts/telemetry"
	"github.com/goravel/framework/support/color"
)

var _ telemetry.Telemetry = (*Application)(nil)

// errorHandlerOnce ensures the default error handler is set once, so a custom
// handler installed by the application is not overwritten on Restart.
var errorHandlerOnce sync.Once

type Application struct {
	loggerProvider otellog.LoggerProvider
	meterProvider  otelmetric.MeterProvider
	tracerProvider oteltrace.TracerProvider
	propagator     propagation.TextMapPropagator
	shutdownFuncs  []ShutdownFunc
	flushFuncs     []FlushFunc
}

func NewApplication(cfg Config) (*Application, error) {
	propagator, err := newCompositeTextMapPropagator(cfg.Propagators)
	if err != nil {
		return nil, err
	}
	otel.SetTextMapPropagator(propagator)
	errorHandlerOnce.Do(func() {
		otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
			color.Warningln("[Telemetry]", err)
		}))
	})

	ctx := context.Background()
	resource, err := newResource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	traceProvider, traceShutdown, traceFlush, err := NewTracerProvider(ctx, cfg, sdktrace.WithResource(resource))
	if err != nil {
		return nil, err
	}

	meterProvider, metricShutdown, metricFlush, err := NewMeterProvider(ctx, cfg, sdkmetric.WithResource(resource))
	if err != nil {
		_ = traceShutdown(ctx)
		return nil, err
	}

	loggerProvider, loggerShutdown, loggerFlush, err := NewLoggerProvider(ctx, cfg, sdklog.WithResource(resource))
	if err != nil {
		_ = traceShutdown(ctx)
		_ = metricShutdown(ctx)
		return nil, err
	}

	return &Application{
		loggerProvider: loggerProvider,
		meterProvider:  meterProvider,
		tracerProvider: traceProvider,
		propagator:     propagator,
		shutdownFuncs: []ShutdownFunc{
			traceShutdown,
			metricShutdown,
			loggerShutdown,
		},
		flushFuncs: []FlushFunc{
			traceFlush,
			metricFlush,
			loggerFlush,
		},
	}, nil
}

func (r *Application) ForceFlush(ctx context.Context) error {
	return callAll(ctx, r.flushFuncs)
}

func (r *Application) Logger(name string, opts ...otellog.LoggerOption) otellog.Logger {
	return r.loggerProvider.Logger(name, opts...)
}

func (r *Application) Meter(name string, opts ...otelmetric.MeterOption) otelmetric.Meter {
	return r.meterProvider.Meter(name, opts...)
}

func (r *Application) MeterProvider() otelmetric.MeterProvider {
	return r.meterProvider
}

func (r *Application) Propagator() propagation.TextMapPropagator {
	return r.propagator
}

func (r *Application) Shutdown(ctx context.Context) error {
	return callAll(ctx, r.shutdownFuncs)
}

func (r *Application) Tracer(name string, opts ...oteltrace.TracerOption) oteltrace.Tracer {
	return r.tracerProvider.Tracer(name, opts...)
}

func (r *Application) TracerProvider() oteltrace.TracerProvider {
	return r.tracerProvider
}
