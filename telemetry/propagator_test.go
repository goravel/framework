package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/errors"
)

func TestNewCompositeTextMapPropagator(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedFields []string
		expectError    error
	}{
		{
			name:        "empty returns error",
			input:       "",
			expectError: errors.TelemetryPropagatorRequired,
		},
		{
			name:           "tracecontext",
			input:          "tracecontext",
			expectedFields: []string{"traceparent", "tracestate"},
		},
		{
			name:           "baggage",
			input:          "baggage",
			expectedFields: []string{"baggage"},
		},
		{
			name:           "b3 single header",
			input:          "b3",
			expectedFields: []string{"b3"},
		},
		{
			name:           "b3 multi header",
			input:          "b3multi",
			expectedFields: []string{"x-b3-traceid", "x-b3-spanid", "x-b3-sampled"},
		},
		{
			name:           "multiple propagators",
			input:          "tracecontext,baggage,b3",
			expectedFields: []string{"traceparent", "baggage", "b3"},
		},
		{
			name:           "handles whitespace",
			input:          "tracecontext, baggage",
			expectedFields: []string{"traceparent", "baggage"},
		},
		{
			name:        "unknown propagator returns error",
			input:       "invalid",
			expectError: errors.TelemetryUnsupportedPropagator.Args("invalid"),
		},
		{
			name:        "mixed valid and invalid returns error",
			input:       "tracecontext,invalid",
			expectError: errors.TelemetryUnsupportedPropagator.Args("invalid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			propagator, err := newCompositeTextMapPropagator(tt.input)

			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
				assert.Nil(t, propagator)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, propagator)

			fields := propagator.Fields()
			for _, expected := range tt.expectedFields {
				assert.Contains(t, fields, expected)
			}
		})
	}
}
