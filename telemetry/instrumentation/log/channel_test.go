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
}

func TestTelemetryChannelTestSuite(t *testing.T) {
	suite.Run(t, new(TelemetryChannelTestSuite))
}

func (s *TelemetryChannelTestSuite) TestHandle() {
	const (
		channelPath  = "logging.channels.otel"
		telemetryKey = "telemetry.instrumentation.log"
	)

	tests := []struct {
		name             string
		setup            func(m *mocksconfig.Config)
		shouldBeEnabled  bool
		expectedInstName string
	}{
		{
			name: "Success: Telemetry enabled with custom name",
			setup: func(m *mocksconfig.Config) {
				m.EXPECT().GetBool(telemetryKey, true).Return(true).Once()
				m.EXPECT().GetString(channelPath+".instrument_name", DefaultInstrumentationName).
					Return("custom-app-logger").Once()
			},
			shouldBeEnabled:  true,
			expectedInstName: "custom-app-logger",
		},
		{
			name: "Success: Telemetry disabled via config",
			setup: func(m *mocksconfig.Config) {
				m.EXPECT().GetBool(telemetryKey, true).Return(false).Once()
			},
			shouldBeEnabled: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockConfig := mocksconfig.NewConfig(s.T())
			mockTelemetry := mockstelemetry.NewTelemetry(s.T())

			tt.setup(mockConfig)

			channel := NewTelemetryChannel(mockConfig, mockTelemetry)
			channelHandler, err := channel.Handle(channelPath)
			s.NoError(err)
			s.NotNil(channelHandler)

			s.Equal(tt.shouldBeEnabled, channelHandler.Enabled(contractslog.LevelInfo))

			if tt.shouldBeEnabled {
				impl, ok := channelHandler.(*handler)
				s.True(ok, "Returned handler must be of type *handler")

				s.Equal(tt.expectedInstName, impl.instrumentName, "Instrumentation name should match config")

				s.NotNil(impl.resolver, "Handler resolver should not be nil")
				s.Equal(mockTelemetry, impl.resolver(), "Resolver should return the injected telemetry service")
			}
		})
	}
}
