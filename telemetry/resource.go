package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func newResource(ctx context.Context, cfg ServiceConfig) (*resource.Resource, error) {
	serviceName := cfg.Name
	if serviceName == "" {
		serviceName = "goravel"
	}

	attrs := []resource.Option{
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	}

	if cfg.Version != "" {
		attrs = append(attrs, resource.WithAttributes(semconv.ServiceVersion(cfg.Version)))
	}

	if cfg.Environment != "" {
		attrs = append(attrs, resource.WithAttributes(semconv.DeploymentEnvironmentName(cfg.Environment)))
	}

	if cfg.InstanceID != "" {
		attrs = append(attrs, resource.WithAttributes(semconv.ServiceInstanceID(cfg.InstanceID)))
	}

	attrs = append(attrs,
		resource.WithFromEnv(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithContainer(),
		resource.WithHost(),
	)

	detected, err := resource.New(ctx, attrs...)
	if err != nil {
		return nil, err
	}

	return resource.Merge(resource.Default(), detected)
}
