package telemetry

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc/credentials"

	"github.com/goravel/framework/errors"
)

type ExporterDriver string

const (
	TraceExporterDriverCustom  ExporterDriver = "custom"
	TraceExporterDriverOTLP    ExporterDriver = "otlp"
	TraceExporterDriverConsole ExporterDriver = "console"
)

func NewTracerProvider(ctx context.Context, cfg Config, opts ...sdktrace.TracerProviderOption) (oteltrace.TracerProvider, ShutdownFunc, error) {
	exporterName := cfg.Traces.Exporter

	// 1. If disabled, return the true No-Op provider (Zero overhead)
	if exporterName == "" {
		tp := noop.NewTracerProvider()
		otel.SetTracerProvider(tp)
		return tp, NoopShutdown(), nil
	}

	exporterCfg, ok := cfg.GetExporter(exporterName)
	if !ok {
		return nil, NoopShutdown(), errors.TelemetryExporterNotFound
	}

	exporter, err := newTraceExporter(ctx, exporterCfg)
	if err != nil {
		return nil, NoopShutdown(), err
	}

	providerOptions := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(newTraceSampler(cfg.Traces.Sampler)),
	}
	providerOptions = append(providerOptions, opts...)

	tp := sdktrace.NewTracerProvider(providerOptions...)
	otel.SetTracerProvider(tp)

	return tp, tp.Shutdown, nil
}

func newTraceExporter(ctx context.Context, cfg ExporterEntry) (sdktrace.SpanExporter, error) {
	switch cfg.Driver {
	case TraceExporterDriverOTLP:
		return newOTLPTraceExporter(ctx, cfg)
	case TraceExporterDriverConsole:
		return newConsoleTraceExporter(cfg)
	case TraceExporterDriverCustom:
		if cfg.Via == nil {
			return nil, errors.TelemetryViaRequired
		}

		if exporter, ok := cfg.Via.(sdktrace.SpanExporter); ok {
			return exporter, nil
		}

		if factory, ok := cfg.Via.(func(context.Context) (sdktrace.SpanExporter, error)); ok {
			return factory(ctx)
		}

		return nil, errors.TelemetryTraceViaTypeMismatch.Args(fmt.Sprintf("%T", cfg.Via))
	default:
		return nil, errors.TelemetryUnsupportedDriver.Args(string(cfg.Driver))
	}
}

func newOTLPTraceExporter(ctx context.Context, cfg ExporterEntry) (sdktrace.SpanExporter, error) {
	switch cfg.Protocol {
	case ProtocolGRPC:
		opts, err := buildOTLPOptions(cfg, otlpOptions[otlptracegrpc.Option]{
			withEndpoint:    otlptracegrpc.WithEndpoint,
			withEndpointURL: otlptracegrpc.WithEndpointURL,
			withInsecure:    otlptracegrpc.WithInsecure,
			withTimeout:     otlptracegrpc.WithTimeout,
			withHeaders:     otlptracegrpc.WithHeaders,
			withCompression: func() otlptracegrpc.Option { return otlptracegrpc.WithCompressor(CompressionGzip) },
			withTLS: func(config *tls.Config) otlptracegrpc.Option {
				return otlptracegrpc.WithTLSCredentials(credentials.NewTLS(config))
			},
			withRetry: func(retry RetryConfig) otlptracegrpc.Option {
				return otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
					Enabled:         retry.IsEnabled(),
					InitialInterval: retry.InitialInterval,
					MaxInterval:     retry.MaxInterval,
					MaxElapsedTime:  retry.MaxElapsedTime,
				})
			},
		})
		if err != nil {
			return nil, err
		}
		return otlptracegrpc.New(ctx, opts...)
	case ProtocolHTTPProtobuf, "":
		opts, err := buildOTLPOptions(cfg, otlpOptions[otlptracehttp.Option]{
			withEndpoint:    otlptracehttp.WithEndpoint,
			withEndpointURL: otlptracehttp.WithEndpointURL,
			withInsecure:    otlptracehttp.WithInsecure,
			withTimeout:     otlptracehttp.WithTimeout,
			withHeaders:     otlptracehttp.WithHeaders,
			withCompression: func() otlptracehttp.Option { return otlptracehttp.WithCompression(otlptracehttp.GzipCompression) },
			withTLS:         otlptracehttp.WithTLSClientConfig,
			withRetry: func(retry RetryConfig) otlptracehttp.Option {
				return otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
					Enabled:         retry.IsEnabled(),
					InitialInterval: retry.InitialInterval,
					MaxInterval:     retry.MaxInterval,
					MaxElapsedTime:  retry.MaxElapsedTime,
				})
			},
		})
		if err != nil {
			return nil, err
		}
		return otlptracehttp.New(ctx, opts...)
	default:
		return nil, errors.TelemetryUnsupportedProtocol.Args(string(cfg.Protocol))
	}
}

func newConsoleTraceExporter(cfg ExporterEntry) (sdktrace.SpanExporter, error) {
	opts := []stdouttrace.Option{
		stdouttrace.WithWriter(os.Stdout),
	}

	if cfg.PrettyPrint {
		opts = append(opts, stdouttrace.WithPrettyPrint())
	}

	return stdouttrace.New(opts...)
}
