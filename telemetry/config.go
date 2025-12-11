package telemetry

import "time"

type Config struct {
	Resource    map[string]string
	Service     ServiceConfig
	Propagators string
	Logs        LogsConfig
	Metrics     MetricsConfig
	Traces      TracesConfig
	Exporters   map[string]ExporterEntry
}

type ServiceConfig struct {
	Name        string
	Version     string
	Environment string
	InstanceID  string `mapstructure:"instance_id"`
}

type TracesConfig struct {
	Exporter string
	Sampler  SamplerConfig
}

type MetricsConfig struct {
	Exporter string
	Reader   MetricsReaderConfig
}

type LogsConfig struct {
	Exporter  string
	Processor LogsProcessorConfig
}

type LogsProcessorConfig struct {
	Interval time.Duration
	Timeout  time.Duration
}

type MetricsReaderConfig struct {
	Interval time.Duration
	Timeout  time.Duration
}

type SamplerConfig struct {
	Type   string
	Ratio  float64
	Parent bool
}

type ExporterEntry struct {
	Driver   ExporterDriver
	Endpoint string
	Insecure bool
	Timeout  time.Duration

	// OTLP-specific
	Protocol Protocol
	Headers  map[string]string

	// Metric Specific
	MetricTemporality MetricTemporality `mapstructure:"metric_temporality"`

	// For custom Exporter
	Via any
}

func (c Config) GetExporter(name string) (ExporterEntry, bool) {
	if exp, ok := c.Exporters[name]; ok {
		return exp, true
	}
	return ExporterEntry{Driver: ExporterDriver(name)}, false
}
