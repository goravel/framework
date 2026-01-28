package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/support/file"
	telemetrylog "github.com/goravel/framework/telemetry/instrumentation/log"
)

// clearChannelCache clears all entries from the channel cache
func clearChannelCache() {
	channelToHandlers.Range(func(key, value any) bool {
		channelToHandlers.Delete(key)
		return true
	})
}

func TestNewApplication(t *testing.T) {
	j := json.New()
	telemetryResolver := func() contractstelemetry.Telemetry {
		return mockstelemetry.NewTelemetry(t)
	}

	app, err := NewApplication(context.Background(), nil, nil, j, telemetryResolver)
	assert.Nil(t, err)
	assert.NotNil(t, app)

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("logging.default").Return("test").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.driver").Return("single").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.path").Return("test").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.level").Return("debug").Times(2)            // Called for file handler and console handler when print=true
	mockConfig.EXPECT().GetString("logging.channels.test.formatter", "text").Return("text").Times(2) // Called for file handler and console handler when print=true
	mockConfig.EXPECT().GetBool("logging.channels.test.print").Return(true).Once()
	app, err = NewApplication(context.Background(), nil, mockConfig, j, telemetryResolver)
	assert.Nil(t, err)
	assert.NotNil(t, app)

	// Clear cache before testing unsupported driver
	clearChannelCache()

	mockConfig = mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("logging.default").Return("test").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.driver").Return("test").Once()

	app, err = NewApplication(context.Background(), nil, mockConfig, j, telemetryResolver)
	assert.EqualError(t, err, errors.LogDriverNotSupported.Args("test").Error())
	assert.Nil(t, app)

	// Cleanup test files
	_ = file.Remove("test")
	_ = file.Remove("dummy")
}

func TestApplication_Channel(t *testing.T) {
	// Clear cache before test
	clearChannelCache()

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("logging.default").Return("test").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.driver").Return("single").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.path").Return("test").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.level").Return("debug").Times(2)            // Called for file handler and console handler
	mockConfig.EXPECT().GetString("logging.channels.test.formatter", "text").Return("text").Times(2) // Called for file handler and console handler
	mockConfig.EXPECT().GetBool("logging.channels.test.print").Return(true).Once()

	app, err := NewApplication(context.Background(), nil, mockConfig, json.New(), nil)
	assert.Nil(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, app.Channel(""))

	mockConfig.EXPECT().GetString("logging.channels.dummy.driver").Return("daily").Once()
	mockConfig.EXPECT().GetString("logging.channels.dummy.path").Return("dummy").Once()
	mockConfig.EXPECT().GetString("logging.channels.dummy.level").Return("debug").Times(2)            // Called for file handler and console handler
	mockConfig.EXPECT().GetString("logging.channels.dummy.formatter", "text").Return("text").Times(2) // Called for file handler and console handler
	mockConfig.EXPECT().GetBool("logging.channels.dummy.print").Return(true).Once()
	mockConfig.EXPECT().GetInt("logging.channels.dummy.days").Return(1).Once()
	writer := app.Channel("dummy")
	assert.NotNil(t, writer)

	mockConfig.EXPECT().GetString("logging.channels.test2.driver").Return("test2").Once()
	mockConfig.EXPECT().GetString("app.env").Return("test").Twice()
	// When an error occurs, Channel returns the original app instead of panicking
	result := app.Channel("test2")
	assert.Equal(t, app, result)

	// Cleanup test files
	_ = file.Remove("test")
	_ = file.Remove("dummy")
}

func TestApplication_Stack(t *testing.T) {
	// Clear cache before test
	clearChannelCache()

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("logging.default").Return("test").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.driver").Return("single").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.path").Return("test").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.level").Return("debug").Times(2)            // Called for file handler and console handler
	mockConfig.EXPECT().GetString("logging.channels.test.formatter", "text").Return("text").Times(2) // Called for file handler and console handler
	mockConfig.EXPECT().GetBool("logging.channels.test.print").Return(true).Once()
	app, err := NewApplication(context.Background(), nil, mockConfig, json.New(), nil)

	assert.Nil(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, app.Stack([]string{}))

	mockConfig.EXPECT().GetString("logging.channels.test2.driver").Return("test2").Once()
	mockConfig.EXPECT().GetString("app.env").Return("test").Twice()
	// When an error occurs, Stack returns the original app instead of panicking
	result := app.Stack([]string{"test2"})
	assert.Equal(t, app, result)

	mockConfig.EXPECT().GetString("logging.channels.dummy.driver").Return("daily").Once()
	mockConfig.EXPECT().GetString("logging.channels.dummy.path").Return("dummy").Once()
	mockConfig.EXPECT().GetString("logging.channels.dummy.level").Return("debug").Times(2)            // Called for file handler and console handler
	mockConfig.EXPECT().GetString("logging.channels.dummy.formatter", "text").Return("text").Times(2) // Called for file handler and console handler
	mockConfig.EXPECT().GetBool("logging.channels.dummy.print").Return(true).Once()
	mockConfig.EXPECT().GetInt("logging.channels.dummy.days").Return(1).Once()
	assert.NotNil(t, app.Stack([]string{"dummy"}))

	// Cleanup test files
	_ = file.Remove("test")
	_ = file.Remove("dummy")
}

func TestGetHandlers(t *testing.T) {
	j := json.New()

	t.Run("single driver", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("logging.channels.test-single.driver").Return("single").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-single.path").Return("test-single.log").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-single.level").Return("info").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-single.formatter", "text").Return("text").Once()
		mockConfig.EXPECT().GetBool("logging.channels.test-single.print").Return(false).Once()

		handlers, err := getHandlers(mockConfig, j, nil, "test-single")
		assert.NoError(t, err)
		assert.Len(t, handlers, 1)

		// Cleanup
		_ = file.Remove("test-single.log")
		clearChannelCache()
	})

	t.Run("single driver with print enabled", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("logging.channels.test-single-print.driver").Return("single").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-single-print.path").Return("test-single-print.log").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-single-print.level").Return("debug").Times(2) // file + console
		mockConfig.EXPECT().GetString("logging.channels.test-single-print.formatter", "text").Return("text").Times(2)
		mockConfig.EXPECT().GetBool("logging.channels.test-single-print.print").Return(true).Once()

		handlers, err := getHandlers(mockConfig, j, nil, "test-single-print")
		assert.NoError(t, err)
		assert.Len(t, handlers, 2) // file handler + console handler

		// Cleanup
		_ = file.Remove("test-single-print.log")
		clearChannelCache()
	})

	t.Run("daily driver", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("logging.channels.test-daily.driver").Return("daily").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-daily.path").Return("test-daily.log").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-daily.level").Return("warning").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-daily.formatter", "text").Return("json").Once()
		mockConfig.EXPECT().GetBool("logging.channels.test-daily.print").Return(false).Once()
		mockConfig.EXPECT().GetInt("logging.channels.test-daily.days").Return(7).Once()

		handlers, err := getHandlers(mockConfig, j, nil, "test-daily")
		assert.NoError(t, err)
		assert.Len(t, handlers, 1)

		// Cleanup
		clearChannelCache()
	})

	t.Run("daily driver with print enabled", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("logging.channels.test-daily-print.driver").Return("daily").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-daily-print.path").Return("test-daily-print.log").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-daily-print.level").Return("error").Times(2)
		mockConfig.EXPECT().GetString("logging.channels.test-daily-print.formatter", "text").Return("text").Times(2)
		mockConfig.EXPECT().GetBool("logging.channels.test-daily-print.print").Return(true).Once()
		mockConfig.EXPECT().GetInt("logging.channels.test-daily-print.days").Return(3).Once()

		handlers, err := getHandlers(mockConfig, j, nil, "test-daily-print")
		assert.NoError(t, err)
		assert.Len(t, handlers, 2)

		// Cleanup
		clearChannelCache()
	})

	t.Run("OTeL driver", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetBool("telemetry.instrumentation.log", true).Return(true).Once()
		mockConfig.EXPECT().GetString("logging.channels.test-otel.driver").Return("otel").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-otel.instrument_name", telemetrylog.DefaultInstrumentationName).
			Return("goravel").Once()
		mockConfig.EXPECT().GetBool("logging.channels.test-otel.print").Return(false).Once()

		mockTelemetry := mockstelemetry.NewTelemetry(t)
		resolver := func() contractstelemetry.Telemetry {
			return mockTelemetry
		}

		handlers, err := getHandlers(mockConfig, j, resolver, "test-otel")
		assert.NoError(t, err)
		assert.Len(t, handlers, 1)

		clearChannelCache()
	})

	t.Run("OTeL driver with print enabled", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetBool("telemetry.instrumentation.log", true).Return(true).Once()
		mockConfig.EXPECT().GetString("logging.channels.test-otel-print.driver").Return("otel").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-otel-print.instrument_name", telemetrylog.DefaultInstrumentationName).
			Return("goravel").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-otel-print.level").Return("debug").Once()
		mockConfig.EXPECT().GetBool("logging.channels.test-otel-print.print").Return(true).Once()
		mockConfig.EXPECT().GetString("logging.channels.test-otel-print.formatter", "text").
			Return("text").Once()

		mockTelemetry := mockstelemetry.NewTelemetry(t)
		resolver := func() contractstelemetry.Telemetry {
			return mockTelemetry
		}

		handlers, err := getHandlers(mockConfig, j, resolver, "test-otel-print")
		assert.NoError(t, err)
		assert.Len(t, handlers, 2)

		clearChannelCache()
	})

	t.Run("stack driver", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("logging.channels.test-stack.driver").Return("stack").Once()
		mockConfig.EXPECT().Get("logging.channels.test-stack.channels").Return([]string{"stack-single1", "stack-single2"}).Once()

		// Expectations for stack-single1 channel
		mockConfig.EXPECT().GetString("logging.channels.stack-single1.driver").Return("single").Once()
		mockConfig.EXPECT().GetString("logging.channels.stack-single1.path").Return("test-stack-single1.log").Once()
		mockConfig.EXPECT().GetString("logging.channels.stack-single1.level").Return("debug").Once()
		mockConfig.EXPECT().GetString("logging.channels.stack-single1.formatter", "text").Return("text").Once()
		mockConfig.EXPECT().GetBool("logging.channels.stack-single1.print").Return(false).Once()

		// Expectations for stack-single2 channel
		mockConfig.EXPECT().GetString("logging.channels.stack-single2.driver").Return("single").Once()
		mockConfig.EXPECT().GetString("logging.channels.stack-single2.path").Return("test-stack-single2.log").Once()
		mockConfig.EXPECT().GetString("logging.channels.stack-single2.level").Return("info").Once()
		mockConfig.EXPECT().GetString("logging.channels.stack-single2.formatter", "text").Return("text").Once()
		mockConfig.EXPECT().GetBool("logging.channels.stack-single2.print").Return(false).Once()

		handlers, err := getHandlers(mockConfig, j, nil, "test-stack")
		assert.NoError(t, err)
		assert.Len(t, handlers, 2) // One from each channel

		// Cleanup
		_ = file.Remove("test-stack-single1.log")
		_ = file.Remove("test-stack-single2.log")
		clearChannelCache()
	})

	t.Run("stack driver with circular reference", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("logging.channels.test-circular.driver").Return("stack").Once()
		mockConfig.EXPECT().Get("logging.channels.test-circular.channels").Return([]string{"test-circular"}).Once()

		handlers, err := getHandlers(mockConfig, j, nil, "test-circular")
		assert.Error(t, err)
		assert.EqualError(t, err, errors.LogDriverCircularReference.Args("stack").Error())
		assert.Nil(t, handlers)

		// Cleanup
		clearChannelCache()
	})

	t.Run("stack driver with invalid channels config", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("logging.channels.test-badstack.driver").Return("stack").Once()
		mockConfig.EXPECT().Get("logging.channels.test-badstack.channels").Return("not-a-slice").Once()

		handlers, err := getHandlers(mockConfig, j, nil, "test-badstack")
		assert.Error(t, err)
		assert.EqualError(t, err, errors.LogChannelNotFound.Args("test-badstack").Error())
		assert.Nil(t, handlers)

		// Cleanup
		clearChannelCache()
	})

	t.Run("custom driver", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("logging.channels.test-custom.driver").Return("custom").Once()

		customLogger := &CustomLogger{}
		mockConfig.EXPECT().Get("logging.channels.test-custom.via").Return(customLogger).Once()

		handlers, err := getHandlers(mockConfig, j, nil, "test-custom")
		assert.NoError(t, err)
		assert.Len(t, handlers, 1)

		// Cleanup
		clearChannelCache()
	})

	t.Run("custom driver without logger", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("logging.channels.test-badcustom.driver").Return("custom").Once()
		mockConfig.EXPECT().Get("logging.channels.test-badcustom.via").Return(nil).Once()

		handlers, err := getHandlers(mockConfig, j, nil, "test-badcustom")
		assert.Error(t, err)
		assert.EqualError(t, err, errors.LogChannelUnimplemented.Args("test-badcustom").Error())
		assert.Nil(t, handlers)

		// Cleanup
		clearChannelCache()
	})

	t.Run("unsupported driver", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("logging.channels.test-unknown.driver").Return("unknown").Once()

		handlers, err := getHandlers(mockConfig, j, nil, "test-unknown")
		assert.Error(t, err)
		assert.EqualError(t, err, errors.LogDriverNotSupported.Args("test-unknown").Error())
		assert.Nil(t, handlers)

		// Cleanup
		clearChannelCache()
	})

	t.Run("cached handlers", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("logging.channels.test-cached.driver").Return("single").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-cached.path").Return("test-cached.log").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-cached.level").Return("info").Once()
		mockConfig.EXPECT().GetString("logging.channels.test-cached.formatter", "text").Return("text").Once()
		mockConfig.EXPECT().GetBool("logging.channels.test-cached.print").Return(false).Once()

		// First call should use mock config
		handlers1, err := getHandlers(mockConfig, j, nil, "test-cached")
		assert.NoError(t, err)
		assert.Len(t, handlers1, 1)

		// Second call should use cached handlers (no mock expectations needed)
		handlers2, err := getHandlers(mockConfig, j, nil, "test-cached")
		assert.NoError(t, err)
		assert.Len(t, handlers2, 1)
		assert.Equal(t, handlers1, handlers2)

		// Cleanup
		_ = file.Remove("test-cached.log")
		clearChannelCache()
	})
}
