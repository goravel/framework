package grpc

import (
	"io"
	"testing"

	"github.com/stretchr/testify/suite"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc/stats"

	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/telemetry"
)

type HandlerTestSuite struct {
	suite.Suite
	mockTelemetry *mockstelemetry.Telemetry
	mockConfig    *mocksconfig.Config
}

func (s *HandlerTestSuite) SetupTest() {
	s.mockTelemetry = mockstelemetry.NewTelemetry(s.T())
	s.mockConfig = mocksconfig.NewConfig(s.T())
}

func (s *HandlerTestSuite) TearDownTest() {
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (s *HandlerTestSuite) TestServerStatsHandler() {
	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "returns nil and logs warning when telemetry facade is nil",
			setup: func() {
				telemetry.Facade = nil
			},
			assert: func() {
				var handler stats.Handler
				out := color.CaptureOutput(func(w io.Writer) {
					handler = NewServerStatsHandler()
				})

				s.Nil(handler)
				s.Contains(out, errors.TelemetryGrpcServerStatsHandlerDisabled.Error())
			},
		},
		{
			name: "Returns nil if config is disabled",
			setup: func() {
				telemetry.ConfigFacade = nil
			},
			assert: func() {
				s.Nil(NewServerStatsHandler())
			},
		},
		{
			name: "Accepts options",
			setup: func() {
				s.mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				s.mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				s.mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()
				s.mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_server.enabled").Return(true).Once()
			},
			assert: func() {
				handler := NewServerStatsHandler(
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
			telemetry.Facade = s.mockTelemetry
			telemetry.ConfigFacade = s.mockConfig
			test.setup()
			test.assert()
		})
	}
}

func (s *HandlerTestSuite) TestClientStatsHandler() {
	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "returns nil and logs warning when telemetry facade is nil",
			setup: func() {
				telemetry.Facade = nil
			},
			assert: func() {
				var handler stats.Handler
				out := color.CaptureOutput(func(w io.Writer) {
					handler = NewClientStatsHandler()
				})

				s.Nil(handler)
				s.Contains(out, errors.TelemetryGrpcClientStatsHandlerDisabled.Error())
			},
		},
		{
			name: "Returns nil if config is disabled",
			setup: func() {
				telemetry.ConfigFacade = nil
			},
			assert: func() {
				s.Nil(NewServerStatsHandler())
			},
		},
		{
			name: "Accepts options",
			setup: func() {
				s.mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				s.mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				s.mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()
				s.mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_client.enabled").Return(true).Once()
			},
			assert: func() {
				handler := NewClientStatsHandler(
					WithSpanAttributes(),
					WithMetricAttributes(),
				)
				s.NotNil(handler)
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			telemetry.Facade = s.mockTelemetry
			telemetry.ConfigFacade = s.mockConfig
			test.setup()
			test.assert()
		})
	}
}
