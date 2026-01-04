package http

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	contractsconfig "github.com/goravel/framework/contracts/config"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/telemetry"
)

type TransportTestSuite struct {
	suite.Suite
	originalTelemetryFacade contractstelemetry.Telemetry
	originalConfigFacade    contractsconfig.Config
}

func (s *TransportTestSuite) SetupTest() {
	s.originalTelemetryFacade = telemetry.TelemetryFacade
	s.originalConfigFacade = telemetry.ConfigFacade
}

func (s *TransportTestSuite) TearDownTest() {
	telemetry.TelemetryFacade = s.originalTelemetryFacade
	telemetry.ConfigFacade = s.originalConfigFacade
}

func TestTransportTestSuite(t *testing.T) {
	suite.Run(t, new(TransportTestSuite))
}

func (s *TransportTestSuite) TestNewTransport() {
	tests := []struct {
		name   string
		setup  func(*mockstelemetry.Telemetry, *mocksconfig.Config)
		base   http.RoundTripper
		assert func(res http.RoundTripper)
	}{
		{
			name: "fallback: returns base when ConfigFacade is nil",
			setup: func(_ *mockstelemetry.Telemetry, _ *mocksconfig.Config) {
				telemetry.ConfigFacade = nil
			},
			base: http.DefaultTransport,
			assert: func(res http.RoundTripper) {
				s.Equal(http.DefaultTransport, res)
			},
		},
		{
			name: "kill switch: returns base when http_client is disabled in config",
			setup: func(_ *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				telemetry.ConfigFacade = mockConfig
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.http_client", true).Return(false).Once()
			},
			base: http.DefaultTransport,
			assert: func(res http.RoundTripper) {
				s.Equal(http.DefaultTransport, res)
			},
		},
		{
			name: "fallback: returns base when TelemetryFacade is nil (even if config enabled)",
			setup: func(_ *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				telemetry.ConfigFacade = mockConfig
				telemetry.TelemetryFacade = nil

				mockConfig.EXPECT().GetBool("telemetry.instrumentation.http_client", true).Return(true).Once()
			},
			base: http.DefaultTransport,
			assert: func(res http.RoundTripper) {
				s.Equal(http.DefaultTransport, res)
			},
		},
		{
			name: "success: returns wrapped transport when enabled and facades exist",
			setup: func(mockTelemetry *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				telemetry.ConfigFacade = mockConfig
				telemetry.TelemetryFacade = mockTelemetry

				mockConfig.EXPECT().GetBool("telemetry.instrumentation.http_client", true).Return(true).Once()

				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()
			},
			base: http.DefaultTransport,
			assert: func(res http.RoundTripper) {
				s.NotNil(res)
				s.NotEqual(http.DefaultTransport, res)
			},
		},
		{
			name: "success: handles nil base automatically (wraps DefaultTransport)",
			setup: func(mockTelemetry *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				telemetry.ConfigFacade = mockConfig
				telemetry.TelemetryFacade = mockTelemetry

				mockConfig.EXPECT().GetBool("telemetry.instrumentation.http_client", true).Return(true).Once()
				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()
			},
			base: nil,
			assert: func(res http.RoundTripper) {
				// otelhttp.NewTransport(nil) will wrap DefaultTransport.
				// So result should be NotNil.
				s.NotNil(res)
				s.NotEqual(http.DefaultTransport, res)
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockTelemetry := mockstelemetry.NewTelemetry(s.T())
			mockConfig := mocksconfig.NewConfig(s.T())

			test.setup(mockTelemetry, mockConfig)

			var res http.RoundTripper
			color.CaptureOutput(func(w io.Writer) {
				res = NewTransport(test.base)
			})

			test.assert(res)
		})
	}
}
