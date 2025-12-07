package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTraceSampler(t *testing.T) {
	tests := []struct {
		name               string
		samplerType        string
		parentBased        bool
		ratio              float64
		expectParentBased  bool
		expectAlwaysOn     bool
		expectAlwaysOff    bool
		expectTraceIDRatio bool
	}{
		{
			name:              "empty type returns default (parent-based always on)",
			samplerType:       "",
			parentBased:       true,
			expectParentBased: true,
			expectAlwaysOn:    true,
		},
		{
			name:              "always_on with parent based",
			samplerType:       "always_on",
			parentBased:       true,
			expectParentBased: true,
			expectAlwaysOn:    true,
		},
		{
			name:           "always_on without parent based",
			samplerType:    "always_on",
			parentBased:    false,
			expectAlwaysOn: true,
		},
		{
			name:              "always_off with parent based",
			samplerType:       "always_off",
			parentBased:       true,
			expectParentBased: true,
			expectAlwaysOff:   true,
		},
		{
			name:            "always_off without parent based",
			samplerType:     "always_off",
			parentBased:     false,
			expectAlwaysOff: true,
		},
		{
			name:               "traceidratio with 50% ratio",
			samplerType:        "traceidratio",
			parentBased:        true,
			ratio:              0.5,
			expectParentBased:  true,
			expectTraceIDRatio: true,
		},
		{
			name:           "unknown type defaults to always_on",
			samplerType:    "invalid",
			parentBased:    false,
			expectAlwaysOn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sampler := newTraceSampler(tt.samplerType, tt.parentBased, tt.ratio)

			assert.NotNil(t, sampler)
			desc := sampler.Description()

			if tt.expectParentBased {
				assert.Contains(t, desc, "ParentBased")
			}
			if tt.expectAlwaysOn {
				assert.Contains(t, desc, "AlwaysOnSampler")
			}
			if tt.expectAlwaysOff {
				assert.Contains(t, desc, "AlwaysOffSampler")
			}
			if tt.expectTraceIDRatio {
				assert.Contains(t, desc, "TraceIDRatioBased")
			}
		})
	}
}
