package main

import (
	"strings"
)

type Stubs struct{}

func (s Stubs) Config(pkg, module string) string {
	content := `package DummyPackage

import (
	"DummyModule/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("telemetry", map[string]any{
		// Service Identification
		//
		// These values identify your service in distributed traces.
		"service": map[string]any{
			"name":        config.Env("APP_NAME", "goravel"),
			"version":     config.Env("APP_VERSION", ""),
			"environment": config.Env("APP_ENV", ""),
		},

		// Propagators
		//
		// Propagators define how trace context is passed between services.
		// Supported: "tracecontext", "baggage", "b3", "b3multi"
		"propagators": config.Env("OTEL_PROPAGATORS", "tracecontext"),

		// Traces Configuration
		//
		// Configure distributed tracing for your application.
		"traces": map[string]any{
			// The exporter determines where traces are sent.
			// Supported: "otlp", "zipkin", "console"
			"exporter": config.Env("OTEL_TRACES_EXPORTER", "otlp_trace"),

			// Sampler Configuration
			//
			// Controls which traces are recorded.
			"sampler": map[string]any{
				"parent": config.Env("OTEL_TRACES_SAMPLER_PARENT", true),
				// Supported: "always_on", "always_off", "traceidratio"
				"type": config.Env("OTEL_TRACES_SAMPLER_TYPE", "always_on"),
				// Sampling ratio for "traceidratio" (0.0 to 1.0)
				"ratio": config.Env("OTEL_TRACES_SAMPLER_RATIO", 0.05),
			},
		},

		// Exporters Configuration
		//
		// Configure exporters for sending telemetry data.
		// Supported drivers: "otlp", "zipkin", "console"
		"exporters": map[string]any{
			"otlp_trace": map[string]any{
				"driver":          "otlp",
				"endpoint":        config.Env("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "http://localhost:4318"),
				"protocol":        config.Env("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http/protobuf"),
				"insecure":        config.Env("OTEL_EXPORTER_OTLP_TRACES_INSECURE", true),
				"timeout":         config.Env("OTEL_EXPORTER_OTLP_TRACES_TIMEOUT", 10000),
				"headers":         config.Env("OTEL_EXPORTER_OTLP_TRACES_HEADERS", ""),
			},
			"zipkin": map[string]any{
				"driver":   "zipkin",
				"endpoint": config.Env("OTEL_EXPORTER_ZIPKIN_ENDPOINT", "http://localhost:9411/api/v2/spans"),
			},
			"console": map[string]any{
				"driver": "console",
			},
		},
	})
}
`

	content = strings.ReplaceAll(content, "DummyPackage", pkg)
	content = strings.ReplaceAll(content, "DummyModule", module)

	return content
}

func (s Stubs) TelemetryFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/telemetry"
)

func Telemetry() telemetry.Telemetry {
	return App().MakeTelemetry()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
