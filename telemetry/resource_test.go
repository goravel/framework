package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

func TestNewResource(t *testing.T) {
	tests := []struct {
		name          string
		cfg           resourceConfig
		expectService string
		expectVersion bool
		expectEnv     bool
	}{
		{
			name:          "default service name when empty",
			cfg:           resourceConfig{},
			expectService: "goravel",
		},
		{
			name: "custom service name",
			cfg: resourceConfig{
				serviceName: "my-service",
			},
			expectService: "my-service",
		},
		{
			name: "with version and environment",
			cfg: resourceConfig{
				serviceName:    "my-service",
				serviceVersion: "1.0.0",
				environment:    "production",
			},
			expectService: "my-service",
			expectVersion: true,
			expectEnv:     true,
		},
		{
			name: "with custom attributes",
			cfg: resourceConfig{
				serviceName: "my-service",
				attributes: []attribute.KeyValue{
					attribute.String("custom.key", "custom.value"),
				},
			},
			expectService: "my-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			res, err := newResource(ctx, tt.cfg)

			assert.NoError(t, err)
			assert.NotNil(t, res)

			attrs := res.Attributes()
			var foundService, foundVersion, foundEnv bool
			for _, attr := range attrs {
				if attr.Key == "service.name" && attr.Value.AsString() == tt.expectService {
					foundService = true
				}
				if attr.Key == "service.version" {
					foundVersion = true
				}
				if attr.Key == "deployment.environment.name" {
					foundEnv = true
				}
			}

			assert.True(t, foundService, "service.name should be present")
			assert.Equal(t, tt.expectVersion, foundVersion, "service.version presence mismatch")
			assert.Equal(t, tt.expectEnv, foundEnv, "deployment.environment presence mismatch")
		})
	}
}
