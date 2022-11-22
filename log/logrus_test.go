package log

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var singleLog = "storage/logs/goravel.log"
var dailyLog = fmt.Sprintf("storage/logs/goravel-%s.log", time.Now().Format("2006-01-02"))

func initMockConfig() *configmocks.Config {
	mockConfig := &configmocks.Config{}
	facades.Config = mockConfig

	mockConfig.On("GetString", "logging.default").Return("stack").Once()
	mockConfig.On("GetString", "logging.channels.stack.driver").Return("stack").Once()
	mockConfig.On("Get", "logging.channels.stack.channels").Return([]string{"single", "daily"}).Once()
	mockConfig.On("GetString", "logging.channels.daily.driver").Return("daily").Once()
	mockConfig.On("GetString", "logging.channels.daily.path").Return(singleLog).Once()
	mockConfig.On("GetInt", "logging.channels.daily.days").Return(7).Once()
	mockConfig.On("GetString", "logging.channels.single.driver").Return("single").Once()
	mockConfig.On("GetString", "logging.channels.single.path").Return(singleLog).Once()

	return mockConfig
}

func mockDriverConfig(mockConfig *configmocks.Config) {
	mockConfig.On("GetString", "logging.channels.daily.level").Return("debug").Once()
	mockConfig.On("GetString", "logging.channels.single.level").Return("debug").Once()
	mockConfig.On("GetString", "app.timezone").Return("UTC")
	mockConfig.On("GetString", "app.env").Return("test")
}

func initFacadesLog() {
	logrusInstance := logrusInstance()
	facades.Log = NewLogrus(logrusInstance, NewWriter(logrusInstance.WithContext(context.Background())))
}

type LogrusTestSuite struct {
	suite.Suite
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(LogrusTestSuite))
}

func (s *LogrusTestSuite) SetupTest() {

}

func (s *LogrusTestSuite) TestLogrus() {
	var mockConfig *configmocks.Config

	beforeEach := func() {
		mockConfig = initMockConfig()
	}

	tests := []struct {
		name   string
		setup  func()
		assert func(name string)
	}{
		{
			name: "WithContext",
			setup: func() {
				mockConfig.On("GetString", "logging.channels.daily.level").Return("debug").Once()
				mockConfig.On("GetString", "logging.channels.single.level").Return("debug").Once()

				initFacadesLog()
			},
			assert: func(name string) {
				writer := facades.Log.WithContext(context.Background())
				assert.Equal(s.T(), reflect.TypeOf(writer).String(), reflect.TypeOf(&Writer{}).String(), name)
			},
		},
		{
			name: "Debug",
			setup: func() {
				mockDriverConfig(mockConfig)

				initFacadesLog()
				facades.Log.Debug("Goravel")
			},
			assert: func(name string) {
				assert.True(s.T(), file.Exists(dailyLog))
				assert.True(s.T(), file.Exists(singleLog))
				assert.True(s.T(), file.Contain(singleLog, "test.debug: Goravel"))
				assert.True(s.T(), file.Contain(dailyLog, "test.debug: Goravel"))
			},
		},
		{
			name: "No Debug",
			setup: func() {
				mockConfig.On("GetString", "logging.channels.daily.level").Return("info").Once()
				mockConfig.On("GetString", "logging.channels.single.level").Return("info").Once()

				initFacadesLog()
				facades.Log.Debug("Goravel")
			},
			assert: func(name string) {
				assert.False(s.T(), file.Exists(dailyLog))
				assert.False(s.T(), file.Exists(singleLog))
			},
		},
		{
			name: "Debugf",
			setup: func() {
				mockDriverConfig(mockConfig)

				initFacadesLog()
				facades.Log.Debugf("Goravel: %s", "World")
			},
			assert: func(name string) {
				assert.True(s.T(), file.Exists(dailyLog))
				assert.True(s.T(), file.Exists(singleLog))
				assert.True(s.T(), file.Contain(singleLog, "test.debug: Goravel: World"))
				assert.True(s.T(), file.Contain(dailyLog, "test.debug: Goravel: World"))
			},
		},
		{
			name: "Info",
			setup: func() {
				mockDriverConfig(mockConfig)

				initFacadesLog()
				facades.Log.Info("Goravel")
			},
			assert: func(name string) {
				assert.True(s.T(), file.Exists(dailyLog))
				assert.True(s.T(), file.Exists(singleLog))
				assert.True(s.T(), file.Contain(singleLog, "test.info: Goravel"))
				assert.True(s.T(), file.Contain(dailyLog, "test.info: Goravel"))
			},
		},
		{
			name: "Infof",
			setup: func() {
				mockDriverConfig(mockConfig)

				initFacadesLog()
				facades.Log.Infof("Goravel: %s", "World")
			},
			assert: func(name string) {
				assert.True(s.T(), file.Exists(dailyLog))
				assert.True(s.T(), file.Exists(singleLog))
				assert.True(s.T(), file.Contain(singleLog, "test.info: Goravel: World"))
				assert.True(s.T(), file.Contain(dailyLog, "test.info: Goravel: World"))
			},
		},
		{
			name: "Warning",
			setup: func() {
				mockDriverConfig(mockConfig)

				initFacadesLog()
				facades.Log.Warning("Goravel")
			},
			assert: func(name string) {
				assert.True(s.T(), file.Exists(dailyLog))
				assert.True(s.T(), file.Exists(singleLog))
				assert.True(s.T(), file.Contain(singleLog, "test.warning: Goravel"))
				assert.True(s.T(), file.Contain(dailyLog, "test.warning: Goravel"))
			},
		},
		{
			name: "Warningf",
			setup: func() {
				mockDriverConfig(mockConfig)

				initFacadesLog()
				facades.Log.Warningf("Goravel: %s", "World")
			},
			assert: func(name string) {
				assert.True(s.T(), file.Exists(dailyLog))
				assert.True(s.T(), file.Exists(singleLog))
				assert.True(s.T(), file.Contain(singleLog, "test.warning: Goravel: World"))
				assert.True(s.T(), file.Contain(dailyLog, "test.warning: Goravel: World"))
			},
		},
		{
			name: "Error",
			setup: func() {
				mockDriverConfig(mockConfig)

				initFacadesLog()
				facades.Log.Error("Goravel")
			},
			assert: func(name string) {
				assert.True(s.T(), file.Exists(dailyLog))
				assert.True(s.T(), file.Exists(singleLog))
				assert.True(s.T(), file.Contain(singleLog, "test.error: Goravel"))
				assert.True(s.T(), file.Contain(dailyLog, "test.error: Goravel"))
			},
		},
		{
			name: "Errorf",
			setup: func() {
				mockDriverConfig(mockConfig)

				initFacadesLog()
				facades.Log.Errorf("Goravel: %s", "World")
			},
			assert: func(name string) {
				assert.True(s.T(), file.Exists(dailyLog))
				assert.True(s.T(), file.Exists(singleLog))
				assert.True(s.T(), file.Contain(singleLog, "test.error: Goravel: World"))
				assert.True(s.T(), file.Contain(dailyLog, "test.error: Goravel: World"))
			},
		},
		{
			name: "Panic",
			setup: func() {
				mockDriverConfig(mockConfig)

				initFacadesLog()
			},
			assert: func(name string) {
				assert.Panics(s.T(), func() {
					facades.Log.Panic("Goravel")
				})
				assert.True(s.T(), file.Exists(dailyLog))
				assert.True(s.T(), file.Exists(singleLog))
				assert.True(s.T(), file.Contain(singleLog, "test.panic: Goravel"))
				assert.True(s.T(), file.Contain(dailyLog, "test.panic: Goravel"))
			},
		},
		{
			name: "Panicf",
			setup: func() {
				mockDriverConfig(mockConfig)

				initFacadesLog()
			},
			assert: func(name string) {
				assert.Panics(s.T(), func() {
					facades.Log.Panicf("Goravel: %s", "World")
				})
				assert.True(s.T(), file.Exists(dailyLog))
				assert.True(s.T(), file.Exists(singleLog))
				assert.True(s.T(), file.Contain(singleLog, "test.panic: Goravel: World"))
				assert.True(s.T(), file.Contain(dailyLog, "test.panic: Goravel: World"))
			},
		},
	}

	for _, test := range tests {
		beforeEach()
		test.setup()
		test.assert(test.name)
		mockConfig.AssertExpectations(s.T())
		file.Remove("storage")
	}
}

func (s *LogrusTestSuite) TestTestWriter() {
	facades.Log = NewLogrus(nil, NewTestWriter())
	assert.Equal(s.T(), facades.Log.WithContext(context.Background()), &TestWriter{})
	assert.NotPanics(s.T(), func() {
		facades.Log.Debug("Goravel")
		facades.Log.Debugf("Goravel")
		facades.Log.Info("Goravel")
		facades.Log.Infof("Goravel")
		facades.Log.Warning("Goravel")
		facades.Log.Warningf("Goravel")
		facades.Log.Error("Goravel")
		facades.Log.Errorf("Goravel")
		facades.Log.Fatal("Goravel")
		facades.Log.Fatalf("Goravel")
		facades.Log.Panic("Goravel")
		facades.Log.Panicf("Goravel")
	})
}

func TestLogrus_Fatal(t *testing.T) {
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	initFacadesLog()

	if os.Getenv("FATAL") == "1" {
		facades.Log.Fatal("Goravel")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLogrus_Fatal")
	cmd.Env = append(os.Environ(), "FATAL=1")
	err := cmd.Run()

	assert.EqualError(t, err, "exit status 1")
	assert.True(t, file.Exists(dailyLog))
	assert.True(t, file.Exists(singleLog))
	assert.True(t, file.Contain(singleLog, "test.fatal: Goravel"))
	assert.True(t, file.Contain(dailyLog, "test.fatal: Goravel"))
	file.Remove("storage")
}

func TestLogrus_Fatalf(t *testing.T) {
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	initFacadesLog()

	if os.Getenv("FATAL") == "1" {
		facades.Log.Fatalf("Goravel")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLogrus_Fatal")
	cmd.Env = append(os.Environ(), "FATAL=1")
	err := cmd.Run()

	assert.EqualError(t, err, "exit status 1")
	assert.True(t, file.Exists(dailyLog))
	assert.True(t, file.Exists(singleLog))
	assert.True(t, file.Contain(singleLog, "test.fatal: Goravel"))
	assert.True(t, file.Contain(dailyLog, "test.fatal: Goravel"))
	file.Remove("storage")
}
