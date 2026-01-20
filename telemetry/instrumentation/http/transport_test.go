package http

import (
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
)

type TransportTestSuite struct {
	suite.Suite
}

func TestTransportTestSuite(t *testing.T) {
	suite.Run(t, new(TransportTestSuite))
}

func (s *TransportTestSuite) TestRoundTrip() {
	req := httptest.NewRequest("GET", "http://example.com", nil)

	tests := []struct {
		name  string
		setup func(t *testing.T) (contractstelemetry.Telemetry, contractsconfig.Config, *MockRoundTripper)
	}{
		{
			name: "Fallback: Base used when Config is nil",
			setup: func(t *testing.T) (contractstelemetry.Telemetry, contractsconfig.Config, *MockRoundTripper) {
				baseMock := &MockRoundTripper{}
				baseMock.On("RoundTrip", req).Return(&http.Response{}, nil).Once()
				return mockstelemetry.NewTelemetry(t), nil, baseMock
			},
		},
		{
			name: "Fallback: Base used when Telemetry is nil",
			setup: func(t *testing.T) (contractstelemetry.Telemetry, contractsconfig.Config, *MockRoundTripper) {
				baseMock := &MockRoundTripper{}
				baseMock.On("RoundTrip", req).Return(&http.Response{}, nil).Once()

				return nil, mocksconfig.NewConfig(t), baseMock
			},
		},
		{
			name: "Kill Switch: Base used when http_client is disabled",
			setup: func(t *testing.T) (contractstelemetry.Telemetry, contractsconfig.Config, *MockRoundTripper) {
				mockConfig := mocksconfig.NewConfig(t)
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.http_client", true).Return(false).Once()

				baseMock := &MockRoundTripper{}
				baseMock.On("RoundTrip", req).Return(&http.Response{}, nil).Once()

				return mockstelemetry.NewTelemetry(t), mockConfig, baseMock
			},
		},
		{
			name: "Success: OTel Transport initialized and used",
			setup: func(t *testing.T) (contractstelemetry.Telemetry, contractsconfig.Config, *MockRoundTripper) {
				mockConfig := mocksconfig.NewConfig(t)
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.http_client", true).Return(true).Once()

				mockTelemetry := mockstelemetry.NewTelemetry(t)
				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()

				baseMock := &MockRoundTripper{}
				baseMock.On("RoundTrip", mock.Anything).Return(&http.Response{}, nil).Once()

				return mockTelemetry, mockConfig, baseMock
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			telemetry, config, baseMock := test.setup(s.T())
			transport := NewTransport(config, telemetry, baseMock)

			_, _ = transport.RoundTrip(req)

			baseMock.AssertExpectations(s.T())
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
