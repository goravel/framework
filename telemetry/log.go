package telemetry

import (
	"context"
	"fmt"
	"os"
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
	LogProcessorSimple = "simple"
	LogProcessorBatch  = "batch"
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

	var processor sdklog.Processor
	if cfg.Logs.Processor.Type == LogProcessorSimple {
		processor = sdklog.NewSimpleProcessor(exporter)
	} else {
		interval := cfg.Logs.Processor.Interval
		timeout := cfg.Logs.Processor.Timeout

		if interval == 0 {
			interval = defaultLogExportInterval
		}
		if timeout == 0 {
			timeout = defaultLogExportTimeout
		}

		batchOptions := []sdklog.BatchProcessorOption{
			sdklog.WithExportInterval(interval),
			sdklog.WithExportTimeout(timeout),
		}
		processor = sdklog.NewBatchProcessor(exporter, batchOptions...)
	}

	providerOptions := []sdklog.LoggerProviderOption{
		sdklog.WithProcessor(processor),
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
		return newConsoleLogExporter(cfg)
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
		opts := buildOTLPOptions[otlploggrpc.Option](cfg,
			otlploggrpc.WithEndpoint,
			otlploggrpc.WithInsecure,
			otlploggrpc.WithTimeout,
			otlploggrpc.WithHeaders,
		)
		return otlploggrpc.New(ctx, opts...)
	default:
		opts := buildOTLPOptions[otlploghttp.Option](cfg,
			otlploghttp.WithEndpoint,
			otlploghttp.WithInsecure,
			otlploghttp.WithTimeout,
			otlploghttp.WithHeaders,
		)
		return otlploghttp.New(ctx, opts...)
	}
}

func newConsoleLogExporter(cfg ExporterEntry) (sdklog.Exporter, error) {
	opts := []stdoutlog.Option{
		stdoutlog.WithWriter(os.Stdout),
	}

	if cfg.PrettyPrint {
		opts = append(opts, stdoutlog.WithPrettyPrint())
	}

	return stdoutlog.New(opts...)
}
