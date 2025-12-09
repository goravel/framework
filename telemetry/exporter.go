package telemetry

import (
	"context"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ExporterDriver string

const (
	ExporterTraceDriverOTLP      ExporterDriver = "otlp"
	ExporterTraceDriverZipkin    ExporterDriver = "zipkin"
	ExporterTraceDriverConsole   ExporterDriver = "console"
	ExporterMetricsDriverOTLP    ExporterDriver = "otlp"
	ExporterMetricsDriverConsole ExporterDriver = "console"
)

type Protocol string

const (
	ProtocolGRPC         Protocol = "grpc"
	ProtocolHTTPProtobuf Protocol = "http/protobuf"
	ProtocolHTTPJSON     Protocol = "http/json"
)

const (
	defaultTimeout = 10000 // milliseconds
)

func newOTLPTraceExporter(ctx context.Context, cfg ExporterEntry) (sdktrace.SpanExporter, error) {
	protocol := cfg.Protocol
	if protocol == "" {
		protocol = ProtocolHTTPProtobuf
	}

	switch protocol {
	case ProtocolGRPC:
		return newOTLPGRPCTraceExporter(ctx, cfg)
	default:
		return newOTLPHTTPTraceExporter(ctx, cfg)
	}
}

func newOTLPGRPCTraceExporter(ctx context.Context, cfg ExporterEntry) (sdktrace.SpanExporter, error) {
	var opts []otlptracegrpc.Option

	if cfg.Endpoint != "" {
		endpoint := strings.TrimPrefix(cfg.Endpoint, "http://")
		endpoint = strings.TrimPrefix(endpoint, "https://")
		opts = append(opts, otlptracegrpc.WithEndpoint(endpoint))
	}

	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	opts = append(opts, otlptracegrpc.WithTimeout(time.Duration(timeout)*time.Millisecond))

	if headers := parseHeaders(cfg.Headers); len(headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(headers))
	}

	return otlptracegrpc.New(ctx, opts...)
}

func newOTLPHTTPTraceExporter(ctx context.Context, cfg ExporterEntry) (sdktrace.SpanExporter, error) {
	var opts []otlptracehttp.Option

	if cfg.Endpoint != "" {
		endpoint := strings.TrimPrefix(cfg.Endpoint, "http://")
		endpoint = strings.TrimPrefix(endpoint, "https://")
		opts = append(opts, otlptracehttp.WithEndpoint(endpoint))
	}

	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	opts = append(opts, otlptracehttp.WithTimeout(time.Duration(timeout)*time.Millisecond))

	if headers := parseHeaders(cfg.Headers); len(headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}

	return otlptracehttp.New(ctx, opts...)
}

func newZipkinTraceExporter(cfg ExporterEntry) (sdktrace.SpanExporter, error) {
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "http://localhost:9411/api/v2/spans"
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
		if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
			headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	return headers
}
