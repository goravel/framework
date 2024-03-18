package middleware

import (
	"context"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/goravel/framework/contracts/filesystem"
	contractshttp "github.com/goravel/framework/contracts/http"
	contractsession "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/contracts/validation"
	configmocks "github.com/goravel/framework/mocks/config"
	sessionmocks "github.com/goravel/framework/mocks/session"
	"github.com/goravel/framework/session"
)

func TestStartSession(t *testing.T) {
	var (
		ctx               *TestContext
		mockConfig        *configmocks.Config
		mockSessionFacade *sessionmocks.Manager
	)

	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "Test StartSession",
			setup: func() {
				mockConfig.On("Get", "lottery").Return([]int{1, 2}).Once()
				mockSessionFacade.On("Driver").Return(nil, nil).Once()
				mockSessionFacade.On("BuildSession", nil).Return(nil).Once()
				mockSessionFacade.On("SetID", "").Once()
				mockSessionFacade.On("Start").Once()
				mockSessionFacade.On("GetName").Return("").Once()
				mockSessionFacade.On("GetID").Return("").Once()
				mockSessionFacade.On("Save").Return(nil).Once()
			},
			assert: func() {
				StartSession()(ctx)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx = new(TestContext)
			mockConfig = &configmocks.Config{}
			mockSessionFacade = &sessionmocks.Manager{}
			session.ConfigFacade = mockConfig
			session.Facade = mockSessionFacade
			test.setup()
			test.assert()

			mockConfig.AssertExpectations(t)
			mockSessionFacade.AssertExpectations(t)
		})
	}
}

type TestContext struct {
	response contractshttp.ContextResponse
}

func (r *TestContext) Deadline() (deadline time.Time, ok bool) {
	panic("do not need to implement it")
}

func (r *TestContext) Done() <-chan struct{} {
	panic("do not need to implement it")
}

func (r *TestContext) Err() error {
	panic("do not need to implement it")
}

func (r *TestContext) Value(any) any {
	panic("do not need to implement it")
}

func (r *TestContext) Context() context.Context {
	panic("do not need to implement it")
}

func (r *TestContext) WithValue(string, any) {
	panic("do not need to implement it")
}

func (r *TestContext) Request() contractshttp.ContextRequest {
	return new(TestRequest)
}

func (r *TestContext) Response() contractshttp.ContextResponse {
	if r.response == nil {
		r.response = &TestResponse{
			Headers: make(map[string]string),
		}
	}

	return r.response
}

type TestRequest struct {
	ctx *TestContext
}

func (r *TestRequest) Header(string, ...string) string {
	panic("do not need to implement it")
}

func (r *TestRequest) Headers() nethttp.Header {
	panic("do not need to implement it")
}

func (r *TestRequest) Method() string {

	panic("do not need to implement it")
}

func (r *TestRequest) Path() string {
	return "/test"
}

func (r *TestRequest) Url() string {

	panic("do not need to implement it")
}

func (r *TestRequest) FullUrl() string {
	panic("do not need to implement it")
}

func (r *TestRequest) Ip() string {
	return "127.0.0.1"
}

func (r *TestRequest) Host() string {
	panic("do not need to implement it")
}

func (r *TestRequest) All() map[string]any {
	panic("do not need to implement it")
}

func (r *TestRequest) Cookie(string, ...string) string {
	panic("do not need to implement it")
}

func (r *TestRequest) Bind(any) error {
	panic("do not need to implement it")
}

func (r *TestRequest) Route(string) string {
	panic("do not need to implement it")
}

func (r *TestRequest) RouteInt(string) int {
	panic("do not need to implement it")
}

func (r *TestRequest) RouteInt64(string) int64 {
	panic("do not need to implement it")
}

func (r *TestRequest) Query(string, ...string) string {
	panic("do not need to implement it")
}

func (r *TestRequest) QueryInt(string, ...int) int {
	panic("do not need to implement it")
}

func (r *TestRequest) QueryInt64(string, ...int64) int64 {
	panic("do not need to implement it")
}

func (r *TestRequest) QueryBool(string, ...bool) bool {
	panic("do not need to implement it")
}

func (r *TestRequest) QueryArray(string) []string {
	panic("do not need to implement it")
}

func (r *TestRequest) QueryMap(string) map[string]string {
	panic("do not need to implement it")
}

func (r *TestRequest) Queries() map[string]string {
	panic("do not need to implement it")
}

func (r *TestRequest) HasSession() bool {
	panic("do not need to implement it")
}

func (r *TestRequest) Session() contractsession.Session {
	panic("do not need to implement it")
}

func (r *TestRequest) SetSession(contractsession.Session) contractshttp.ContextRequest {
	panic("do not need to implement it")
}

func (r *TestRequest) Input(string, ...string) string {
	panic("do not need to implement it")
}

func (r *TestRequest) InputArray(string, ...[]string) []string {
	panic("do not need to implement it")
}

func (r *TestRequest) InputMap(string, ...map[string]string) map[string]string {
	panic("do not need to implement it")
}

func (r *TestRequest) InputInt(string, ...int) int {
	panic("do not need to implement it")
}

func (r *TestRequest) InputInt64(string, ...int64) int64 {
	panic("do not need to implement it")
}

func (r *TestRequest) InputBool(string, ...bool) bool {
	panic("do not need to implement it")
}

func (r *TestRequest) File(string) (filesystem.File, error) {
	panic("do not need to implement it")
}

func (r *TestRequest) AbortWithStatus(int) {}

func (r *TestRequest) AbortWithStatusJson(int, any) {}

func (r *TestRequest) Next() {}

func (r *TestRequest) Origin() *nethttp.Request {
	panic("do not need to implement it")
}

func (r *TestRequest) Validate(map[string]string, ...validation.Option) (validation.Validator, error) {
	panic("do not need to implement it")
}

func (r *TestRequest) ValidateRequest(contractshttp.FormRequest) (validation.Errors, error) {
	panic("do not need to implement it")
}

type TestResponse struct {
	Headers map[string]string
}

func (r *TestResponse) Cookie(cookie contractshttp.Cookie) contractshttp.ContextResponse {
	panic("do not need to implement it")
}

func (r *TestResponse) Data(int, string, []byte) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) Download(string, string) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) File(string) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) Header(key, value string) contractshttp.ContextResponse {
	r.Headers[key] = value

	return r
}

func (r *TestResponse) Json(int, any) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) Origin() contractshttp.ResponseOrigin {
	panic("do not need to implement it")
}

func (r *TestResponse) Redirect(int, string) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) String(int, string, ...any) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) Success() contractshttp.ResponseSuccess {
	panic("do not need to implement it")
}

func (r *TestResponse) Status(int) contractshttp.ResponseStatus {
	panic("do not need to implement it")
}

func (r *TestResponse) WithoutCookie(string) contractshttp.ContextResponse {
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
