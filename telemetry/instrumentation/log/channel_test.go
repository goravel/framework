package log

import (
	"testing"

	"github.com/stretchr/testify/suite"

	contractslog "github.com/goravel/framework/contracts/log"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
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
}

func (s *TelemetryChannelTestSuite) TestHandle() {
	channelPath := "logging.channels.otel"

	tests := []struct {
		name             string
		configSetup      func()
		expectedEnabled  bool
		expectedInstName string
	}{
		{
			name: "Success: Enabled with default instrumentation name",
			configSetup: func() {
				s.mockConfig.EXPECT().GetBool("telemetry.instrumentation.log", true).Return(true).Once()
				s.mockConfig.EXPECT().GetString(channelPath+".instrument_name", defaultInstrumentationName).Return(defaultInstrumentationName).Once()
			},
			expectedEnabled:  true,
			expectedInstName: defaultInstrumentationName,
		},
		{
			name: "Success: Enabled with custom instrumentation name",
			configSetup: func() {
				s.mockConfig.EXPECT().GetBool("telemetry.instrumentation.log", true).Return(true).Once()
				s.mockConfig.EXPECT().GetString(channelPath+".instrument_name", defaultInstrumentationName).Return("my-custom-logger").Once()
			},
			expectedEnabled:  true,
			expectedInstName: "my-custom-logger",
		},
		{
			name: "Success: Disabled via config",
			configSetup: func() {
				s.mockConfig.EXPECT().GetBool("telemetry.instrumentation.log", true).Return(false).Once()
			},
			expectedEnabled: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.configSetup()

			channel := NewTelemetryChannel(s.mockConfig, s.mockTelemetry)

			h, err := channel.Handle(channelPath)

			s.NoError(err)
			s.NotNil(h)

			impl, ok := h.(*handler)
			s.True(ok, "Handler should be of type *handler")
			s.Equal(tt.expectedEnabled, impl.enabled)

			if tt.expectedEnabled {
				s.Equal(tt.expectedInstName, impl.instrumentName)
				s.Equal(s.mockTelemetry, impl.telemetry)
			} else {
				s.False(h.Enabled(contractslog.LevelInfo))
			}
		})
	}
}
