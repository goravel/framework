package telemetry

type Config struct {
	Service     ServiceConfig
	Propagators string
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

type SamplerConfig struct {
	Type   string
	Ratio  float64
	Parent bool
}

type ExporterEntry struct {
	Driver   ExporterDriver
	Endpoint string
	Insecure bool
	Timeout  int

	// OTLP-specific
	Protocol       Protocol
	TracesTimeout  int      `mapstructure:"traces_timeout"`
	TracesHeaders  string   `mapstructure:"traces_headers"`
	TracesProtocol Protocol `mapstructure:"traces_protocol"`
}

func (c Config) GetExporter(name string) (ExporterEntry, bool) {
	if exp, ok := c.Exporters[name]; ok {
		return exp, true
	}
	return ExporterEntry{Driver: ExporterDriver(name)}, false
}
