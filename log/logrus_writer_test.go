package log

import (
	"bytes"
	"context"
	"fmt"
	nethttp "net/http"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/filesystem"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

var singleLog = "storage/logs/goravel.log"
var dailyLog = fmt.Sprintf("storage/logs/goravel-%s.log", carbon.Now().ToDateString())

func TestLogrus(t *testing.T) {
	var (
		mockConfig *configmock.Config
		log        *Application
	)

	beforeEach := func() {
		mockConfig = initMockConfig()
	}

	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "WithContext",
			setup: func() {
				mockConfig.On("GetString", "logging.channels.daily.level").Return("debug").Once()
				mockConfig.On("GetString", "logging.channels.single.level").Return("debug").Once()

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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
				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
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

				log = NewApplication(mockConfig)
				log.With(map[string]any{"key": "value"}).Info("Goravel")
			},
			assert: func() {
				assert.True(t, file.Contain(singleLog, "test.info: Goravel\ncontext: {\"key\":\"value\"}"))
				assert.True(t, file.Contain(dailyLog, "test.info: Goravel\ncontext: {\"key\":\"value\"}"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()
			test.assert()

			mockConfig.AssertExpectations(t)
		})
	}

	_ = file.Remove("storage")
}

func TestLogrus_Fatal(t *testing.T) {
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log := NewApplication(mockConfig)

	if os.Getenv("FATAL") == "1" {
		log.Fatal("Goravel")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLogrus_Fatal")
	cmd.Env = append(os.Environ(), "FATAL=1")
	err := cmd.Run()

	assert.EqualError(t, err, "exit status 1")
	assert.True(t, file.Contain(singleLog, "test.fatal: Goravel"))
	assert.True(t, file.Contain(dailyLog, "test.fatal: Goravel"))

	_ = file.Remove("storage")
}

func TestLogrus_Fatalf(t *testing.T) {
	mockConfig := initMockConfig()
	mockDriverConfig(mockConfig)
	log := NewApplication(mockConfig)

	if os.Getenv("FATAL") == "1" {
		log.Fatalf("Goravel")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLogrus_Fatal")
	cmd.Env = append(os.Environ(), "FATAL=1")
	err := cmd.Run()

	assert.EqualError(t, err, "exit status 1")
	assert.True(t, file.Contain(singleLog, "test.fatal: Goravel"))
	assert.True(t, file.Contain(dailyLog, "test.fatal: Goravel"))

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

func (r *TestRequest) Bind(obj any) error {
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

func (r *TestRequest) Form(key string, defaultValue ...string) string {
	panic("do not need to implement it")
}

func (r *TestRequest) Json(key string, defaultValue ...string) string {
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

func (r *TestResponse) Origin() contractshttp.ResponseOrigin {
	return &TestResponseOrigin{ctx: r}
}

func (r *TestResponse) Redirect(code int, location string) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) String(code int, format string, values ...any) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) Success() contractshttp.ResponseSuccess {
	panic("do not need to implement it")
}

func (r *TestResponse) Status(code int) contractshttp.ResponseStatus {
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
