package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/goravel/framework/errors"
)

const (
	ExporterMetricsDriverOTLP    = "otlp"
	ExporterMetricsDriverConsole = "console"
)

type MetricTemporality string

const (
	TemporalityCumulative MetricTemporality = "cumulative"
	TemporalityDelta      MetricTemporality = "delta"
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

	exporter, err := newMetricExporter(ctx, exporterCfg)
	if err != nil {
		return nil, NoopShutdown(), err
	}

	providerOptions := []sdkmetric.Option{
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
	}
	providerOptions = append(providerOptions, opts...)

	mp := sdkmetric.NewMeterProvider(providerOptions...)

	otel.SetMeterProvider(mp)

	return mp, mp.Shutdown, nil
}

func newMetricExporter(ctx context.Context, cfg ExporterEntry) (sdkmetric.Exporter, error) {
	switch cfg.Driver {
	case ExporterMetricsDriverOTLP:
		return newOTLPMetricExporter(ctx, cfg)
	case ExporterMetricsDriverConsole:
		return newConsoleMetricExporter()
	default:
		return nil, errors.TelemetryUnsupportedDriver.Args(string(cfg.Driver))
	}
}

func newOTLPMetricExporter(ctx context.Context, cfg ExporterEntry) (sdkmetric.Exporter, error) {
	panic("not implemented")
}

func newConsoleMetricExporter() (sdkmetric.Exporter, error) {
	return stdoutmetric.New(
		stdoutmetric.WithWriter(os.Stdout),
		stdoutmetric.WithPrettyPrint(),
	)
}
