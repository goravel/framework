package log

import (
	"bytes"
	"context"
	"fmt"
	nethttp "net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/filesystem"
	contractshttp "github.com/goravel/framework/contracts/http"
	logcontracts "github.com/goravel/framework/contracts/log"
	contractsession "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/foundation/json"
	configmock "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

var singleLog = "storage/logs/goravel.log"
var dailyLog = fmt.Sprintf("storage/logs/goravel-%s.log", carbon.Now().ToDateString())

func TestLogrus(t *testing.T) {
	var (
		mockConfig *configmock.Config
		log        *Application
		j          = json.NewJson()
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
				mockConfig.On("GetString", "logging.channels.daily.level").Return("debug").Once()
				mockConfig.On("GetString", "logging.channels.single.level").Return("debug").Once()

				log, err = NewApplication(mockConfig, j)
			},
			assert: func() {
				writer := log.WithContext(context.Background())
				assert.Equal(t, reflect.TypeOf(writer).String(), reflect.TypeOf(&Writer{}).String())
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
				assert.True(t, file.Contain(singleLog, "test.debug: Debug Goravel"))
				assert.True(t, file.Contain(dailyLog, "test.debug: Debug Goravel"))
			},
		},
		{
			name: "No Debug",
			setup: func() {
				mockConfig.On("GetString", "logging.channels.daily.level").Return("info").Once()
				mockConfig.On("GetString", "logging.channels.single.level").Return("info").Once()
				mockConfig.On("GetString", "app.timezone").Return("UTC").Once()
				mockConfig.On("GetString", "app.env").Return("test").Once()
				log, err = NewApplication(mockConfig, j)
				log.Debug("No Debug Goravel")
			},
			assert: func() {
				assert.False(t, file.Contain(singleLog, "test.debug: No Debug Goravel"))
				assert.False(t, file.Contain(dailyLog, "test.debug: No Debug Goravel"))
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
				assert.True(t, file.Contain(singleLog, "test.debug: Goravel: World"))
				assert.True(t, file.Contain(dailyLog, "test.debug: Goravel: World"))
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
				assert.True(t, file.Contain(singleLog, "test.info: Goravel"))
				assert.True(t, file.Contain(dailyLog, "test.info: Goravel"))
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
				assert.True(t, file.Contain(singleLog, "test.info: Goravel: World"))
				assert.True(t, file.Contain(dailyLog, "test.info: Goravel: World"))
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
				assert.True(t, file.Contain(singleLog, "test.warning: Goravel"))
				assert.True(t, file.Contain(dailyLog, "test.warning: Goravel"))
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
				assert.True(t, file.Contain(singleLog, "test.warning: Goravel: World"))
				assert.True(t, file.Contain(dailyLog, "test.warning: Goravel: World"))
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
				assert.True(t, file.Contain(singleLog, "test.error: Goravel"))
				assert.True(t, file.Contain(dailyLog, "test.error: Goravel"))
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
				assert.True(t, file.Contain(singleLog, "test.error: Goravel: World"))
				assert.True(t, file.Contain(dailyLog, "test.error: Goravel: World"))
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
				assert.True(t, file.Contain(singleLog, "test.panic: Goravel"))
				assert.True(t, file.Contain(dailyLog, "test.panic: Goravel"))
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
				assert.True(t, file.Contain(singleLog, "test.panic: Goravel: World"))
				assert.True(t, file.Contain(dailyLog, "test.panic: Goravel: World"))
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
				assert.True(t, file.Contain(singleLog, "test.info: Goravel\ncode: \"code\""))
				assert.True(t, file.Contain(dailyLog, "test.info: Goravel\ncode: \"code\""))
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
				assert.True(t, file.Contain(singleLog, "test.info: Goravel\nhint: \"hint\""))
				assert.True(t, file.Contain(dailyLog, "test.info: Goravel\nhint: \"hint\""))
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
				assert.True(t, file.Contain(singleLog, "test.info: Goravel\ndomain: \"domain\""))
				assert.True(t, file.Contain(dailyLog, "test.info: Goravel\ndomain: \"domain\""))
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
				assert.True(t, file.Contain(singleLog, "test.info: Goravel\nowner: \"team@goravel.dev\""))
				assert.True(t, file.Contain(dailyLog, "test.info: Goravel\nowner: \"team@goravel.dev\""))
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
					`request: {`,
					`"method":"GET`,
					`"uri":"http://localhost:3000/"`,
					`"Sec-Fetch-User":["?1"]`,
					`"Host":["localhost:3000"]`,
					`"body":{`,
					`"key1":"value1"`,
					`"key2":"value2"`,
				}

				for _, part := range expectedParts {
					assert.True(t, file.Contain(singleLog, part), part)
					assert.True(t, file.Contain(dailyLog, part), part)
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
					`response: {`,
					`"status":200`,
					`"header":{"Content-Type":["text/plain; charset=utf-8"]}`,
					`"body":{}`,
					`"size":4`,
				}

				for _, part := range expectedParts {
					assert.True(t, file.Contain(singleLog, part))
					assert.True(t, file.Contain(dailyLog, part))
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
				assert.True(t, file.Contain(singleLog, "test.info: Goravel\ntags: [\"tag\"]"))
				assert.True(t, file.Contain(dailyLog, "test.info: Goravel\ntags: [\"tag\"]"))
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
				assert.True(t, file.Contain(singleLog, "test.info: Goravel\nuser: {\"name\":\"kkumar-gcc\"}"))
				assert.True(t, file.Contain(dailyLog, "test.info: Goravel\nuser: {\"name\":\"kkumar-gcc\"}"))
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
				assert.True(t, file.Contain(singleLog, "test.info: Goravel\ncontext: {\"key\":\"value\"}"))
				assert.True(t, file.Contain(dailyLog, "test.info: Goravel\ncontext: {\"key\":\"value\"}"))
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
				assert.True(t, file.Contain(singleLog, "test.info: Goravel\ntrace:"))
				assert.True(t, file.Contain(dailyLog, "test.info: Goravel\ntrace:"))
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
				assert.True(t, file.Contain(singleLog, "test.error: test error\ntrace:"))
				assert.True(t, file.Contain(singleLog, "test.info: test info"))
				assert.False(t, file.Contain(dailyLog, "test.info: test info\ntrace:"))
				assert.True(t, file.Contain(dailyLog, "test.error: test error"))
				assert.True(t, file.Contain(dailyLog, "test.info: test info"))
				assert.False(t, file.Contain(singleLog, "test.info: test info\ntrace:"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()
			assert.Nil(t, err)
			test.assert()

			mockConfig.AssertExpectations(t)
		})
	}
	_ = file.Remove("storage")
}

func TestLogrusWithCustomLogger(t *testing.T) {
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "logging.default").Return("customLogger").Once()
	mockConfig.On("GetString", "logging.channels.customLogger.driver").Return("custom").Twice()
	mockConfig.On("Get", "logging.channels.customLogger.via").Return(&CustomLogger{}).Twice()
	mockConfig.On("GetString", "app.timezone").Return("UTC")
	mockConfig.On("GetString", "app.env").Return("test")

	filename := "custom.log"

	logger, err := NewApplication(mockConfig, json.NewJson())
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
	assert.True(t, file.Contain(filename, expectedContent), "Log file content does not match expected output")

	assert.Nil(t, file.Remove(filename))
}

func TestLogrus_Fatal(t *testing.T) {
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log, err := NewApplication(mockConfig, json.NewJson())
	assert.Nil(t, err)
	assert.NotNil(t, log)

	if os.Getenv("FATAL") == "1" {
		log.Fatal("Goravel")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLogrus_Fatal")
	cmd.Env = append(os.Environ(), "FATAL=1")
	err = cmd.Run()

	assert.EqualError(t, err, "exit status 1")
	assert.True(t, file.Contain(singleLog, "test.fatal: Goravel"))
	assert.True(t, file.Contain(dailyLog, "test.fatal: Goravel"))

	_ = file.Remove("storage")
}

func TestLogrus_Fatalf(t *testing.T) {
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log, err := NewApplication(mockConfig, json.NewJson())
	assert.Nil(t, err)
	assert.NotNil(t, log)

	if os.Getenv("FATAL") == "1" {
		log.Fatalf("Goravel")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLogrus_Fatal")
	cmd.Env = append(os.Environ(), "FATAL=1")
	err = cmd.Run()

	assert.EqualError(t, err, "exit status 1")
	assert.True(t, file.Contain(singleLog, "test.fatal: Goravel"))
	assert.True(t, file.Contain(dailyLog, "test.fatal: Goravel"))

	_ = file.Remove("storage")
}

func Benchmark_Debug(b *testing.B) {
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log, err := NewApplication(mockConfig, json.NewJson())
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
	log, err := NewApplication(mockConfig, json.NewJson())
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
	log, err := NewApplication(mockConfig, json.NewJson())
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
	log, err := NewApplication(mockConfig, json.NewJson())
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
	log, err := NewApplication(mockConfig, json.NewJson())
	assert.Nil(b, err)
	assert.NotNil(b, log)

	for i := 0; i < b.N; i++ {
		defer func() {
			recover() //nolint:errcheck
		}()
		log.Panic("Goravel")
	}

	_ = file.Remove("storage")
}

func initMockConfig() *configmock.Config {
	mockConfig := &configmock.Config{}

	mockConfig.On("GetString", "logging.default").Return("stack").Once()
	mockConfig.On("GetString", "logging.channels.stack.driver").Return("stack").Once()
	mockConfig.On("Get", "logging.channels.stack.channels").Return([]string{"single", "daily"}).Once()
	mockConfig.On("GetString", "logging.channels.daily.driver").Return("daily").Once()
	mockConfig.On("GetString", "logging.channels.daily.path").Return(singleLog).Once()
	mockConfig.On("GetInt", "logging.channels.daily.days").Return(7).Once()
	mockConfig.On("GetBool", "logging.channels.daily.print").Return(false).Once()
	mockConfig.On("GetString", "logging.channels.single.driver").Return("single").Once()
	mockConfig.On("GetString", "logging.channels.single.path").Return(singleLog).Once()
	mockConfig.On("GetBool", "logging.channels.single.print").Return(false).Once()

	return mockConfig
}

func mockDriverConfig(mockConfig *configmock.Config) {
	mockConfig.On("GetString", "logging.channels.daily.level").Return("debug").Once()
	mockConfig.On("GetString", "logging.channels.single.level").Return("debug").Once()
	mockConfig.On("GetString", "app.timezone").Return("UTC")
	mockConfig.On("GetString", "app.env").Return("test")
}

type CustomLogger struct {
}

func (logger *CustomLogger) Handle(channel string) (logcontracts.Hook, error) {
	return &CustomHook{}, nil
}

type CustomHook struct {
}

func (h *CustomHook) Levels() []logcontracts.Level {
	return []logcontracts.Level{
		logcontracts.InfoLevel,
	}
}

func (h *CustomHook) Fire(entry logcontracts.Entry) error {
	with := entry.With()
	filename, ok := with["filename"]
	if ok {
		var builder strings.Builder
		message := entry.Message()
		if len(message) > 0 {
			builder.WriteString(fmt.Sprintf("%s: %v\n", entry.Level(), message))
		}

		code := entry.Code()
		if len(code) > 0 {
			builder.WriteString(fmt.Sprintf("custom_code: %v\n", code))
		}

		user := entry.User()
		if user != nil {
			builder.WriteString(fmt.Sprintf("custom_user: %v\n", user))
		}

		err := file.Create(cast.ToString(filename), builder.String())
		if err != nil {
			return err
		}
	}
	return nil
}

type TestRequest struct{}

func (r *TestRequest) Header(key string, defaultValue ...string) string {
	panic("do not need to implement it")
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

func (r *TestRequest) Url() string {
	panic("do not need to implement it")
}

func (r *TestRequest) FullUrl() string {
	return "http://localhost:3000/"
}

func (r *TestRequest) Ip() string {
	panic("do not need to implement it")
}

func (r *TestRequest) Host() string {
	panic("do not need to implement it")
}

func (r *TestRequest) All() map[string]any {
	return map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
}

func (r *TestRequest) Cookie(key string, defaultValue ...string) string {
	panic("do not need to implement it")
}

func (r *TestRequest) Bind(obj any) error {
	panic("do not need to implement it")
}

func (r *TestRequest) BindQuery(any) error {
	panic("do not need to implement it")
}

func (r *TestRequest) Route(key string) string {
	panic("do not need to implement it")
}

func (r *TestRequest) RouteInt(key string) int {
	panic("do not need to implement it")
}

func (r *TestRequest) RouteInt64(key string) int64 {
	panic("do not need to implement it")
}

func (r *TestRequest) Query(key string, defaultValue ...string) string {
	panic("do not need to implement it")
}

func (r *TestRequest) QueryInt(key string, defaultValue ...int) int {
	panic("do not need to implement it")
}

func (r *TestRequest) QueryInt64(key string, defaultValue ...int64) int64 {
	panic("do not need to implement it")
}

func (r *TestRequest) QueryBool(key string, defaultValue ...bool) bool {
	panic("do not need to implement it")
}

func (r *TestRequest) QueryArray(key string) []string {
	panic("do not need to implement it")
}

func (r *TestRequest) QueryMap(key string) map[string]string {
	panic("do not need to implement it")
}

func (r *TestRequest) Queries() map[string]string {
	panic("do not need to implement it")
}

func (r *TestRequest) HasSession() bool {
	panic("do not need to implement it")
}

func (r *TestRequest) SetSession(contractsession.Session) contractshttp.ContextRequest {
	panic("do not need to implement it")
}

func (r *TestRequest) Session() contractsession.Session {
	panic("do not need to implement it")
}

func (r *TestRequest) Input(key string, defaultValue ...string) string {
	panic("do not need to implement it")
}

func (r *TestRequest) InputArray(key string, defaultValue ...[]string) []string {
	panic("do not need to implement it")
}

func (r *TestRequest) InputMap(key string, defaultValue ...map[string]string) map[string]string {
	panic("do not need to implement it")
}

func (r *TestRequest) InputInt(key string, defaultValue ...int) int {
	panic("do not need to implement it")
}

func (r *TestRequest) InputInt64(key string, defaultValue ...int64) int64 {
	panic("do not need to implement it")
}

func (r *TestRequest) InputBool(key string, defaultValue ...bool) bool {
	panic("do not need to implement it")
}

func (r *TestRequest) File(name string) (filesystem.File, error) {
	panic("do not need to implement it")
}

func (r *TestRequest) AbortWithStatus(code int) {}

func (r *TestRequest) AbortWithStatusJson(code int, jsonObj any) {
	panic("do not need to implement it")
}

func (r *TestRequest) Next() {}

func (r *TestRequest) Origin() *nethttp.Request {
	panic("do not need to implement it")
}

func (r *TestRequest) Validate(rules map[string]string, options ...validation.Option) (validation.Validator, error) {
	panic("do not need to implement it")
}

func (r *TestRequest) ValidateRequest(request contractshttp.FormRequest) (validation.Errors, error) {
	panic("do not need to implement it")
}

type TestResponse struct {
}

func (r *TestResponse) Cookie(cookie contractshttp.Cookie) contractshttp.ContextResponse {
	panic("do not need to implement it")
}

func (r *TestResponse) Data(code int, contentType string, data []byte) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) Download(filepath, filename string) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) File(filepath string) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) Header(key, value string) contractshttp.ContextResponse {
	panic("do not need to implement it")
}

func (r *TestResponse) Json(code int, obj any) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) NoContent(...int) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) Origin() contractshttp.ResponseOrigin {
	return &TestResponseOrigin{ctx: r}
}

func (r *TestResponse) Redirect(code int, location string) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) String(code int, format string, values ...any) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) Success() contractshttp.ResponseStatus {
	panic("do not need to implement it")
}

func (r *TestResponse) Status(code int) contractshttp.ResponseStatus {
	panic("do not need to implement it")
}

func (r *TestResponse) Stream(int, func(w contractshttp.StreamWriter) error) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) WithoutCookie(name string) contractshttp.ContextResponse {
	panic("do not need to implement it")
}

func (r *TestResponse) Writer() nethttp.ResponseWriter {
	panic("do not need to implement it")
}

func (r *TestResponse) Flush() {
	panic("do not need to implement it")
}

func (r *TestResponse) View() contractshttp.ResponseView {
	panic("do not need to implement it")
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
