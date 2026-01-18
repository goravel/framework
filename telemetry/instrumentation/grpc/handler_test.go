package grpc

import (
	"testing"

	"github.com/stretchr/testify/suite"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc/stats"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/telemetry"
)

type HandlerTestSuite struct {
	suite.Suite
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (s *HandlerTestSuite) TestServerStatsHandler() {
	tests := []struct {
		name   string
		setup  func(*mockstelemetry.Telemetry, *mocksconfig.Config)
		assert func(*mockstelemetry.Telemetry, *mocksconfig.Config)
	}{
		{
			name: "Returns nil if config is disabled",
			setup: func(_ *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_server", true).Return(false).Once()
			},
			assert: func(t *mockstelemetry.Telemetry, c *mocksconfig.Config) {
				s.Nil(NewServerStatsHandler(c, t))
			},
		},
		{
			name: "Returns nil (no warning) if telemetry is nil",
			setup: func(_ *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_server", true).Return(true).Once()
			},
			assert: func(_ *mockstelemetry.Telemetry, c *mocksconfig.Config) {
				s.Nil(NewServerStatsHandler(c, nil))
			},
		},
		{
			name: "Returns handler when enabled and dependencies set",
			setup: func(mockTelemetry *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_server", true).Return(true).Once()
				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()
			},
			assert: func(t *mockstelemetry.Telemetry, c *mocksconfig.Config) {
				s.NotNil(NewServerStatsHandler(c, t))
			},
		},
		{
			name: "Accepts options",
			setup: func(mockTelemetry *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_server", true).Return(true).Once()
				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()
			},
			assert: func(t *mockstelemetry.Telemetry, c *mocksconfig.Config) {
				handler := NewServerStatsHandler(c, t,
					WithFilter(func(info *stats.RPCTagInfo) bool { return true }),
					WithMessageEvents(ReceivedEvents, SentEvents),
					WithMetricAttributes(telemetry.String("key", "value")),
				)
				s.NotNil(handler)
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockTelemetry := mockstelemetry.NewTelemetry(s.T())
			mockConfig := mocksconfig.NewConfig(s.T())

			if test.setup != nil {
				test.setup(mockTelemetry, mockConfig)
			}
			test.assert(mockTelemetry, mockConfig)
		})
	}
}

func (s *HandlerTestSuite) TestClientStatsHandler() {
	tests := []struct {
		name   string
		setup  func(*mockstelemetry.Telemetry, *mocksconfig.Config)
		assert func(*mockstelemetry.Telemetry, *mocksconfig.Config)
	}{
		{
			name: "Returns nil if config is disabled",
			setup: func(_ *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_client", true).Return(false).Once()
			},
			assert: func(t *mockstelemetry.Telemetry, c *mocksconfig.Config) {
				s.Nil(NewClientStatsHandler(c, t))
			},
		},
		{
			name: "Returns nil (no warning) if telemetry is nil",
			setup: func(_ *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_client", true).Return(true).Once()
			},
			assert: func(_ *mockstelemetry.Telemetry, c *mocksconfig.Config) {
				s.Nil(NewClientStatsHandler(c, nil))
			},
		},
		{
			name: "Returns handler when dependencies set",
			setup: func(mockTelemetry *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_client", true).Return(true).Once()
				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()
			},
			assert: func(t *mockstelemetry.Telemetry, c *mocksconfig.Config) {
				s.NotNil(NewClientStatsHandler(c, t))
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockTelemetry := mockstelemetry.NewTelemetry(s.T())
			mockConfig := mocksconfig.NewConfig(s.T())

			if test.setup != nil {
				test.setup(mockTelemetry, mockConfig)
			}
			test.assert(mockTelemetry, mockConfig)
		})
	}
}
