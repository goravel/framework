package telemetry

import (
	"context"
	"crypto/tls"
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
	"google.golang.org/grpc/credentials"

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

func NewLoggerProvider(ctx context.Context, cfg Config, opts ...sdklog.LoggerProviderOption) (log.LoggerProvider, ShutdownFunc, FlushFunc, error) {
	exporterName := cfg.Logs.Exporter
	if exporterName == "" {
		lp := noop.NewLoggerProvider()
		global.SetLoggerProvider(lp)
		return lp, NoopShutdown(), NoopFlush(), nil
	}

	exporterCfg, ok := cfg.GetExporter(exporterName)
	if !ok {
		return nil, NoopShutdown(), NoopFlush(), errors.TelemetryExporterNotFound
	}

	exporter, err := newLogExporter(ctx, exporterCfg)
	if err != nil {
		return nil, NoopShutdown(), NoopFlush(), err
	}

	processor, err := newLogProcessor(exporter, cfg.Logs.Processor)
	if err != nil {
		return nil, NoopShutdown(), NoopFlush(), err
	}

	providerOptions := []sdklog.LoggerProviderOption{
		sdklog.WithProcessor(processor),
	}
	providerOptions = append(providerOptions, opts...)

	lp := sdklog.NewLoggerProvider(providerOptions...)
	global.SetLoggerProvider(lp)

	return lp, lp.Shutdown, lp.ForceFlush, nil
}

func newLogProcessor(exporter sdklog.Exporter, cfg ProcessorConfig) (sdklog.Processor, error) {
	switch cfg.Type {
	case ProcessorSimple:
		return sdklog.NewSimpleProcessor(exporter), nil
	case ProcessorBatch, "":
		interval := cfg.Interval
		if interval == 0 {
			interval = defaultLogExportInterval
		}
		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = defaultLogExportTimeout
		}
		return sdklog.NewBatchProcessor(exporter,
			sdklog.WithExportInterval(interval),
			sdklog.WithExportTimeout(timeout),
		), nil
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
	switch cfg.Protocol {
	case ProtocolGRPC:
		opts, err := buildOTLPOptions(cfg, otlpOptions[otlploggrpc.Option]{
			withEndpoint:    otlploggrpc.WithEndpoint,
			withInsecure:    otlploggrpc.WithInsecure,
			withTimeout:     otlploggrpc.WithTimeout,
			withHeaders:     otlploggrpc.WithHeaders,
			withCompression: func() otlploggrpc.Option { return otlploggrpc.WithCompressor(string(CompressionGzip)) },
			withTLS: func(config *tls.Config) otlploggrpc.Option {
				return otlploggrpc.WithTLSCredentials(credentials.NewTLS(config))
			},
			withRetry: func(retry RetryConfig) otlploggrpc.Option {
				return otlploggrpc.WithRetry(otlploggrpc.RetryConfig{
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
		return otlploggrpc.New(ctx, opts...)
	case ProtocolHTTPProtobuf, "":
		opts, err := buildOTLPOptions(cfg, otlpOptions[otlploghttp.Option]{
			withEndpoint:    otlploghttp.WithEndpoint,
			withURLPath:     otlploghttp.WithURLPath,
			withInsecure:    otlploghttp.WithInsecure,
			withTimeout:     otlploghttp.WithTimeout,
			withHeaders:     otlploghttp.WithHeaders,
			withCompression: func() otlploghttp.Option { return otlploghttp.WithCompression(otlploghttp.GzipCompression) },
			withTLS:         otlploghttp.WithTLSClientConfig,
			withRetry: func(retry RetryConfig) otlploghttp.Option {
				return otlploghttp.WithRetry(otlploghttp.RetryConfig{
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
		return otlploghttp.New(ctx, opts...)
	default:
		return nil, errors.TelemetryUnsupportedProtocol.Args(string(cfg.Protocol))
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
