package main

import (
	"strings"
)

type Stubs struct{}

func (s Stubs) Config(pkg, module string) string {
	content := `package DummyPackage

import (
    "time"

	"DummyModule/app/facades"
)

func init() {
    config := facades.Config()
    config.Add("telemetry", map[string]any{
       
       // Service Identification
       //
       // Identifies your service in traces and metrics.
       "service": map[string]any{
          "name":        config.Env("APP_NAME", "goravel"),
          "version":     config.Env("APP_VERSION", ""),
          "environment": config.Env("APP_ENV", ""),
       },
       
       // Resource Attributes
       //
       // Additional user-defined attributes to attach to the Resource object.
       "resource": map[string]string{},

       // Propagators
       //
       // Defines how trace context is passed between services.
       // Supported: "tracecontext", "baggage", "b3", "b3multi"
       "propagators": config.Env("OTEL_PROPAGATORS", "tracecontext"),

       // Traces Configuration
       //
       // Configures distributed tracing for your application.
       "traces": map[string]any{
          // Exporter
          //
          // The exporter determines where traces are sent.
          "exporter": config.Env("OTEL_TRACES_EXPORTER", "otlptrace"),

          // Sampler Configuration
          //
          // Controls which traces are recorded.
          "sampler": map[string]any{
             "parent": config.Env("OTEL_TRACES_SAMPLER_PARENT", true),
             // Supported: "always_on", "always_off", "traceidratio"
             "type":   config.Env("OTEL_TRACES_SAMPLER_TYPE", "always_on"),
             // Sampling ratio for "traceidratio" (0.0 to 1.0)
             "ratio":  config.Env("OTEL_TRACES_SAMPLER_RATIO", 0.05),
          },
       },

       // Metrics Configuration
       //
       // Configures time-series metrics collection.
       "metrics": map[string]any{
          // Exporter
          //
          // The exporter determines where metrics are sent.
          "exporter": config.Env("OTEL_METRICS_EXPORTER", "otlpmetric"),

          // Reader Configuration
          //
          // Applies to push-based exporters (PeriodicReader timing).
          "reader": map[string]any{
             "interval": config.GetDuration("OTEL_METRIC_EXPORT_INTERVAL", 60*time.Second),
             "timeout":  config.GetDuration("OTEL_METRIC_EXPORT_TIMEOUT", 30*time.Second),
          },
       },

       // Exporters Configuration
       //
       // Configures transport and protocol details for telemetry destinations.
       // Supported drivers: "otlp", "zipkin", "console"
       "exporters": map[string]any{
          
          // OTLP Trace Exporter
          "otlptrace": map[string]any{
             "driver":          "otlp",
             "endpoint":        config.Env("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "http://localhost:4318"),
             "protocol":        config.Env("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "http/protobuf"),
             "insecure":        config.Env("OTEL_EXPORTER_OTLP_TRACES_INSECURE", true),
             "timeout":         config.GetDuration("OTEL_EXPORTER_OTLP_TRACES_TIMEOUT", 10*time.Second),
          },
          
          // OTLP Metric Exporter
          "otlpmetric": map[string]any{
             "driver":          "otlp",
             "endpoint":        config.Env("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "http://localhost:4318"),
             "protocol":        config.Env("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", "http/protobuf"),
             "insecure":        config.Env("OTEL_EXPORTER_OTLP_METRICS_INSECURE", true),
             "timeout":         config.GetDuration("OTEL_EXPORTER_OTLP_METRICS_TIMEOUT", 10*time.Second),
             "metric_temporality": config.Env("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY", "cumulative"), 
          },
          
          // Zipkin Exporter
          "zipkin": map[string]any{
             "driver":   "zipkin",
             "endpoint": config.Env("OTEL_EXPORTER_ZIPKIN_ENDPOINT", "http://localhost:9411/api/v2/spans"),
          },
          
          // Console Exporter
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
