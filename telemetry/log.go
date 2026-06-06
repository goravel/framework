package telemetry

import (
	"context"
	"fmt"
	"os"

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

	processor, err := newLogProcessor(exporter, cfg.Logs.Processor)
	if err != nil {
		return nil, NoopShutdown(), err
	}

	providerOptions := []sdklog.LoggerProviderOption{
		sdklog.WithProcessor(processor),
	}
	providerOptions = append(providerOptions, opts...)

	lp := sdklog.NewLoggerProvider(providerOptions...)
	global.SetLoggerProvider(lp)

	return lp, lp.Shutdown, nil
}

func newLogProcessor(exporter sdklog.Exporter, cfg ProcessorConfig) (sdklog.Processor, error) {
	switch cfg.Type {
	case ProcessorSimple:
		return sdklog.NewSimpleProcessor(exporter), nil
	case ProcessorBatch, "":
		var opts []sdklog.BatchProcessorOption
		if cfg.Interval > 0 {
			opts = append(opts, sdklog.WithExportInterval(cfg.Interval))
		}
		if cfg.Timeout > 0 {
			opts = append(opts, sdklog.WithExportTimeout(cfg.Timeout))
		}
		if cfg.MaxQueueSize > 0 {
			opts = append(opts, sdklog.WithMaxQueueSize(cfg.MaxQueueSize))
		}
		if cfg.MaxBatchSize > 0 {
			opts = append(opts, sdklog.WithExportMaxBatchSize(cfg.MaxBatchSize))
		}
		return sdklog.NewBatchProcessor(exporter, opts...), nil
	default:
		return nil, errors.TelemetryUnsupportedProcessor.Args(cfg.Type)
	}
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
