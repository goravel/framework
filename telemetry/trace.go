package telemetry

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/goravel/framework/errors"
)

type TraceExporterFactoryFunc func(ctx context.Context) (sdktrace.SpanExporter, error)

type ExporterDriver string

const (
	TraceExporterDriverCustom  ExporterDriver = "custom"
	TraceExporterDriverOTLP    ExporterDriver = "otlp"
	TraceExporterDriverZipkin  ExporterDriver = "zipkin"
	TraceExporterDriverConsole ExporterDriver = "console"
)

type Protocol string

const (
	ProtocolGRPC         Protocol = "grpc"
	ProtocolHTTPProtobuf Protocol = "http/protobuf"
	ProtocolHTTPJSON     Protocol = "http/json"
)

const defaultTimeout = 10 * time.Second

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
	case TraceExporterDriverZipkin:
		return newZipkinTraceExporter(cfg)
	case TraceExporterDriverConsole:
		return newConsoleTraceExporter()
	case TraceExporterDriverCustom:
		if cfg.Via == nil {
			return nil, errors.TelemetryViaRequired
		}

		traceFactory, ok := cfg.Via.(TraceExporterFactoryFunc)
		if !ok {
			return nil, errors.TelemetryTraceViaTypeMismatch.Args(fmt.Sprintf("%T", cfg.Via))
		}
		return traceFactory(ctx)

	default:
		return nil, errors.TelemetryUnsupportedDriver.Args(string(cfg.Driver))
	}
}

func newOTLPTraceExporter(ctx context.Context, cfg ExporterEntry) (sdktrace.SpanExporter, error) {
	protocol := cfg.Protocol
	if protocol == "" {
		protocol = ProtocolHTTPProtobuf
	}

	switch protocol {
	case ProtocolGRPC:
		opts := buildOTLPTraceOptions[otlptracegrpc.Option](cfg,
			otlptracegrpc.WithEndpoint,
			otlptracegrpc.WithInsecure,
			otlptracegrpc.WithTimeout,
			otlptracegrpc.WithHeaders,
		)
		return otlptracegrpc.New(ctx, opts...)
	default:
		opts := buildOTLPTraceOptions[otlptracehttp.Option](cfg,
			otlptracehttp.WithEndpoint,
			otlptracehttp.WithInsecure,
			otlptracehttp.WithTimeout,
			otlptracehttp.WithHeaders,
		)
		return otlptracehttp.New(ctx, opts...)
	}
}

func buildOTLPTraceOptions[T any](
	cfg ExporterEntry,
	withEndpoint func(string) T,
	withInsecure func() T,
	withTimeout func(time.Duration) T,
	withHeaders func(map[string]string) T,
) []T {
	var opts []T

	if cfg.Endpoint != "" {
		endpoint := strings.TrimPrefix(cfg.Endpoint, "http://")
		endpoint = strings.TrimPrefix(endpoint, "https://")
		opts = append(opts, withEndpoint(endpoint))
	}

	if cfg.Insecure {
		opts = append(opts, withInsecure())
	}

	timeout := defaultTimeout
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}
	opts = append(opts, withTimeout(timeout))

	if headers := parseHeaders(cfg.Headers); len(headers) > 0 {
		opts = append(opts, withHeaders(headers))
	}

	return opts
}

func newZipkinTraceExporter(cfg ExporterEntry) (sdktrace.SpanExporter, error) {
	endpoint := cfg.Endpoint
	if endpoint == "" {
		return nil, errors.TelemetryZipkinEndpointRequired
	}
	return zipkin.New(endpoint)
}

func newConsoleTraceExporter() (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(os.Stdout),
		stdouttrace.WithPrettyPrint(),
	)
}

func parseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}

	for _, pair := range strings.Split(headerStr, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		// SplitN ensures we only split on the first '=' in case the value contains '=' (e.g. base64)
		if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
			headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return headers
}
