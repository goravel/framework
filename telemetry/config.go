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
	Protocol    Protocol
	Headers     map[string]string
	Compression string
	TLS         TLSConfig
	Retry       RetryConfig

	// Metric Specific
	MetricTemporality MetricTemporality `json:"metric_temporality"`

	// Console Driver Specific
	PrettyPrint bool `json:"pretty_print"`

	// For custom Exporter
	Via any
}

type TLSConfig struct {
	CA   string
	Cert string
	Key  string
}

type RetryConfig struct {
	Enabled         *bool
	InitialInterval time.Duration
	MaxInterval     time.Duration
	MaxElapsedTime  time.Duration
}

func (r RetryConfig) IsEnabled() bool {
	return r.Enabled == nil || *r.Enabled
}

func (r RetryConfig) withDefaults() RetryConfig {
	if r.InitialInterval == 0 {
		r.InitialInterval = 5 * time.Second
	}
	if r.MaxInterval == 0 {
		r.MaxInterval = 30 * time.Second
	}
	if r.MaxElapsedTime == 0 {
		r.MaxElapsedTime = time.Minute
	}
	return r
}

func (c Config) GetExporter(name string) (ExporterEntry, bool) {
	entry, ok := c.Exporters[name]
	return entry, ok
}
