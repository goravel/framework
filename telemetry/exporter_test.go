package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConsoleTraceExporter(t *testing.T) {
	exp, err := newConsoleTraceExporter()

	assert.NoError(t, err)
	assert.NotNil(t, exp)
}

func TestNewZipkinTraceExporter(t *testing.T) {
	tests := []struct {
		name string
		cfg  ExporterEntry
	}{
		{
			name: "default endpoint",
			cfg:  ExporterEntry{},
		},
		{
			name: "custom endpoint",
			cfg:  ExporterEntry{Endpoint: "http://zipkin:9411/api/v2/spans"},
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
		cfg  ExporterEntry
	}{
		{
			name: "default protocol (http/protobuf)",
			cfg:  ExporterEntry{Endpoint: "localhost:4318", Insecure: true},
		},
		{
			name: "grpc protocol",
			cfg: ExporterEntry{
				Endpoint: "localhost:4317",
				Protocol: ProtocolGRPC,
				Insecure: true,
			},
		},
		{
			name: "with headers and timeout",
			cfg: ExporterEntry{
				Endpoint: "localhost:4318",
				Protocol: ProtocolHTTPProtobuf,
				Insecure: true,
				Timeout:  5000,
				Headers:  "Authorization=Bearer token",
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
