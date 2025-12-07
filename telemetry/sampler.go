package telemetry

import sdktrace "go.opentelemetry.io/otel/sdk/trace"

const (
	samplerAlwaysOn     = "always_on"
	samplerAlwaysOff    = "always_off"
	samplerTraceIDRatio = "traceidratio"

	defaultRatio = 0.05
)

var (
	defaultTraceSampler = sdktrace.ParentBased(sdktrace.AlwaysSample())
)

type samplerConfig struct {
	samplerType string
	parentBased bool
	ratio       float64
}

func newTraceSampler(cfg samplerConfig) sdktrace.Sampler {
	if cfg.samplerType == "" {
		return defaultTraceSampler
	}

	var sampler sdktrace.Sampler
	switch cfg.samplerType {
	case samplerAlwaysOff:
		sampler = sdktrace.NeverSample()
	case samplerTraceIDRatio:
		sampler = sdktrace.TraceIDRatioBased(cfg.ratio)
	default:
		sampler = sdktrace.AlwaysSample()
	}

	if cfg.parentBased {
		return sdktrace.ParentBased(sampler)
	}

	return sampler
}
