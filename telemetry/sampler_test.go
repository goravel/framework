package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTraceSampler(t *testing.T) {
	tests := []struct {
		name               string
		cfg                SamplerConfig
		expectParentBased  bool
		expectAlwaysOn     bool
		expectAlwaysOff    bool
		expectTraceIDRatio bool
	}{
		{
			name:           "empty type defaults to always_on",
			cfg:            SamplerConfig{},
			expectAlwaysOn: true,
		},
		{
			name:              "always_on with parent based",
			cfg:               SamplerConfig{Type: "always_on", Parent: true},
			expectParentBased: true,
			expectAlwaysOn:    true,
		},
		{
			name:           "always_on without parent based",
			cfg:            SamplerConfig{Type: "always_on", Parent: false},
			expectAlwaysOn: true,
		},
		{
			name:              "always_off with parent based",
			cfg:               SamplerConfig{Type: "always_off", Parent: true},
			expectParentBased: true,
			expectAlwaysOff:   true,
		},
		{
			name:            "always_off without parent based",
			cfg:             SamplerConfig{Type: "always_off", Parent: false},
			expectAlwaysOff: true,
		},
		{
			name:               "traceidratio with 50% ratio",
			cfg:                SamplerConfig{Type: "traceidratio", Parent: true, Ratio: 0.5},
			expectParentBased:  true,
			expectTraceIDRatio: true,
		},
		{
			name:           "unknown type defaults to always_on",
			cfg:            SamplerConfig{Type: "invalid", Parent: false},
			expectAlwaysOn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sampler := newTraceSampler(tt.cfg)

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
