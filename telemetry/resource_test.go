package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"github.com/goravel/framework/errors"
)

func getAttrValue(res *resource.Resource, key attribute.Key) (string, bool) {
	for _, attr := range res.Attributes() {
		if attr.Key == key {
			return attr.Value.AsString(), true
		}
	}
	return "", false
}

func TestNewResource(t *testing.T) {
	tests := []struct {
		name           string
		cfg            Config
		expectedValues map[attribute.Key]string
		expectedErr    error
	}{
		{
			name:        "1. Service Name Missing (Required Error)",
			cfg:         Config{Service: ServiceConfig{Name: ""}},
			expectedErr: errors.TelemetryServiceNameRequired,
		},
		{
			name: "2. Basic Service Identification",
			cfg: Config{
				Service: ServiceConfig{Name: "my-service"},
			},
			expectedValues: map[attribute.Key]string{
				semconv.ServiceNameKey: "my-service",
			},
		},
		{
			name: "3. All Standard Service Attributes Present",
			cfg: Config{
				Service: ServiceConfig{
					Name:        "my-service",
					Version:     "1.0.0",
					Environment: "staging",
				},
			},
			expectedValues: map[attribute.Key]string{
				semconv.ServiceNameKey:               "my-service",
				semconv.ServiceVersionKey:            "1.0.0",
				semconv.DeploymentEnvironmentNameKey: "staging",
			},
		},
		{
			name: "4. With Custom User Resources",
			cfg: Config{
				Service: ServiceConfig{Name: "my-service"},
				Resource: map[string]string{
					"region": "us-east-1",
					"team":   "backend",
					"":       "ignore_empty_key",
				},
			},
			expectedValues: map[attribute.Key]string{
				semconv.ServiceNameKey: "my-service",
				"region":               "us-east-1",
				"team":                 "backend",
			},
		},
		{
			name: "5. Mixing Config and SDK Defaults",
			cfg: Config{
				Service: ServiceConfig{
					Name: "override-name",
				},
				// Note: OpenTelemetry SDK detects OS, Process, etc.
				// We only check if the configured value is present.
			},
			expectedValues: map[attribute.Key]string{
				semconv.ServiceNameKey: "override-name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			res, err := newResource(ctx, tt.cfg)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr, "Expected a specific error")
				assert.Nil(t, res, "Result should be nil on error")
				return
			}

			assert.NoError(t, err, "Should not return an error for valid config")
			assert.NotNil(t, res, "Resource should not be nil")

			for key, expectedVal := range tt.expectedValues {
				actualVal, found := getAttrValue(res, key)

				assert.True(t, found, "Expected attribute %s to be present", key)
				assert.Equal(t, expectedVal, actualVal, "Value mismatch for attribute %s", key)
			}

			// Since newResource also calls resource.WithOS(), resource.WithProcess(),
			// we can perform a smoke test for one of the automatically added attributes.
			_, foundOS := getAttrValue(res, semconv.OSDescriptionKey)
			assert.True(t, foundOS, "Resource should include OS information from automatic detectors")
		})
	}
}
