package telemetry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockOption string

var withEndpoint = func(s string) MockOption {
	return MockOption("endpoint=" + s)
}

var withInsecure = func() MockOption {
	return MockOption("insecure=true")
}

var withTimeout = func(d time.Duration) MockOption {
	return MockOption("timeout=" + d.String())
}

var withHeaders = func(h map[string]string) MockOption {
	if val, ok := h["Authorization"]; ok {
		return MockOption("header_auth=" + val)
	}
	return MockOption("headers_present")
}

func TestBuildOTLPOptions(t *testing.T) {
	tests := []struct {
		name     string
		cfg      ExporterEntry
		expected []MockOption
	}{
		{
			name: "Empty Config (Defaults)",
			cfg:  ExporterEntry{},
			expected: []MockOption{
				"timeout=10s",
			},
		},
		{
			name: "Endpoint Stripping (HTTP)",
			cfg: ExporterEntry{
				Endpoint: "http://localhost:4318",
			},
			expected: []MockOption{
				"endpoint=localhost:4318",
				"timeout=10s",
			},
		},
		{
			name: "Endpoint Stripping (HTTPS)",
			cfg: ExporterEntry{
				Endpoint: "https://otel.com",
			},
			expected: []MockOption{
				"endpoint=otel.com",
				"timeout=10s",
			},
		},
		{
			name: "Insecure Enabled",
			cfg: ExporterEntry{
				Endpoint: "localhost:4318",
				Insecure: true,
			},
			expected: []MockOption{
				"endpoint=localhost:4318",
				"insecure=true",
				"timeout=10s",
			},
		},
		{
			name: "Custom Timeout",
			cfg: ExporterEntry{
				Timeout: 5 * time.Second,
			},
			expected: []MockOption{
				"timeout=5s",
			},
		},
		{
			name: "With Headers",
			cfg: ExporterEntry{
				Headers: map[string]string{
					"Authorization": "Bearer token",
				},
			},
			expected: []MockOption{
				"timeout=10s",
				"header_auth=Bearer token",
			},
		},
		{
			name: "Full Configuration",
			cfg: ExporterEntry{
				Endpoint: "https://api.honeycomb.io",
				Insecure: false,
				Timeout:  500 * time.Millisecond,
				Headers: map[string]string{
					"Authorization": "key",
				},
			},
			expected: []MockOption{
				"endpoint=api.honeycomb.io",
				"timeout=500ms",
				"header_auth=key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := buildOTLPOptions(
				tt.cfg,
				withEndpoint,
				withInsecure,
				withTimeout,
				withHeaders,
			)

			assert.Equal(t, tt.expected, opts)
		})
	}
}
