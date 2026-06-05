package telemetry

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/errors"
)

type MockOption string

var mockOTLPOptions = otlpOptions[MockOption]{
	withEndpoint:    func(s string) MockOption { return MockOption("endpoint=" + s) },
	withEndpointURL: func(s string) MockOption { return MockOption("endpoint_url=" + s) },
	withInsecure:    func() MockOption { return MockOption("insecure=true") },
	withTimeout:     func(d time.Duration) MockOption { return MockOption("timeout=" + d.String()) },
	withHeaders: func(h map[string]string) MockOption {
		if val, ok := h["Authorization"]; ok {
			return MockOption("header_auth=" + val)
		}
		return MockOption("headers_present")
	},
	withCompression: func() MockOption { return MockOption("compression=gzip") },
	withTLS:         func(*tls.Config) MockOption { return MockOption("tls=true") },
	withRetry: func(r RetryConfig) MockOption {
		return MockOption("retry=" + r.MaxElapsedTime.String())
	},
}

func TestBuildOTLPOptions(t *testing.T) {
	caFile, _, _ := writeTestCerts(t)

	tests := []struct {
		name        string
		cfg         ExporterEntry
		expected    []MockOption
		expectError error
	}{
		{
			name: "Empty Config (Defaults)",
			cfg:  ExporterEntry{},
			expected: []MockOption{
				"timeout=10s",
			},
		},
		{
			name: "Endpoint Without Scheme",
			cfg: ExporterEntry{
				Endpoint: "localhost:4318",
			},
			expected: []MockOption{
				"endpoint=localhost:4318",
				"timeout=10s",
			},
		},
		{
			name: "Endpoint With URL",
			cfg: ExporterEntry{
				Endpoint: "https://otel.com/otel",
			},
			expected: []MockOption{
				"endpoint_url=https://otel.com/otel",
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
			name: "With Compression And Retry",
			cfg: ExporterEntry{
				Endpoint:    "localhost:4318",
				Compression: "gzip",
				Retry:       RetryConfig{MaxElapsedTime: 10 * time.Second},
			},
			expected: []MockOption{
				"endpoint=localhost:4318",
				"timeout=10s",
				"compression=gzip",
				"retry=10s",
			},
		},
		{
			name: "With TLS",
			cfg: ExporterEntry{
				Endpoint: "localhost:4318",
				TLS:      TLSConfig{CA: caFile},
			},
			expected: []MockOption{
				"endpoint=localhost:4318",
				"timeout=10s",
				"tls=true",
			},
		},
		{
			name: "Unsupported Compression",
			cfg: ExporterEntry{
				Compression: "zstd",
			},
			expectError: errors.TelemetryUnsupportedCompression.Args("zstd"),
		},
		{
			name: "TLS Conflicts With Insecure",
			cfg: ExporterEntry{
				Insecure: true,
				TLS:      TLSConfig{CA: caFile},
			},
			expectError: errors.TelemetryTLSConflictsWithInsecure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := buildOTLPOptions(tt.cfg, mockOTLPOptions)

			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
				assert.Nil(t, opts)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, opts)
		})
	}
}

func TestNewTLSConfig(t *testing.T) {
	caFile, certFile, keyFile := writeTestCerts(t)

	tests := []struct {
		name        string
		entry       ExporterEntry
		expectError error
		expectNil   bool
		expectCerts int
		expectCA    bool
	}{
		{
			name:      "No TLS fields returns nil",
			entry:     ExporterEntry{},
			expectNil: true,
		},
		{
			name:     "CA only",
			entry:    ExporterEntry{TLS: TLSConfig{CA: caFile}},
			expectCA: true,
		},
		{
			name:        "CA with client keypair",
			entry:       ExporterEntry{TLS: TLSConfig{CA: caFile, Cert: certFile, Key: keyFile}},
			expectCA:    true,
			expectCerts: 1,
		},
		{
			name:        "Conflicts with insecure",
			entry:       ExporterEntry{Insecure: true, TLS: TLSConfig{CA: caFile}},
			expectError: errors.TelemetryTLSConflictsWithInsecure,
		},
		{
			name:        "Cert without key",
			entry:       ExporterEntry{TLS: TLSConfig{Cert: certFile}},
			expectError: errors.TelemetryTLSClientCertIncomplete,
		},
		{
			name:        "Key without cert",
			entry:       ExporterEntry{TLS: TLSConfig{Key: keyFile}},
			expectError: errors.TelemetryTLSClientCertIncomplete,
		},
		{
			name:        "Invalid CA file",
			entry:       ExporterEntry{TLS: TLSConfig{CA: filepath.Join(t.TempDir(), "missing.pem")}},
			expectError: nil, // os wrapped error; asserted via assert.Error below
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := newTLSConfig(tt.entry)

			if tt.name == "Invalid CA file" {
				assert.Error(t, err)
				assert.Nil(t, cfg)
				return
			}
			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
				assert.Nil(t, cfg)
				return
			}

			assert.NoError(t, err)
			if tt.expectNil {
				assert.Nil(t, cfg)
				return
			}
			assert.Equal(t, tt.expectCA, cfg.RootCAs != nil)
			assert.Len(t, cfg.Certificates, tt.expectCerts)
		})
	}
}

func writeTestCerts(t *testing.T) (caFile, certFile, keyFile string) {
	t.Helper()
	dir := t.TempDir()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "goravel-test"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	caFile = filepath.Join(dir, "ca.pem")
	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")
	require.NoError(t, os.WriteFile(caFile, certPEM, 0600))
	require.NoError(t, os.WriteFile(certFile, certPEM, 0600))
	require.NoError(t, os.WriteFile(keyFile, keyPEM, 0600))
	return caFile, certFile, keyFile
}
