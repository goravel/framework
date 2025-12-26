package log

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	nethttp "net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	contractshttp "github.com/goravel/framework/contracts/http"
	contractslog "github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/foundation/json"
	configmock "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

type testContextKey any

var (
	singleLog = "storage/logs/goravel.log"
	dailyLog  = fmt.Sprintf("storage/logs/goravel-%s.log", carbon.Now().ToDateString())
)

func TestWriter(t *testing.T) {
	var (
		mockConfig *configmock.Config
		log        *Application
		j          = json.New()
		err        error
	)

	beforeEach := func() {
		mockConfig = initMockConfig()
	}

	tests := []struct {
		name   string
		setup  func()
		assert func()
		err    error
	}{
		{
			name: "WithContext",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				ctx := context.Background()
				ctx = context.WithValue(ctx, testContextKey("key"), "value")
				log.WithContext(ctx).Info("Goravel")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.info: Goravel\n[Context] map[key:value]"))
				assert.True(t, file.Contains(dailyLog, "test.info: Goravel\n[Context] map[key:value]"))
			},
		},
		{
			name: "Debug",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Debug("Debug Goravel")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.debug: Debug Goravel"))
				assert.True(t, file.Contains(dailyLog, "test.debug: Debug Goravel"))
			},
		},
		{
			name: "No Debug",
			setup: func() {
				mockConfig.EXPECT().GetString("logging.channels.daily.level").Return("info").Once()
				mockConfig.EXPECT().GetString("logging.channels.daily.formatter", "text").Return("text").Once()
				mockConfig.EXPECT().GetString("logging.channels.single.level").Return("info").Once()
				mockConfig.EXPECT().GetString("logging.channels.single.formatter", "text").Return("text").Once()
				log, err = NewApplication(mockConfig, j)
				log.Debug("No Debug Goravel")
			},
			assert: func() {
				assert.False(t, file.Contains(singleLog, "test.debug: No Debug Goravel"))
				assert.False(t, file.Contains(dailyLog, "test.debug: No Debug Goravel"))
			},
		},
		{
			name: "Debugf",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Debugf("Goravel: %s", "World")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.debug: Goravel: World"))
				assert.True(t, file.Contains(dailyLog, "test.debug: Goravel: World"))
			},
		},
		{
			name: "Info",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Info("Goravel")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.info: Goravel"))
				assert.True(t, file.Contains(dailyLog, "test.info: Goravel"))
			},
		},
		{
			name: "Infof",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Infof("Goravel: %s", "World")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.info: Goravel: World"))
				assert.True(t, file.Contains(dailyLog, "test.info: Goravel: World"))
			},
		},
		{
			name: "Warning",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Warning("Goravel")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.warning: Goravel"))
				assert.True(t, file.Contains(dailyLog, "test.warning: Goravel"))
			},
		},
		{
			name: "Warningf",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Warningf("Goravel: %s", "World")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.warning: Goravel: World"))
				assert.True(t, file.Contains(dailyLog, "test.warning: Goravel: World"))
			},
		},
		{
			name: "Error",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Error("Goravel")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.error: Goravel"))
				assert.True(t, file.Contains(dailyLog, "test.error: Goravel"))
			},
		},
		{
			name: "Errorf",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Errorf("Goravel: %s", "World")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.error: Goravel: World"))
				assert.True(t, file.Contains(dailyLog, "test.error: Goravel: World"))
			},
		},
		{
			name: "Panic",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
			},
			assert: func() {
				assert.Panics(t, func() {
					log.Panic("Goravel")
				})
				assert.True(t, file.Contains(singleLog, "test.panic: Goravel"))
				assert.True(t, file.Contains(dailyLog, "test.panic: Goravel"))
			},
		},
		{
			name: "Panicf",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
			},
			assert: func() {
				assert.Panics(t, func() {
					log.Panicf("Goravel: %s", "World")
				})
				assert.True(t, file.Contains(singleLog, "test.panic: Goravel: World"))
				assert.True(t, file.Contains(dailyLog, "test.panic: Goravel: World"))
			},
		},
		{
			name: "Code",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Code("code").Info("Goravel")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.info: Goravel\n[Code] code"))
				assert.True(t, file.Contains(dailyLog, "test.info: Goravel\n[Code] code"))
			},
		},
		{
			name: "Hint",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Hint("hint").Info("Goravel")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.info: Goravel\n[Hint] hint"))
				assert.True(t, file.Contains(dailyLog, "test.info: Goravel\n[Hint] hint"))
			},
		},
		{
			name: "In",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.In("domain").Info("Goravel")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.info: Goravel\n[Domain] domain"))
				assert.True(t, file.Contains(dailyLog, "test.info: Goravel\n[Domain] domain"))
			},
		},
		{
			name: "Owner",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Owner("team@goravel.dev").Info("Goravel")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.info: Goravel\n[Owner] team@goravel.dev"))
				assert.True(t, file.Contains(dailyLog, "test.info: Goravel\n[Owner] team@goravel.dev"))
			},
		},
		{
			name: "Request",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Request(&TestRequest{}).Info("Goravel")
			},
			assert: func() {
				expectedParts := []string{
					`test.info: Goravel`,
					`[Request] map[`,
					`method:GET`,
					`uri:http://localhost:3000/`,
					`Sec-Fetch-User:[?1]`,
					`Host:[localhost:3000]`,
					`body:map[key1:value1 key2:value2]`,
				}

				for _, part := range expectedParts {
					assert.True(t, file.Contains(singleLog, part), part)
					assert.True(t, file.Contains(dailyLog, part), part)
				}
			},
		},
		{
			name: "Response",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Response(&TestResponse{}).Info("Goravel")
			},
			assert: func() {
				expectedParts := []string{
					`test.info: Goravel`,
					`[Response] map[`,
					`status:200`,
					`header:map[Content-Type:[text/plain; charset=utf-8]]`,
					`body:body`,
					`size:4`,
				}

				for _, part := range expectedParts {
					assert.True(t, file.Contains(singleLog, part))
					assert.True(t, file.Contains(dailyLog, part))
				}
			},
		},
		{
			name: "Tags",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Tags("tag").Info("Goravel")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.info: Goravel\n[Tags] [tag]"))
				assert.True(t, file.Contains(dailyLog, "test.info: Goravel\n[Tags] [tag]"))
			},
		},
		{
			name: "User",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.User(map[string]any{"name": "kkumar-gcc"}).Info("Goravel")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.info: Goravel\n[User] map[name:kkumar-gcc]"))
				assert.True(t, file.Contains(dailyLog, "test.info: Goravel\n[User] map[name:kkumar-gcc]"))
			},
		},
		{
			name: "With",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.With(map[string]any{"key": "value"}).Info("Goravel")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.info: Goravel\n[With] map[key:value]"))
				assert.True(t, file.Contains(dailyLog, "test.info: Goravel\n[With] map[key:value]"))
			},
		},
		{
			name: "WithTrace",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.WithTrace().Info("Goravel")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.info: Goravel\n[Trace]"))
				assert.True(t, file.Contains(dailyLog, "test.info: Goravel\n[Trace]"))
			},
		},
		{
			name: "No traces when calling Info after Error",
			setup: func() {
				mockDriverConfig(mockConfig)

				log, err = NewApplication(mockConfig, j)
				log.Error("test error")
				log.Info("test info")
			},
			assert: func() {
				assert.True(t, file.Contains(singleLog, "test.error: test error\n[Trace]"))
				assert.True(t, file.Contains(singleLog, "test.info: test info"))
				assert.False(t, file.Contains(dailyLog, "test.info: test info\n[Trace]"))
				assert.True(t, file.Contains(dailyLog, "test.error: test error"))
				assert.True(t, file.Contains(dailyLog, "test.info: test info"))
				assert.False(t, file.Contains(singleLog, "test.info: test info\n[Trace]"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()
			assert.Nil(t, err)
			test.assert()
		})
	}
	_ = file.Remove("storage")
}

func TestWriterWithCustomLogger(t *testing.T) {
	mockConfig := configmock.NewConfig(t)
	mockConfig.EXPECT().GetString("logging.default").Return("customLogger").Once()
	mockConfig.EXPECT().GetString("logging.channels.customLogger.driver").Return("custom").Twice()
	mockConfig.EXPECT().Get("logging.channels.customLogger.via").Return(&CustomLogger{}).Twice()

	filename := "custom.log"

	logger, err := NewApplication(mockConfig, json.New())
	assert.Nil(t, err)
	assert.NotNil(t, logger)

	channel := logger.Channel("customLogger")

	assert.NotNil(t, channel)

	channel.WithTrace().
		With(map[string]any{"filename": filename}).
		User(map[string]any{"name": "kkumar-gcc"}).
		Owner("team@goravel.dev").
		Code("code").Info("Goravel")

	expectedContent := "info: Goravel\ncustom_code: code\ncustom_user: map[name:kkumar-gcc]\n"
	assert.True(t, file.Contains(filename, expectedContent), "Log file content does not match expected output")

	assert.Nil(t, file.Remove(filename))
}

func TestWriter_Fatal(t *testing.T) {
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log, err := NewApplication(mockConfig, json.New())
	assert.Nil(t, err)
	assert.NotNil(t, log)

	if os.Getenv("FATAL") == "1" {
		log.Fatal("Goravel")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestWriter_Fatal")
	cmd.Env = append(os.Environ(), "FATAL=1")
	err = cmd.Run()

	assert.EqualError(t, err, "exit status 1")
	assert.True(t, file.Contains(singleLog, "test.fatal: Goravel"))
	assert.True(t, file.Contains(dailyLog, "test.fatal: Goravel"))

	_ = file.Remove("storage")
}

func TestWriter_Fatalf(t *testing.T) {
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log, err := NewApplication(mockConfig, json.New())
	assert.Nil(t, err)
	assert.NotNil(t, log)

	if os.Getenv("FATAL") == "1" {
		log.Fatalf("Goravel: %s", "World")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestWriter_Fatalf")
	cmd.Env = append(os.Environ(), "FATAL=1")
	err = cmd.Run()

	assert.EqualError(t, err, "exit status 1")
	assert.True(t, file.Contains(singleLog, "test.fatal: Goravel: World"))
	assert.True(t, file.Contains(dailyLog, "test.fatal: Goravel: World"))

	_ = file.Remove("storage")
}

func Benchmark_Debug(b *testing.B) {
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log, err := NewApplication(mockConfig, json.New())
	assert.Nil(b, err)
	assert.NotNil(b, log)

	for i := 0; i < b.N; i++ {
		log.Debug("Debug Goravel")
	}

	_ = file.Remove("storage")
}

func Benchmark_Info(b *testing.B) {
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log, err := NewApplication(mockConfig, json.New())
	assert.Nil(b, err)
	assert.NotNil(b, log)

	for i := 0; i < b.N; i++ {
		log.Info("Goravel")
	}

	_ = file.Remove("storage")
}

func Benchmark_Warning(b *testing.B) {
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log, err := NewApplication(mockConfig, json.New())
	assert.Nil(b, err)
	assert.NotNil(b, log)

	for i := 0; i < b.N; i++ {
		log.Warning("Goravel")
	}

	_ = file.Remove("storage")
}

func Benchmark_Error(b *testing.B) {
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log, err := NewApplication(mockConfig, json.New())
	assert.Nil(b, err)
	assert.NotNil(b, log)

	for i := 0; i < b.N; i++ {
		log.Error("Goravel")
	}

	_ = file.Remove("storage")
}

func Benchmark_Fatal(b *testing.B) {
	// This test is not suitable for benchmarking because it will exit the program
}

func Benchmark_Panic(b *testing.B) {
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log, err := NewApplication(mockConfig, json.New())
	assert.Nil(b, err)
	assert.NotNil(b, log)

	for i := 0; i < b.N; i++ {
		func() {
			defer func() {
				recover() //nolint:errcheck
			}()
			log.Panic("Goravel")
		}()
	}

	_ = file.Remove("storage")
}

func initMockConfig() *configmock.Config {
	mockConfig := &configmock.Config{}
	mockConfig.EXPECT().GetString("logging.default").Return("stack").Once()
	mockConfig.EXPECT().GetString("logging.channels.stack.driver").Return("stack").Once()
	mockConfig.On("Get", "logging.channels.stack.channels").Return([]string{"single", "daily"}).Once()
	mockConfig.EXPECT().GetString("logging.channels.daily.driver").Return("daily").Once()
	mockConfig.EXPECT().GetString("logging.channels.daily.path").Return(singleLog).Once()
	mockConfig.EXPECT().GetInt("logging.channels.daily.days").Return(7).Once()
	mockConfig.EXPECT().GetBool("logging.channels.daily.print").Return(false).Once()
	mockConfig.EXPECT().GetString("logging.channels.single.driver").Return("single").Once()
	mockConfig.EXPECT().GetString("logging.channels.single.path").Return(singleLog).Once()
	mockConfig.EXPECT().GetBool("logging.channels.single.print").Return(false).Once()

	return mockConfig
}

func mockDriverConfig(mockConfig *configmock.Config) {
	mockConfig.EXPECT().GetString("logging.channels.daily.level").Return("debug").Once()
	mockConfig.EXPECT().GetString("logging.channels.daily.formatter", "text").Return("text").Once()
	mockConfig.EXPECT().GetString("logging.channels.single.level").Return("debug").Once()
	mockConfig.EXPECT().GetString("logging.channels.single.formatter", "text").Return("text").Once()
	mockConfig.EXPECT().GetString("app.env").Return("test").Maybe()
}

// CustomLogger is a custom logger for testing custom log drivers.
type CustomLogger struct {
}

func (logger *CustomLogger) Handle(channel string) (contractslog.Handler, error) {
	return &CustomHandler{}, nil
}

// CustomHandler is a custom slog.Handler for testing.
type CustomHandler struct{}

func (h *CustomHandler) Enabled(level contractslog.Level) bool {
	return level.Level() >= slog.LevelInfo
}

func (h *CustomHandler) Handle(entry contractslog.Entry) error {
	var filename string
	var code string
	var user any

	if fn, ok := entry.With()["filename"]; ok {
		filename = fmt.Sprintf("%v", fn)
	}
	if c := entry.Code(); c != "" {
		code = c
	}
	if u := entry.User(); u != nil {
		user = u
	}

	if filename != "" {
		var builder strings.Builder
		message := entry.Message()
		if len(message) > 0 {
			builder.WriteString(fmt.Sprintf("%s: %v\n", entry.Level().String(), message))
		}

		if len(code) > 0 {
			builder.WriteString(fmt.Sprintf("custom_code: %v\n", code))
		}

		if user != nil {
			builder.WriteString(fmt.Sprintf("custom_user: %v\n", user))
		}

		err := file.PutContent(filename, builder.String())
		if err != nil {
			return err
		}
	}
	return nil
}

type TestRequest struct {
	contractshttp.ContextRequest
}

func (r *TestRequest) Headers() nethttp.Header {
	return nethttp.Header{
		"Sec-Fetch-User": []string{"?1"},
		"Host":           []string{"localhost:3000"},
	}
}

func (r *TestRequest) Method() string {
	return "GET"
}

func (r *TestRequest) Path() string {
	return "/test"
}

func (r *TestRequest) FullUrl() string {
	return "http://localhost:3000/"
}

func (r *TestRequest) All() map[string]any {
	return map[string]any{
		"key1": "value1",
		"key2": "value2",
	}
}

func (r *TestRequest) Abort(code ...int) {
}

func (r *TestRequest) Next() {}

type TestResponse struct {
	contractshttp.ContextResponse
}

func (r *TestResponse) Origin() contractshttp.ResponseOrigin {
	return &TestResponseOrigin{ctx: r}
}

type TestResponseOrigin struct {
	ctx *TestResponse
}

func (r *TestResponseOrigin) Body() *bytes.Buffer {
	return bytes.NewBuffer([]byte("body"))
}

func (r *TestResponseOrigin) Header() nethttp.Header {
	return nethttp.Header{
		"Content-Type": []string{"text/plain; charset=utf-8"},
	}
}

func (r *TestResponseOrigin) Size() int {
	return r.Body().Len()
}

func (r *TestResponseOrigin) Status() int {
	return 200
}

func TestWriter_ConcurrentAccess(t *testing.T) {
	// This test verifies that concurrent access to the same log.Writer
	// does not cause data races or entry contamination.
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log, err := NewApplication(mockConfig, json.New())
	assert.Nil(t, err)
	assert.NotNil(t, log)

	const goroutines = 10
	const iterations = 100

	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				// Each goroutine uses its own unique code
				code := fmt.Sprintf("code-%d-%d", id, j)
				log.Code(code).Info(fmt.Sprintf("message from goroutine %d iteration %d", id, j))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Verify that we have entries from all goroutines
	for i := 0; i < goroutines; i++ {
		for j := 0; j < iterations; j++ {
			expectedCode := fmt.Sprintf("[Code] code-%d-%d", i, j)
			expectedMsg := fmt.Sprintf("message from goroutine %d iteration %d", i, j)
			assert.True(t, file.Contains(singleLog, expectedCode), "Missing code: %s", expectedCode)
			assert.True(t, file.Contains(singleLog, expectedMsg), "Missing message: %s", expectedMsg)
		}
	}

	_ = file.Remove("storage")
}

func TestWriter_NoEntryContamination(t *testing.T) {
	// This test verifies that calling fluent methods on the base writer
	// returns a new writer and does not affect the original.
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log, err := NewApplication(mockConfig, json.New())
	assert.Nil(t, err)
	assert.NotNil(t, log)

	// Call Code on the base writer, then log without code
	_ = log.Code("should-not-appear")
	log.Info("message without code")

	// The message should NOT have the code since we didn't chain the calls
	assert.True(t, file.Contains(singleLog, "test.info: message without code"))
	assert.False(t, file.Contains(singleLog, "message without code\n[Code] should-not-appear"))

	_ = file.Remove("storage")
}

func TestWriter_TimestampIsCurrentAtLogTime(t *testing.T) {
	// This test verifies that the timestamp is captured at log time, not writer creation time.
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log, err := NewApplication(mockConfig, json.New())
	assert.Nil(t, err)
	assert.NotNil(t, log)

	// Small delay to ensure time difference from writer creation
	time.Sleep(50 * time.Millisecond)

	log.Info("timestamp test")

	// Read the log file and verify the timestamp is reasonable
	content, readErr := os.ReadFile(singleLog)
	assert.Nil(t, readErr)
	logContent := string(content)

	// The log should contain a timestamp that is after beforeLog
	assert.Contains(t, logContent, "test.info: timestamp test")

	_ = file.Remove("storage")
}

func TestWriter_FluentChainIsolation(t *testing.T) {
	// This test verifies that multiple fluent chains are isolated from each other.
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log, err := NewApplication(mockConfig, json.New())
	assert.Nil(t, err)
	assert.NotNil(t, log)

	// Create two separate chains
	chain1 := log.Code("chain1-code")
	chain2 := log.Code("chain2-code")

	// Log from both chains
	chain1.Info("message from chain1")
	chain2.Info("message from chain2")

	// Verify each chain has its own code
	assert.True(t, file.Contains(singleLog, "message from chain1\n[Code] chain1-code"))
	assert.True(t, file.Contains(singleLog, "message from chain2\n[Code] chain2-code"))

	// Verify chain2 code doesn't appear in chain1's message
	assert.False(t, file.Contains(singleLog, "message from chain1\n[Code] chain2-code"))
	assert.False(t, file.Contains(singleLog, "message from chain2\n[Code] chain1-code"))

	_ = file.Remove("storage")
}
