package log

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/log/noop"

	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/telemetry"
)

type TelemetryChannelTestSuite struct {
	suite.Suite
	mockConfig    *mocksconfig.Config
	mockTelemetry *mockstelemetry.Telemetry
}

func TestTelemetryChannelTestSuite(t *testing.T) {
	suite.Run(t, new(TelemetryChannelTestSuite))
}

func (s *TelemetryChannelTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockTelemetry = mockstelemetry.NewTelemetry(s.T())

	telemetry.ConfigFacade = s.mockConfig
	telemetry.TelemetryFacade = s.mockTelemetry
}

func (s *TelemetryChannelTestSuite) TearDownTest() {
	telemetry.ConfigFacade = nil
	telemetry.TelemetryFacade = nil
}

func (s *TelemetryChannelTestSuite) TestHandle_Disabled() {
	s.mockConfig.EXPECT().GetBool("telemetry.instrumentation.log.enabled").Return(false).Once()

	channel := NewTelemetryChannel()
	hk, err := channel.Handle("logging.channels.otel")

	s.NoError(err)
	s.NotNil(hk)
	s.Nil(hk.Levels())
}

func (s *TelemetryChannelTestSuite) TestHandle_Enabled_DefaultName() {
	s.mockConfig.EXPECT().GetBool("telemetry.instrumentation.log.enabled").Return(true).Once()
	s.mockConfig.EXPECT().GetString("telemetry.instrumentation.log.name", defaultInstrumentationName).Return(defaultInstrumentationName).Once()

	s.mockTelemetry.On("Logger", defaultInstrumentationName).Return(noop.NewLoggerProvider().Logger("test")).Once()

	channel := NewTelemetryChannel()
	h, err := channel.Handle("logging.channels.otel")

	s.NoError(err)
	s.NotNil(h)
	s.NotEmpty(h.Levels())
	s.mockTelemetry.AssertExpectations(s.T())
}

func (s *TelemetryChannelTestSuite) TestHandle_Enabled_CustomName() {
	s.mockConfig.EXPECT().GetBool("telemetry.instrumentation.log.enabled").Return(true).Once()
	s.mockConfig.EXPECT().GetString("telemetry.instrumentation.log.name", defaultInstrumentationName).Return("my-service-logs").Once()

	s.mockTelemetry.On("Logger", "my-service-logs").Return(noop.NewLoggerProvider().Logger("test")).Once()

	channel := NewTelemetryChannel()
	h, err := channel.Handle("logging.channels.otel")

	s.NoError(err)
	s.NotNil(h)
	s.mockTelemetry.AssertExpectations(s.T())
}

func (s *TelemetryChannelTestSuite) TestHandle_Error_FacadeNotSet() {
	s.mockConfig.EXPECT().GetBool("telemetry.instrumentation.log.enabled").Return(true).Once()
	s.mockConfig.EXPECT().GetString("telemetry.instrumentation.log.name", defaultInstrumentationName).Return("app").Once()

	telemetry.TelemetryFacade = nil

	channel := NewTelemetryChannel()
	h, err := channel.Handle("logging.channels.otel")

	s.ErrorIs(err, errors.TelemetryFacadeNotSet)
	s.Nil(h)
}
