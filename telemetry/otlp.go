package telemetry

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"strings"
	"time"

	"github.com/goravel/framework/errors"
)

type Protocol string

const (
	ProtocolGRPC         Protocol = "grpc"
	ProtocolHTTPProtobuf Protocol = "http/protobuf"
)

const defaultOTLPTimeout = 10 * time.Second

const CompressionGzip = "gzip"

type otlpOptions[T any] struct {
	withEndpoint    func(string) T
	withEndpointURL func(string) T
	withInsecure    func() T
	withTimeout     func(time.Duration) T
	withHeaders     func(map[string]string) T
	withCompression func() T
	withTLS         func(*tls.Config) T
	withRetry       func(RetryConfig) T
}

func buildOTLPOptions[T any](cfg ExporterEntry, builders otlpOptions[T]) ([]T, error) {
	var opts []T

	if cfg.Endpoint != "" {
		if strings.Contains(cfg.Endpoint, "://") {
			opts = append(opts, builders.withEndpointURL(cfg.Endpoint))
		} else {
			opts = append(opts, builders.withEndpoint(cfg.Endpoint))
		}
	}

	if cfg.Insecure {
		opts = append(opts, builders.withInsecure())
	}

	timeout := defaultOTLPTimeout
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}
	opts = append(opts, builders.withTimeout(timeout))

	if len(cfg.Headers) > 0 {
		opts = append(opts, builders.withHeaders(cfg.Headers))
	}

	switch cfg.Compression {
	case CompressionGzip:
		opts = append(opts, builders.withCompression())
	case "":
	default:
		return nil, errors.TelemetryUnsupportedCompression.Args(cfg.Compression)
	}

	tlsConfig, err := newTLSConfig(cfg)
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		opts = append(opts, builders.withTLS(tlsConfig))
	}

	if cfg.Retry != (RetryConfig{}) {
		opts = append(opts, builders.withRetry(cfg.Retry.withDefaults()))
	}

	return opts, nil
}

func newTLSConfig(cfg ExporterEntry) (*tls.Config, error) {
	tlsCfg := cfg.TLS
	if tlsCfg.CA == "" && tlsCfg.Cert == "" && tlsCfg.Key == "" {
		return nil, nil
	}

	if cfg.Insecure {
		return nil, errors.TelemetryTLSConflictsWithInsecure
	}

	if (tlsCfg.Cert == "") != (tlsCfg.Key == "") {
		return nil, errors.TelemetryTLSClientCertIncomplete
	}

	config := &tls.Config{}

	if tlsCfg.CA != "" {
		pemBytes, err := os.ReadFile(tlsCfg.CA)
		if err != nil {
			return nil, err
		}

		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pemBytes) {
			return nil, errors.TelemetryTLSInvalidCA
		}
		config.RootCAs = pool
	}

	if tlsCfg.Cert != "" {
		cert, err := tls.LoadX509KeyPair(tlsCfg.Cert, tlsCfg.Key)
		if err != nil {
			return nil, err
		}
		config.Certificates = []tls.Certificate{cert}
	}

	return config, nil
}
