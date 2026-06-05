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

func buildOTLPOptions[T any](
	cfg ExporterEntry,
	withEndpoint func(string) T,
	withInsecure func() T,
	withTimeout func(time.Duration) T,
	withHeaders func(map[string]string) T,
) []T {
	var opts []T

	if cfg.Endpoint != "" {
		endpoint := strings.TrimPrefix(cfg.Endpoint, "http://")
		endpoint = strings.TrimPrefix(endpoint, "https://")
		opts = append(opts, withEndpoint(endpoint))
	}

	if cfg.Insecure {
		opts = append(opts, withInsecure())
	}

	timeout := defaultOTLPTimeout
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}
	opts = append(opts, withTimeout(timeout))

	if headers := cfg.Headers; len(headers) > 0 {
		opts = append(opts, withHeaders(headers))
	}

	return opts
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
