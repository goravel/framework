package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
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

func (s *TransportTestSuite) TestRoundTrip() {
	req := httptest.NewRequest("GET", "http://example.com", nil)

	tests := []struct {
		name  string
		setup func(*mockstelemetry.Telemetry, *mocksconfig.Config, *MockRoundTripper)
	}{
		{
			name: "Fallback: Base used when ConfigFacade is nil",
			setup: func(_ *mockstelemetry.Telemetry, _ *mocksconfig.Config, baseMock *MockRoundTripper) {
				telemetry.ConfigFacade = nil
				baseMock.On("RoundTrip", req).Return(&http.Response{}, nil).Once()
			},
		},
		{
			name: "Kill Switch: Base used when http_client is disabled",
			setup: func(_ *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config, baseMock *MockRoundTripper) {
				telemetry.ConfigFacade = mockConfig
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.http_client", true).Return(false).Once()

				baseMock.On("RoundTrip", req).Return(&http.Response{}, nil).Once()
			},
		},
		{
			name: "Fallback: Base used (with warning) when TelemetryFacade is nil",
			setup: func(_ *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config, baseMock *MockRoundTripper) {
				telemetry.ConfigFacade = mockConfig
				telemetry.TelemetryFacade = nil

				mockConfig.EXPECT().GetBool("telemetry.instrumentation.http_client", true).Return(true).Once()

				baseMock.On("RoundTrip", req).Return(&http.Response{}, nil).Once()
			},
		},
		{
			name: "Success: OTel Transport initialized and used",
			setup: func(mockTelemetry *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config, baseMock *MockRoundTripper) {
				telemetry.ConfigFacade = mockConfig
				telemetry.TelemetryFacade = mockTelemetry

				mockConfig.EXPECT().GetBool("telemetry.instrumentation.http_client", true).Return(true).Once()
				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()
				baseMock.On("RoundTrip", mock.Anything).Return(&http.Response{}, nil).Once()
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockTelemetry := mockstelemetry.NewTelemetry(s.T())
			mockConfig := mocksconfig.NewConfig(s.T())
			mockBase := &MockRoundTripper{}

			test.setup(mockTelemetry, mockConfig, mockBase)

			transport := NewTransport(mockBase)

			color.CaptureOutput(func(w io.Writer) {
				_, _ = transport.RoundTrip(req)
			})

			mockTelemetry.AssertExpectations(s.T())
			mockConfig.AssertExpectations(s.T())
			mockBase.AssertExpectations(s.T())
		})
	}
}

type MockRoundTripper struct {
	mock.Mock
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}
