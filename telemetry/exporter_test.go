package telemetry

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConsoleTraceExporter(t *testing.T) {
	tests := []struct {
		name string
		cfg  consoleExporterConfig
	}{
		{
			name: "default writer",
			cfg:  consoleExporterConfig{},
		},
		{
			name: "custom writer with pretty print",
			cfg: consoleExporterConfig{
				writer:      &bytes.Buffer{},
				prettyPrint: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, err := newConsoleTraceExporter(tt.cfg)

			assert.NoError(t, err)
			assert.NotNil(t, exp)
		})
	}
}

func TestNewZipkinTraceExporter(t *testing.T) {
	tests := []struct {
		name string
		cfg  zipkinExporterConfig
	}{
		{
			name: "default endpoint",
			cfg:  zipkinExporterConfig{},
		},
		{
			name: "custom endpoint",
			cfg: zipkinExporterConfig{
				endpoint: "http://zipkin:9411/api/v2/spans",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, err := newZipkinTraceExporter(tt.cfg)

			assert.NoError(t, err)
			assert.NotNil(t, exp)
		})
	}
}

func TestNewOTLPTraceExporter(t *testing.T) {
	tests := []struct {
		name string
		cfg  otlpExporterConfig
	}{
		{
			name: "default protocol (http/protobuf)",
			cfg:  otlpExporterConfig{endpoint: "localhost:4318", insecure: true},
		},
		{
			name: "grpc protocol",
			cfg: otlpExporterConfig{
				endpoint: "localhost:4317",
				protocol: protocolGRPC,
				insecure: true,
			},
		},
		{
			name: "with headers and timeout",
			cfg: otlpExporterConfig{
				endpoint: "localhost:4318",
				protocol: protocolHTTPProtobuf,
				insecure: true,
				timeout:  5000,
				headers:  map[string]string{"Authorization": "Bearer token"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			exp, err := newOTLPTraceExporter(ctx, tt.cfg)

			assert.NoError(t, err)
			assert.NotNil(t, exp)
		})
	}
}

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:     "single header",
			input:    "Authorization=Bearer token",
			expected: map[string]string{"Authorization": "Bearer token"},
		},
		{
			name:     "multiple headers",
			input:    "X-Api-Key=abc123,X-Tenant=tenant1",
			expected: map[string]string{"X-Api-Key": "abc123", "X-Tenant": "tenant1"},
		},
		{
			name:     "handles whitespace",
			input:    " X-Api-Key = abc123 , X-Tenant = tenant1 ",
			expected: map[string]string{"X-Api-Key": "abc123", "X-Tenant": "tenant1"},
		},
		{
			name:     "skips invalid entries",
			input:    "valid=value,invalid,another=one",
			expected: map[string]string{"valid": "value", "another": "one"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHeaders(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
