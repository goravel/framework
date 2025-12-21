package grpc

import (
	"testing"

	"github.com/stretchr/testify/suite"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc/stats"

	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/telemetry"
)

type HandlerTestSuite struct {
	suite.Suite
	originalFacade contractstelemetry.Telemetry
}

func (s *HandlerTestSuite) SetupTest() {
	s.originalFacade = telemetry.TelemetryFacade
}

func (s *HandlerTestSuite) TearDownTest() {
	telemetry.TelemetryFacade = s.originalFacade
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (s *HandlerTestSuite) TestServerStatsHandler() {
	tests := []struct {
		name   string
		setup  func(*mockstelemetry.Telemetry)
		assert func()
	}{
		{
			name: "returns nil when telemetry facade is nil",
			setup: func(_ *mockstelemetry.Telemetry) {
				telemetry.TelemetryFacade = nil
			},
			assert: func() {
				s.Nil(ServerStatsHandler())
			},
		},
		{
			name: "returns handler when telemetry facade is set",
			setup: func(mockTelemetry *mockstelemetry.Telemetry) {
				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider())
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider())
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator())

				telemetry.TelemetryFacade = mockTelemetry
			},
			assert: func() {
				s.NotNil(ServerStatsHandler())
			},
		},
		{
			name: "accepts options",
			setup: func(mockTelemetry *mockstelemetry.Telemetry) {
				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider())
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider())
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator())

				telemetry.TelemetryFacade = mockTelemetry
			},
			assert: func() {
				handler := ServerStatsHandler(
					WithFilter(func(info *stats.RPCTagInfo) bool { return true }),
					WithMessageEvents(ReceivedEvents, SentEvents),
				)
				s.NotNil(handler)
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockTelemetry := mockstelemetry.NewTelemetry(s.T())

			test.setup(mockTelemetry)
			test.assert()
		})
	}
}

func (s *HandlerTestSuite) TestClientStatsHandler() {
	tests := []struct {
		name   string
		setup  func(*mockstelemetry.Telemetry)
		assert func()
	}{
		{
			name: "returns nil when telemetry facade is nil",
			setup: func(_ *mockstelemetry.Telemetry) {
				telemetry.TelemetryFacade = nil
			},
			assert: func() {
				s.Nil(ClientStatsHandler())
			},
		},
		{
			name: "returns handler when telemetry facade is set",
			setup: func(mockTelemetry *mockstelemetry.Telemetry) {
				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider())
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider())
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator())

				telemetry.TelemetryFacade = mockTelemetry
			},
			assert: func() {
				s.NotNil(ClientStatsHandler())
			},
		},
		{
			name: "accepts options",
			setup: func(mockTelemetry *mockstelemetry.Telemetry) {
				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider())
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider())
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator())

				telemetry.TelemetryFacade = mockTelemetry
			},
			assert: func() {
				handler := ClientStatsHandler(
					WithSpanAttributes(),
					WithMetricAttributes(),
				)
				s.NotNil(handler)
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockTelemetry := mockstelemetry.NewTelemetry(s.T())

			test.setup(mockTelemetry)
			test.assert()
		})
	}
}