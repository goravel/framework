package telemetry

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	exporterOTLP    = "otlp"
	exporterZipkin  = "zipkin"
	exporterConsole = "console"

	protocolGRPC         = "grpc"
	protocolHTTPProtobuf = "http/protobuf"
	protocolHTTPJSON     = "http/json"

	defaultTimeout = 10000 // milliseconds
)

type otlpExporterConfig struct {
	endpoint string
	protocol string
	insecure bool
	timeout  int // milliseconds
	headers  map[string]string
}

type zipkinExporterConfig struct {
	endpoint string
}

type consoleExporterConfig struct {
	writer      io.Writer
	prettyPrint bool
}

func newOTLPTraceExporter(ctx context.Context, cfg otlpExporterConfig) (sdktrace.SpanExporter, error) {
	protocol := cfg.protocol
	if protocol == "" {
		protocol = protocolHTTPProtobuf
	}

	switch protocol {
	case protocolGRPC:
		return newOTLPGRPCTraceExporter(ctx, cfg)
	default:
		return newOTLPHTTPTraceExporter(ctx, cfg)
	}
}

func newOTLPGRPCTraceExporter(ctx context.Context, cfg otlpExporterConfig) (sdktrace.SpanExporter, error) {
	var opts []otlptracegrpc.Option

	endpoint := cfg.endpoint
	if endpoint != "" {
		endpoint = strings.TrimPrefix(endpoint, "http://")
		endpoint = strings.TrimPrefix(endpoint, "https://")
		opts = append(opts, otlptracegrpc.WithEndpoint(endpoint))
	}

	if cfg.insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	timeout := cfg.timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	opts = append(opts, otlptracegrpc.WithTimeout(time.Duration(timeout)*time.Millisecond))

	if len(cfg.headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(cfg.headers))
	}

	return otlptracegrpc.New(ctx, opts...)
}

func newOTLPHTTPTraceExporter(ctx context.Context, cfg otlpExporterConfig) (sdktrace.SpanExporter, error) {
	var opts []otlptracehttp.Option

	endpoint := cfg.endpoint
	if endpoint != "" {
		endpoint = strings.TrimPrefix(endpoint, "http://")
		endpoint = strings.TrimPrefix(endpoint, "https://")
		opts = append(opts, otlptracehttp.WithEndpoint(endpoint))
	}

	if cfg.insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	timeout := cfg.timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	opts = append(opts, otlptracehttp.WithTimeout(time.Duration(timeout)*time.Millisecond))

	if len(cfg.headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.headers))
	}

	return otlptracehttp.New(ctx, opts...)
}

func newZipkinTraceExporter(cfg zipkinExporterConfig) (sdktrace.SpanExporter, error) {
	endpoint := cfg.endpoint
	if endpoint == "" {
		endpoint = "http://localhost:9411/api/v2/spans"
	}
	return zipkin.New(endpoint)
}

func newConsoleTraceExporter(cfg consoleExporterConfig) (sdktrace.SpanExporter, error) {
	var opts []stdouttrace.Option

	writer := cfg.writer
	if writer == nil {
		writer = os.Stdout
	}
	opts = append(opts, stdouttrace.WithWriter(writer))

	if cfg.prettyPrint {
		opts = append(opts, stdouttrace.WithPrettyPrint())
	}

	return stdouttrace.New(opts...)
}

// parseHeaders parses a comma-separated string of key=value pairs into a map.
// Format: "key1=value1,key2=value2"
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
