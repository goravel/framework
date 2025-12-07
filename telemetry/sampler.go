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

func newTraceSampler(samplerType string, parentBased bool, ratio float64) sdktrace.Sampler {
	if samplerType == "" {
		return defaultTraceSampler
	}

	var sampler sdktrace.Sampler
	switch samplerType {
	case samplerAlwaysOff:
		sampler = sdktrace.NeverSample()
	case samplerTraceIDRatio:
		sampler = sdktrace.TraceIDRatioBased(ratio)
	default:
		sampler = sdktrace.AlwaysSample()
	}

	if parentBased {
		return sdktrace.ParentBased(sampler)
	}

	return sampler
}
