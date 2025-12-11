package telemetry

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/log/noop"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	"github.com/goravel/framework/errors"
)

const (
	LogExporterDriverCustom  ExporterDriver = "custom"
	LogExporterDriverOTLP    ExporterDriver = "otlp"
	LogExporterDriverConsole ExporterDriver = "console"
)

const (
	defaultLogExportInterval = 1 * time.Second
	defaultLogExportTimeout  = 30 * time.Second
)

func NewLoggerProvider(ctx context.Context, cfg Config, opts ...sdklog.LoggerProviderOption) (log.LoggerProvider, ShutdownFunc, error) {
	exporterName := cfg.Logs.Exporter
	if exporterName == "" {
		lp := noop.NewLoggerProvider()
		global.SetLoggerProvider(lp)
		return lp, NoopShutdown(), nil
	}

	exporterCfg, ok := cfg.GetExporter(exporterName)
	if !ok {
		return nil, NoopShutdown(), errors.TelemetryExporterNotFound
	}

	exporter, err := newLogExporter(ctx, exporterCfg)
	if err != nil {
		return nil, NoopShutdown(), err
	}

	interval := cfg.Logs.Processor.Interval
	timeout := cfg.Logs.Processor.Timeout

	if interval == 0 {
		interval = defaultLogExportInterval
	}
	if timeout == 0 {
		timeout = defaultLogExportTimeout
	}

	processorOptions := []sdklog.BatchProcessorOption{
		sdklog.WithExportInterval(interval),
		sdklog.WithExportTimeout(timeout),
	}

	providerOptions := []sdklog.LoggerProviderOption{
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter, processorOptions...)),
	}
	providerOptions = append(providerOptions, opts...)

	lp := sdklog.NewLoggerProvider(providerOptions...)
	global.SetLoggerProvider(lp)

	return lp, lp.Shutdown, nil
}

func newLogExporter(ctx context.Context, cfg ExporterEntry) (sdklog.Exporter, error) {
	switch cfg.Driver {
	case LogExporterDriverOTLP:
		return newOTLPLogExporter(ctx, cfg)
	case LogExporterDriverConsole:
		return newConsoleLogExporter()
	case LogExporterDriverCustom:
		if cfg.Via == nil {
			return nil, errors.TelemetryViaRequired
		}

		if exporter, ok := cfg.Via.(sdklog.Exporter); ok {
			return exporter, nil
		}

		if factory, ok := cfg.Via.(func(context.Context) (sdklog.Exporter, error)); ok {
			return factory(ctx)
		}

		return nil, errors.TelemetryLogViaTypeMismatch.Args(fmt.Sprintf("%T", cfg.Via))
	default:
		return nil, errors.TelemetryUnsupportedDriver.Args(string(cfg.Driver))
	}
}

func newOTLPLogExporter(ctx context.Context, cfg ExporterEntry) (sdklog.Exporter, error) {
	protocol := cfg.Protocol
	if protocol == "" {
		protocol = ProtocolHTTPProtobuf
	}

	switch protocol {
	case ProtocolGRPC:
		opts := buildOTLPLogOptions[otlploggrpc.Option](cfg,
			otlploggrpc.WithEndpoint,
			otlploggrpc.WithInsecure,
			otlploggrpc.WithTimeout,
			otlploggrpc.WithHeaders,
		)
		return otlploggrpc.New(ctx, opts...)
	default:
		opts := buildOTLPLogOptions[otlploghttp.Option](cfg,
			otlploghttp.WithEndpoint,
			otlploghttp.WithInsecure,
			otlploghttp.WithTimeout,
			otlploghttp.WithHeaders,
		)
		return otlploghttp.New(ctx, opts...)
	}
}

func buildOTLPLogOptions[T any](
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

	timeout := defaultLogExportTimeout
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}
	opts = append(opts, withTimeout(timeout))

	if headers := cfg.Headers; len(headers) > 0 {
		opts = append(opts, withHeaders(headers))
	}

	return opts
}

func newConsoleLogExporter() (sdklog.Exporter, error) {
	return stdoutlog.New(
		stdoutlog.WithWriter(os.Stdout),
		stdoutlog.WithPrettyPrint(),
	)
}
