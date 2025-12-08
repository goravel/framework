package telemetry

import "github.com/goravel/framework/support/config"

// Service identification keys
const (
	configServiceName    config.Key = "telemetry.service.name"
	configServiceVersion config.Key = "telemetry.service.version"
	configEnvironment    config.Key = "telemetry.service.environment"
)

// Propagator keys
const (
	configPropagators config.Key = "telemetry.propagators"
)

// Traces keys
const (
	configTracesExporter      config.Key = "telemetry.traces.exporter"
	configTracesSamplerType   config.Key = "telemetry.traces.sampler.type"
	configTracesSamplerParent config.Key = "telemetry.traces.sampler.parent"
	configTracesSamplerRatio  config.Key = "telemetry.traces.sampler.ratio"
)

// Exporter keys with placeholder for exporter name
// Usage: configExporterDriver.With("otlp") -> "telemetry.exporters.otlp.driver"
const (
	configExporterDriver         config.Key = "telemetry.exporters.%s.driver"
	configExporterEndpoint       config.Key = "telemetry.exporters.%s.endpoint"
	configExporterProtocol       config.Key = "telemetry.exporters.%s.protocol"
	configExporterInsecure       config.Key = "telemetry.exporters.%s.insecure"
	configExporterTimeout        config.Key = "telemetry.exporters.%s.timeout"
	configExporterTracesTimeout  config.Key = "telemetry.exporters.%s.traces_timeout"
	configExporterTracesHeaders  config.Key = "telemetry.exporters.%s.traces_headers"
	configExporterTracesProtocol config.Key = "telemetry.exporters.%s.traces_protocol"
)
