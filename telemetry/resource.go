package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"github.com/goravel/framework/errors"
)

func newResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	serviceCfg := cfg.Service
	serviceName := serviceCfg.Name
	if serviceName == "" {
		return nil, errors.TelemetryServiceNameRequired
	}

	attrs := []resource.Option{
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	}

	if serviceCfg.Version != "" {
		attrs = append(attrs, resource.WithAttributes(semconv.ServiceVersion(serviceCfg.Version)))
	}

	if serviceCfg.Environment != "" {
		attrs = append(attrs, resource.WithAttributes(semconv.DeploymentEnvironmentName(serviceCfg.Environment)))
	}

	if serviceCfg.InstanceID != "" {
		attrs = append(attrs, resource.WithAttributes(semconv.ServiceInstanceID(serviceCfg.InstanceID)))
	}

	for k, v := range cfg.Resource {
		if k != "" {
			attrs = append(attrs, resource.WithAttributes(attribute.String(k, v)))
		}
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
