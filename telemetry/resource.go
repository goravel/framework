package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type resourceConfig struct {
	serviceName    string
	serviceVersion string
	environment    string
	attributes     []attribute.KeyValue
}

func newResource(ctx context.Context, cfg resourceConfig) (*resource.Resource, error) {
	serviceName := cfg.serviceName
	if serviceName == "" {
		serviceName = "goravel"
	}

	attrs := []attribute.KeyValue{
		semconv.ServiceName(serviceName),
	}

	if cfg.serviceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersion(cfg.serviceVersion))
	}

	if cfg.environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironmentName(cfg.environment))
	}

	attrs = append(attrs, cfg.attributes...)

	detected, err := resource.New(ctx,
		resource.WithAttributes(attrs...),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, err
	}

	return resource.Merge(resource.Default(), detected)
}
