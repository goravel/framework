package grpc

import (
	"io"
	"testing"

	"github.com/stretchr/testify/suite"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc/stats"

	contractsconfig "github.com/goravel/framework/contracts/config"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/telemetry"
)

type HandlerTestSuite struct {
	suite.Suite
	originalTelemetryFacade contractstelemetry.Telemetry
	originalConfigFacade    contractsconfig.Config
}

func (s *HandlerTestSuite) SetupTest() {
	s.originalTelemetryFacade = telemetry.TelemetryFacade
	s.originalConfigFacade = telemetry.ConfigFacade
}

func (s *HandlerTestSuite) TearDownTest() {
	telemetry.TelemetryFacade = s.originalTelemetryFacade
	telemetry.ConfigFacade = s.originalConfigFacade
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (s *HandlerTestSuite) TestServerStatsHandler() {
	tests := []struct {
		name   string
		setup  func(*mockstelemetry.Telemetry, *mocksconfig.Config)
		assert func()
	}{
		{
			name: "returns nil immediately if config is disabled",
			setup: func(_ *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				telemetry.ConfigFacade = mockConfig
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_server", true).Return(false).Once()
			},
			assert: func() {
				s.Nil(NewServerStatsHandler())
			},
		},
		{
			name: "returns nil and logs warning when config enabled but telemetry facade is nil",
			setup: func(_ *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				telemetry.ConfigFacade = mockConfig
				telemetry.TelemetryFacade = nil

				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_server", true).Return(true).Once()
			},
			assert: func() {
				var handler stats.Handler
				out := color.CaptureOutput(func(w io.Writer) {
					handler = NewServerStatsHandler()
				})

				s.Nil(handler)
				s.Contains(out, "[Telemetry] Facade not initialized. gRPC server stats instrumentation is disabled.")
			},
		},
		{
			name: "returns handler when enabled and facade is set",
			setup: func(mockTelemetry *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				telemetry.ConfigFacade = mockConfig
				telemetry.TelemetryFacade = mockTelemetry

				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_server", true).Return(true).Once()

				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()
			},
			assert: func() {
				s.NotNil(NewServerStatsHandler())
			},
		},
		{
			name: "accepts options",
			setup: func(mockTelemetry *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				telemetry.ConfigFacade = mockConfig
				telemetry.TelemetryFacade = mockTelemetry

				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_server", true).Return(true).Once()

				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()
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
			mockTelemetry := mockstelemetry.NewTelemetry(s.T())
			mockConfig := mocksconfig.NewConfig(s.T())

			test.setup(mockTelemetry, mockConfig)
			test.assert()
		})
	}
}

func (s *HandlerTestSuite) TestClientStatsHandler() {
	tests := []struct {
		name   string
		setup  func(*mockstelemetry.Telemetry, *mocksconfig.Config)
		assert func()
	}{
		{
			name: "returns nil immediately if config is disabled",
			setup: func(_ *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				telemetry.ConfigFacade = mockConfig
				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_client", true).Return(false).Once()
			},
			assert: func() {
				s.Nil(NewClientStatsHandler())
			},
		},
		{
			name: "returns nil and logs warning when telemetry facade is nil",
			setup: func(_ *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				telemetry.ConfigFacade = mockConfig
				telemetry.TelemetryFacade = nil

				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_client", true).Return(true).Once()
			},
			assert: func() {
				var handler stats.Handler
				out := color.CaptureOutput(func(w io.Writer) {
					handler = NewClientStatsHandler()
				})

				s.Nil(handler)
				s.Contains(out, "[Telemetry] Facade not initialized. gRPC client stats instrumentation is disabled.")
			},
		},
		{
			name: "returns handler when telemetry facade is set",
			setup: func(mockTelemetry *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				telemetry.ConfigFacade = mockConfig
				telemetry.TelemetryFacade = mockTelemetry

				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_client", true).Return(true).Once()

				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()
			},
			assert: func() {
				s.NotNil(NewClientStatsHandler())
			},
		},
		{
			name: "accepts options",
			setup: func(mockTelemetry *mockstelemetry.Telemetry, mockConfig *mocksconfig.Config) {
				telemetry.ConfigFacade = mockConfig
				telemetry.TelemetryFacade = mockTelemetry

				mockConfig.EXPECT().GetBool("telemetry.instrumentation.grpc_client", true).Return(true).Once()

				mockTelemetry.EXPECT().TracerProvider().Return(tracenoop.NewTracerProvider()).Once()
				mockTelemetry.EXPECT().MeterProvider().Return(metricnoop.NewMeterProvider()).Once()
				mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()
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
			mockTelemetry := mockstelemetry.NewTelemetry(s.T())
			mockConfig := mocksconfig.NewConfig(s.T())

			test.setup(mockTelemetry, mockConfig)
			test.assert()
		})
	}
}
