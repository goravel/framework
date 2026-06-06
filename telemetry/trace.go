package telemetry

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

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

	processorOption, err := newTraceProcessorOption(exporter, cfg.Traces.Processor)
	if err != nil {
		return nil, NoopShutdown(), err
	}

	providerOptions := []sdktrace.TracerProviderOption{
		processorOption,
		sdktrace.WithSampler(newTraceSampler(cfg.Traces.Sampler)),
	}

	if limitsOption, ok := newSpanLimitsOption(cfg.Traces.Limits); ok {
		providerOptions = append(providerOptions, limitsOption)
	}

	providerOptions = append(providerOptions, opts...)

	tp := sdktrace.NewTracerProvider(providerOptions...)
	otel.SetTracerProvider(tp)

	return tp, tp.Shutdown, nil
}

func newTraceProcessorOption(exporter sdktrace.SpanExporter, cfg ProcessorConfig) (sdktrace.TracerProviderOption, error) {
	switch cfg.Type {
	case ProcessorSimple:
		return sdktrace.WithSyncer(exporter), nil
	case ProcessorBatch, "":
		var opts []sdktrace.BatchSpanProcessorOption
		if cfg.Interval > 0 {
			opts = append(opts, sdktrace.WithBatchTimeout(cfg.Interval))
		}
		if cfg.Timeout > 0 {
			opts = append(opts, sdktrace.WithExportTimeout(cfg.Timeout))
		}
		if cfg.MaxQueueSize > 0 {
			opts = append(opts, sdktrace.WithMaxQueueSize(cfg.MaxQueueSize))
		}
		if cfg.MaxBatchSize > 0 {
			opts = append(opts, sdktrace.WithMaxExportBatchSize(cfg.MaxBatchSize))
		}
		return sdktrace.WithBatcher(exporter, opts...), nil
	default:
		return nil, errors.TelemetryUnsupportedProcessor.Args(cfg.Type)
	}
}

func newSpanLimitsOption(cfg SpanLimitsConfig) (sdktrace.TracerProviderOption, bool) {
	if cfg == (SpanLimitsConfig{}) {
		return nil, false
	}

	limits := sdktrace.NewSpanLimits()
	if cfg.AttributeValueLength != 0 {
		limits.AttributeValueLengthLimit = cfg.AttributeValueLength
	}
	if cfg.AttributeCount != 0 {
		limits.AttributeCountLimit = cfg.AttributeCount
	}
	if cfg.EventCount != 0 {
		limits.EventCountLimit = cfg.EventCount
	}
	if cfg.LinkCount != 0 {
		limits.LinkCountLimit = cfg.LinkCount
	}
	if cfg.AttributePerEventCount != 0 {
		limits.AttributePerEventCountLimit = cfg.AttributePerEventCount
	}
	if cfg.AttributePerLinkCount != 0 {
		limits.AttributePerLinkCountLimit = cfg.AttributePerLinkCount
	}

	return sdktrace.WithRawSpanLimits(limits), true
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
	protocol := cfg.Protocol
	if protocol == "" {
		protocol = ProtocolHTTPProtobuf
	}

	switch protocol {
	case ProtocolGRPC:
		opts := buildOTLPOptions(cfg,
			otlptracegrpc.WithEndpoint,
			otlptracegrpc.WithInsecure,
			otlptracegrpc.WithTimeout,
			otlptracegrpc.WithHeaders,
		)
		return otlptracegrpc.New(ctx, opts...)
	default:
		opts := buildOTLPOptions(cfg,
			otlptracehttp.WithEndpoint,
			otlptracehttp.WithInsecure,
			otlptracehttp.WithTimeout,
			otlptracehttp.WithHeaders,
		)
		return otlptracehttp.New(ctx, opts...)
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
