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

func (s *TelemetryChannelTestSuite) TestHandle_Success_DefaultName() {
	channelPath := "logging.channels.otel"
	s.mockConfig.EXPECT().GetString(channelPath+".name", defaultInstrumentationName).Return(defaultInstrumentationName).Once()

	s.mockTelemetry.On("Logger", defaultInstrumentationName).Return(noop.NewLoggerProvider().Logger("test")).Once()

	channel := NewTelemetryChannel()
	h, err := channel.Handle(channelPath)

	s.NoError(err)
	s.NotNil(h)
	s.NotEmpty(h.Levels())
	s.mockTelemetry.AssertExpectations(s.T())
}

func (s *TelemetryChannelTestSuite) TestHandle_Success_CustomName() {
	channelPath := "logging.channels.otel"
	customName := "my-service-logs"

	s.mockConfig.EXPECT().GetString(channelPath+".name", defaultInstrumentationName).Return(customName).Once()

	s.mockTelemetry.On("Logger", customName).Return(noop.NewLoggerProvider().Logger("test")).Once()

	channel := NewTelemetryChannel()
	h, err := channel.Handle(channelPath)

	s.NoError(err)
	s.NotNil(h)
	s.mockTelemetry.AssertExpectations(s.T())
}

func (s *TelemetryChannelTestSuite) TestHandle_Error_TelemetryFacadeNotSet() {
	telemetry.TelemetryFacade = nil

	channel := NewTelemetryChannel()
	h, err := channel.Handle("logging.channels.otel")

	s.ErrorIs(err, errors.TelemetryFacadeNotSet)
	s.Nil(h)
}

func (s *TelemetryChannelTestSuite) TestHandle_Error_ConfigFacadeNotSet() {
	telemetry.ConfigFacade = nil

	channel := NewTelemetryChannel()
	h, err := channel.Handle("logging.channels.otel")

	s.ErrorIs(err, errors.ConfigFacadeNotSet)
	s.Nil(h)
}
