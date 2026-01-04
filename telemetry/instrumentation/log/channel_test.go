package log

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/log/noop"

	contractslog "github.com/goravel/framework/contracts/log"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockslog "github.com/goravel/framework/mocks/log"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/telemetry"
)

type TelemetryChannelTestSuite struct {
	suite.Suite
	mockConfig    *mocksconfig.Config
	mockTelemetry *mockstelemetry.Telemetry
	mockEntry     *mockslog.Entry
}

func TestTelemetryChannelTestSuite(t *testing.T) {
	suite.Run(t, new(TelemetryChannelTestSuite))
}

func (s *TelemetryChannelTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockTelemetry = mockstelemetry.NewTelemetry(s.T())
	s.mockEntry = mockslog.NewEntry(s.T())

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

	s.mockTelemetry.On("Logger", defaultInstrumentationName).Return(noop.NewLoggerProvider().Logger("test")).Once()

	s.mockEntry.On("Context").Return(context.Background())
	s.mockEntry.On("Time").Return(time.Now())
	s.mockEntry.On("Message").Return("test message")
	s.mockEntry.On("Level").Return(contractslog.InfoLevel)
	s.mockEntry.On("Code").Return("")
	s.mockEntry.On("Domain").Return("")
	s.mockEntry.On("Hint").Return("")
	s.mockEntry.On("Owner").Return(nil)
	s.mockEntry.On("User").Return(nil)
	s.mockEntry.On("With").Return(map[string]any{})
	s.mockEntry.On("Data").Return(map[string]any{})

	channel := NewTelemetryChannel(s.mockConfig)
	h, err := channel.Handle(channelPath)
	s.NoError(err)
	s.NoError(h.Handle(s.mockEntry))

	s.mockTelemetry.AssertExpectations(s.T())
}
