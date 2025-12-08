package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResource(t *testing.T) {
	tests := []struct {
		name          string
		cfg           ServiceConfig
		expectService string
		expectVersion bool
		expectEnv     bool
	}{
		{
			name:          "default service name when empty",
			cfg:           ServiceConfig{},
			expectService: "goravel",
		},
		{
			name:          "custom service name",
			cfg:           ServiceConfig{Name: "my-service"},
			expectService: "my-service",
		},
		{
			name: "with version and environment",
			cfg: ServiceConfig{
				Name:        "my-service",
				Version:     "1.0.0",
				Environment: "production",
			},
			expectService: "my-service",
			expectVersion: true,
			expectEnv:     true,
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
