package telemetry

type Config struct {
	Service     ServiceConfig
	Propagators string
	Metrics     MetricsConfig
	Traces      TracesConfig
	Exporters   map[string]ExporterEntry
}

type ServiceConfig struct {
	Name        string
	Version     string
	Environment string
	InstanceID  string
}

type TracesConfig struct {
	Exporter string
	Sampler  SamplerConfig
}

type MetricsConfig struct {
	Exporter string
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
	Protocol Protocol
	Headers  string
}

func (c Config) GetExporter(name string) (ExporterEntry, bool) {
	if exp, ok := c.Exporters[name]; ok {
		return exp, true
	}
	return ExporterEntry{Driver: ExporterDriver(name)}, false
}
