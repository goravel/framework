package telemetry

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc/credentials"

	"github.com/goravel/framework/errors"
)

const (
	MetricsExporterDriverCustom  ExporterDriver = "custom"
	MetricsExporterDriverOTLP    ExporterDriver = "otlp"
	MetricsExporterDriverConsole ExporterDriver = "console"
)

type MetricTemporality string

const (
	TemporalityCumulative MetricTemporality = "cumulative"
	TemporalityDelta      MetricTemporality = "delta"
	TemporalityLowMemory  MetricTemporality = "lowmemory"
)

const (
	// defaultMetricExportInterval is the default duration for which the PeriodicReader
	// will wait between collection and export cycles (60 seconds).
	defaultMetricExportInterval = 60 * time.Second

	// defaultMetricExportTimeout is the default maximum duration the PeriodicReader
	// allows for a single export operation to complete (30 seconds).
	defaultMetricExportTimeout = 30 * time.Second
)

func NewMeterProvider(ctx context.Context, cfg Config, opts ...sdkmetric.Option) (metric.MeterProvider, ShutdownFunc, error) {
	exporterName := cfg.Metrics.Exporter
	if exporterName == "" {
		mp := noop.NewMeterProvider()
		otel.SetMeterProvider(mp)
		return mp, NoopShutdown(), nil
	}

	exporterCfg, ok := cfg.GetExporter(exporterName)
	if !ok {
		return nil, NoopShutdown(), errors.TelemetryExporterNotFound
	}

	reader, err := newMetricReader(ctx, exporterCfg, cfg.Metrics.Reader)
	if err != nil {
		return nil, NoopShutdown(), err
	}

	providerOptions := []sdkmetric.Option{
		sdkmetric.WithReader(reader),
	}
	providerOptions = append(providerOptions, opts...)

	mp := sdkmetric.NewMeterProvider(providerOptions...)
	otel.SetMeterProvider(mp)

	return mp, mp.Shutdown, nil
}

func newMetricReader(ctx context.Context, cfg ExporterEntry, readerCfg MetricsReaderConfig) (sdkmetric.Reader, error) {
	interval := readerCfg.Interval
	timeout := readerCfg.Timeout

	if interval == 0 {
		interval = defaultMetricExportInterval
	}
	if timeout == 0 {
		timeout = defaultMetricExportTimeout
	}

	periodicOptions := []sdkmetric.PeriodicReaderOption{
		sdkmetric.WithInterval(interval),
		sdkmetric.WithTimeout(timeout),
	}

	switch cfg.Driver {
	case MetricsExporterDriverOTLP:
		exporter, err := newOTLPMetricExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}
		return sdkmetric.NewPeriodicReader(exporter, periodicOptions...), nil

	case MetricsExporterDriverConsole:
		exporter, err := newConsoleMetricExporter(cfg)
		if err != nil {
			return nil, err
		}
		return sdkmetric.NewPeriodicReader(exporter, periodicOptions...), nil

	case MetricsExporterDriverCustom:
		if cfg.Via == nil {
			return nil, errors.TelemetryViaRequired
		}

		if reader, ok := cfg.Via.(sdkmetric.Reader); ok {
			return reader, nil
		}

		if factory, ok := cfg.Via.(func(context.Context) (sdkmetric.Reader, error)); ok {
			return factory(ctx)
		}
		return nil, errors.TelemetryMetricViaTypeMismatch.Args(fmt.Sprintf("%T", cfg.Via))

	default:
		return nil, errors.TelemetryUnsupportedDriver.Args(string(cfg.Driver))
	}
}

func newOTLPMetricExporter(ctx context.Context, cfg ExporterEntry) (sdkmetric.Exporter, error) {
	temporalitySelector := getTemporalitySelector(cfg.MetricTemporality)

	switch cfg.Protocol {
	case ProtocolGRPC:
		opts, err := buildOTLPOptions(cfg, otlpOptions[otlpmetricgrpc.Option]{
			withEndpoint:    otlpmetricgrpc.WithEndpoint,
			withEndpointURL: otlpmetricgrpc.WithEndpointURL,
			withInsecure:    otlpmetricgrpc.WithInsecure,
			withTimeout:     otlpmetricgrpc.WithTimeout,
			withHeaders:     otlpmetricgrpc.WithHeaders,
			withCompression: func() otlpmetricgrpc.Option { return otlpmetricgrpc.WithCompressor(CompressionGzip) },
			withTLS: func(config *tls.Config) otlpmetricgrpc.Option {
				return otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(config))
			},
			withRetry: func(retry RetryConfig) otlpmetricgrpc.Option {
				return otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig{
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
		opts = append(opts, otlpmetricgrpc.WithTemporalitySelector(temporalitySelector))
		return otlpmetricgrpc.New(ctx, opts...)
	case ProtocolHTTPProtobuf, "":
		opts, err := buildOTLPOptions(cfg, otlpOptions[otlpmetrichttp.Option]{
			withEndpoint:    otlpmetrichttp.WithEndpoint,
			withEndpointURL: otlpmetrichttp.WithEndpointURL,
			withInsecure:    otlpmetrichttp.WithInsecure,
			withTimeout:     otlpmetrichttp.WithTimeout,
			withHeaders:     otlpmetrichttp.WithHeaders,
			withCompression: func() otlpmetrichttp.Option { return otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression) },
			withTLS:         otlpmetrichttp.WithTLSClientConfig,
			withRetry: func(retry RetryConfig) otlpmetrichttp.Option {
				return otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{
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
		opts = append(opts, otlpmetrichttp.WithTemporalitySelector(temporalitySelector))
		return otlpmetrichttp.New(ctx, opts...)
	default:
		return nil, errors.TelemetryUnsupportedProtocol.Args(string(cfg.Protocol))
	}
}

func newConsoleMetricExporter(cfg ExporterEntry) (sdkmetric.Exporter, error) {
	opts := []stdoutmetric.Option{
		stdoutmetric.WithWriter(os.Stdout),
	}

	if cfg.PrettyPrint {
		opts = append(opts, stdoutmetric.WithPrettyPrint())
	}

	return stdoutmetric.New(opts...)
}

func getTemporalitySelector(t MetricTemporality) sdkmetric.TemporalitySelector {
	switch t {
	case TemporalityDelta:
		return sdkmetric.DeltaTemporalitySelector
	case TemporalityLowMemory:
		return sdkmetric.LowMemoryTemporalitySelector
	default:
		return sdkmetric.DefaultTemporalitySelector
	}
}
