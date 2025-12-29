package http

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/telemetry"
)

type TransportTestSuite struct {
	suite.Suite
	originalFacade contractstelemetry.Telemetry
}

func (s *TransportTestSuite) SetupTest() {
	s.originalFacade = telemetry.TelemetryFacade
}

func (s *TransportTestSuite) TearDownTest() {
	telemetry.TelemetryFacade = s.originalFacade
}

func TestTransportTestSuite(t *testing.T) {
	suite.Run(t, new(TransportTestSuite))
}

func (s *TransportTestSuite) TestNewTransport() {
	tests := []struct {
		name   string
		setup  func(*mockstelemetry.Telemetry)
		base   http.RoundTripper
		assert func(res http.RoundTripper)
	}{
		{
			name: "fallback: returns base when facade is nil",
			setup: func(_ *mockstelemetry.Telemetry) {
				telemetry.TelemetryFacade = nil
			},
			base: http.DefaultTransport,
			assert: func(res http.RoundTripper) {
				// We expect it to return exactly what we passed
				s.Equal(http.DefaultTransport, res)
			},
		},
		{
			name: "fallback: returns DefaultTransport when facade is nil AND base is nil",
			setup: func(_ *mockstelemetry.Telemetry) {
				telemetry.TelemetryFacade = nil
			},
			base: nil,
			assert: func(res http.RoundTripper) {
				s.NotNil(res)
				s.Equal(http.DefaultTransport, res)
			},
		},
		{
			name: "success: returns wrapped transport when facade is set",
			setup: func(mockTelemetry *mockstelemetry.Telemetry) {
				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()

				telemetry.TelemetryFacade = mockTelemetry
			},
			base: http.DefaultTransport,
			assert: func(res http.RoundTripper) {
				s.NotNil(res)
				s.NotEqual(http.DefaultTransport, res)
			},
		},
		{
			name: "success: handles nil base automatically (otelhttp behavior)",
			setup: func(mockTelemetry *mockstelemetry.Telemetry) {
				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()

				telemetry.TelemetryFacade = mockTelemetry
			},
			base: nil,
			assert: func(res http.RoundTripper) {
				// otelhttp.NewTransport(nil) will wrap DefaultTransport.
				// So result should be NotNil.
				s.NotNil(res)
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockTelemetry := mockstelemetry.NewTelemetry(s.T())

			test.setup(mockTelemetry)

			var res http.RoundTripper
			color.CaptureOutput(func(w io.Writer) {
				res = NewTransport(test.base)
			})

			test.assert(res)
		})
	}
}
