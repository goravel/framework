package log

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	lognoop "go.opentelemetry.io/otel/log/noop"

	contractslog "github.com/goravel/framework/contracts/log"
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

	telemetry.TelemetryFacade = s.mockTelemetry
}

func (s *TelemetryChannelTestSuite) TearDownTest() {
	telemetry.TelemetryFacade = nil
}

func (s *TelemetryChannelTestSuite) TestHandle_Factory_Success_DefaultName() {
	channelPath := "logging.channels.otel"

	s.mockConfig.EXPECT().GetBool("telemetry.instrumentation.log", true).Return(true).Once()
	s.mockConfig.EXPECT().GetString(channelPath+".instrument_name", defaultInstrumentationName).Return(defaultInstrumentationName).Once()

	channel := NewTelemetryChannel(s.mockConfig)
	h, err := channel.Handle(channelPath)

	s.NoError(err)
	s.NotNil(h)

	impl, ok := h.(*handler)
	s.True(ok)
	s.True(impl.enabled)
	s.Equal(defaultInstrumentationName, impl.instrumentName)
}

func (s *TelemetryChannelTestSuite) TestHandle_Factory_Success_CustomName() {
	channelPath := "logging.channels.otel"
	customName := "my-service-logs"

	s.mockConfig.EXPECT().GetBool("telemetry.instrumentation.log", true).Return(true).Once()
	s.mockConfig.EXPECT().GetString(channelPath+".instrument_name", defaultInstrumentationName).Return(customName).Once()

	channel := NewTelemetryChannel(s.mockConfig)
	h, err := channel.Handle(channelPath)

	s.NoError(err)
	s.NotNil(h)

	impl, ok := h.(*handler)
	s.True(ok)
	s.Equal(customName, impl.instrumentName)
}

func (s *TelemetryChannelTestSuite) TestHandle_Factory_Disabled() {
	s.mockConfig.EXPECT().GetBool("telemetry.instrumentation.log", true).Return(false).Once()

	channel := NewTelemetryChannel(s.mockConfig)
	h, err := channel.Handle("logging.channels.otel")

	s.NoError(err)
	s.NotNil(h)

	impl, ok := h.(*handler)
	s.True(ok)
	s.False(impl.enabled)
	s.False(h.Enabled(contractslog.LevelInfo))
}

func (s *TelemetryChannelTestSuite) TestHandle_Runtime_LazyLoading_TriggersTelemetry() {
	channelPath := "logging.channels.otel"

	s.mockConfig.EXPECT().GetBool("telemetry.instrumentation.log", true).Return(true).Once()
	s.mockConfig.EXPECT().GetString(channelPath+".instrument_name", defaultInstrumentationName).Return(defaultInstrumentationName).Once()

	s.mockTelemetry.On("Logger", defaultInstrumentationName).Return(lognoop.NewLoggerProvider().Logger("test")).Once()

	entry := &TestEntry{
		ctx:     context.Background(),
		level:   contractslog.LevelInfo,
		time:    time.Now(),
		message: "test message",
	}

	channel := NewTelemetryChannel(s.mockConfig)
	h, err := channel.Handle(channelPath)
	s.NoError(err)
	s.NoError(h.Handle(entry))

	s.mockTelemetry.AssertExpectations(s.T())
}
