package http

import (
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
)

type TransportTestSuite struct {
	suite.Suite
}

func TestTransportTestSuite(t *testing.T) {
	suite.Run(t, new(TransportTestSuite))
}

func (s *TransportTestSuite) TestNewTransport() {
	baseTransport := &http.Transport{}

	tests := []struct {
		name          string
		setup         func(t *testing.T) (contractstelemetry.Telemetry, contractsconfig.Config)
		expectWrapped bool
	}{
		{
			name: "Fallback: Returns base when Config is nil",
			setup: func(t *testing.T) (contractstelemetry.Telemetry, contractsconfig.Config) {
				return mockstelemetry.NewTelemetry(t), nil
			},
			expectWrapped: false,
		},
		{
			name: "Fallback: Returns base when Telemetry is nil",
			setup: func(t *testing.T) (contractstelemetry.Telemetry, contractsconfig.Config) {
				return nil, mocksconfig.NewConfig(t)
			},
			expectWrapped: false,
		},
		{
			name: "Kill Switch: Returns base when http_client is disabled",
			setup: func(t *testing.T) (contractstelemetry.Telemetry, contractsconfig.Config) {
				mockConfig := mocksconfig.NewConfig(t)
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.http_client", true).Return(false).Once()
				return mockstelemetry.NewTelemetry(t), mockConfig
			},
			expectWrapped: false,
		},
		{
			name: "Success: Returns wrapped transport when enabled",
			setup: func(t *testing.T) (contractstelemetry.Telemetry, contractsconfig.Config) {
				mockConfig := mocksconfig.NewConfig(t)
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.http_client", true).Return(true).Once()

				mockTelemetry := mockstelemetry.NewTelemetry(t)
				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()

				return mockTelemetry, mockConfig
			},
			expectWrapped: true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			telemetry, config := test.setup(s.T())
			result := NewTransport(config, telemetry, baseTransport)
			s.Equal(test.expectWrapped, baseTransport != result, "Transport wrapping mismatch")
		})
	}
}
