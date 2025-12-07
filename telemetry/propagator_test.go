package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCompositeTextMapPropagator(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedFields []string
		expectEmpty    bool
	}{
		{
			name:           "empty returns default (tracecontext + baggage)",
			input:          "",
			expectedFields: []string{"traceparent", "tracestate", "baggage"},
		},
		{
			name:        "none returns empty propagator",
			input:       "none",
			expectEmpty: true,
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
			name:           "unknown propagator falls back to default",
			input:          "invalid",
			expectedFields: []string{"traceparent", "tracestate", "baggage"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			propagator := newCompositeTextMapPropagator(tt.input)
			fields := propagator.Fields()

			if tt.expectEmpty {
				assert.Empty(t, fields)
				return
			}

			for _, expected := range tt.expectedFields {
				assert.Contains(t, fields, expected)
			}
		})
	}
}
