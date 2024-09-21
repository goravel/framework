package middleware

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/contracts/filesystem"
	contractshttp "github.com/goravel/framework/contracts/http"
	contractsession "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/foundation/json"
	configmocks "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/session"
	"github.com/goravel/framework/support/file"
)

func testHttpSessionMiddleware(next nethttp.Handler, mockConfig *configmocks.Config) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		mockConfigFacade(mockConfig)
		StartSession()(NewTestContext(r.Context(), next, w, r))
	})
}

func mockConfigFacade(mockConfig *configmocks.Config) {
	mockConfig.On("GetString", "session.driver").Return("file").Twice()
	mockConfig.On("GetInt", "session.lifetime").Return(60).Once()
	mockConfig.On("GetString", "session.cookie").Return("goravel_session").Once()
	mockConfig.On("GetString", "session.path").Return("/").Once()
	mockConfig.On("GetString", "session.domain").Return("").Once()
	mockConfig.On("GetBool", "session.secure").Return(false).Once()
	mockConfig.On("GetBool", "session.http_only").Return(true).Once()
	mockConfig.On("GetString", "session.same_site").Return("").Once()
}

func TestStartSession(t *testing.T) {
	mockConfig := &configmocks.Config{}
	session.ConfigFacade = mockConfig
	mockConfig.On("GetInt", "session.lifetime").Return(120).Once()
	mockConfig.On("GetInt", "session.gc_interval", 30).Return(30).Once()
	mockConfig.On("GetString", "session.files").Return("storage/framework/sessions").Once()
	session.SessionFacade = session.NewManager(mockConfig, json.NewJson())
	server := httptest.NewServer(testHttpSessionMiddleware(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		switch r.URL.Path {
		case "/add":
			s := r.Context().Value("session").(contractsession.Session)
			s.Put("foo", "bar").Flash("baz", "qux")
			//nolint:all
			r.WithContext(context.WithValue(r.Context(), "session", s))
		case "/get":
			s := r.Context().Value("session").(contractsession.Session)
			assert.Equal(t, "bar", s.Get("foo"))
			assert.Equal(t, "qux", s.Get("baz"))
		}
	}), mockConfig))
	defer server.Close()

	client := &nethttp.Client{}

	resp, err := client.Get(server.URL + "/add")
	require.NoError(t, err)
	cookie := resp.Cookies()[0]
	assert.Equal(t, "goravel_session", cookie.Name)

	req, err := nethttp.NewRequest("GET", server.URL+"/get", nil)
	require.NoError(t, err)
	req.Header.Set("Cookie", cookie.String())

	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, cookie.Name, resp.Cookies()[0].Name)
	assert.Equal(t, cookie.Value, resp.Cookies()[0].Value)

	assert.NoError(t, file.Remove("storage"))
	mockConfig.AssertExpectations(t)
}

type TestContext struct {
	ctx     context.Context
	next    nethttp.Handler
	request *nethttp.Request
	writer  nethttp.ResponseWriter
}

func NewTestContext(ctx context.Context, next nethttp.Handler, w nethttp.ResponseWriter, r *nethttp.Request) *TestContext {
	return &TestContext{
		ctx:     ctx,
		next:    next,
		request: r,
		writer:  w,
	}
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

func (r *TestContext) Value(key any) any {
	return r.ctx.Value(key)
}

func (r *TestContext) Context() context.Context {
	return r.ctx
}

func (r *TestContext) WithValue(key any, value any) {
	r.ctx = context.WithValue(r.ctx, key, value)
}

func (r *TestContext) Request() contractshttp.ContextRequest {
	return NewTestRequest(r)
}

func (r *TestContext) Response() contractshttp.ContextResponse {
	return NewTestResponse(r)
}

type TestRequest struct {
	ctx *TestContext
}

func NewTestRequest(ctx *TestContext) *TestRequest {
	return &TestRequest{
		ctx: ctx,
	}
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

func (r *TestRequest) Cookie(key string, defaultValue ...string) string {
	cookie, err := r.ctx.request.Cookie(key)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}

		return ""
	}

	val, _ := url.QueryUnescape(cookie.Value)
	return val
}

func (r *TestRequest) Bind(any) error {
	panic("do not need to implement it")
}

func (r *TestRequest) BindQuery(any) error {
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
	_, ok := r.ctx.Value("session").(contractsession.Session)
	return ok
}

func (r *TestRequest) Session() contractsession.Session {
	s, ok := r.ctx.Value("session").(contractsession.Session)
	if !ok {
		return nil
	}
	return s
}

func (r *TestRequest) SetSession(session contractsession.Session) contractshttp.ContextRequest {
	r.ctx.WithValue("session", session)
	r.ctx.request = r.ctx.request.WithContext(r.ctx.Context())
	return r
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

func (r *TestRequest) AbortWithStatusJson(int, any) {
	panic("do not need to implement it")
}

func (r *TestRequest) Next() {
	r.ctx.next.ServeHTTP(r.ctx.writer, r.ctx.request)
}

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
	ctx *TestContext
}

func NewTestResponse(ctx *TestContext) *TestResponse {
	return &TestResponse{
		ctx: ctx,
	}
}

func (r *TestResponse) Cookie(cookie contractshttp.Cookie) contractshttp.ContextResponse {
	path := cookie.Path
	if path == "" {
		path = "/"
	}
	nethttp.SetCookie(r.ctx.writer, &nethttp.Cookie{
		Name:     cookie.Name,
		Value:    url.QueryEscape(cookie.Value),
		MaxAge:   cookie.MaxAge,
		Path:     path,
		Domain:   cookie.Domain,
		Secure:   cookie.Secure,
		HttpOnly: cookie.HttpOnly,
	})

	return r
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

func (r *TestResponse) Header(string, string) contractshttp.ContextResponse {
	return r
}

func (r *TestResponse) Json(int, any) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) NoContent(...int) contractshttp.Response {
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

func (r *TestResponse) Success() contractshttp.ResponseStatus {
	panic("do not need to implement it")
}

func (r *TestResponse) Status(int) contractshttp.ResponseStatus {
	panic("do not need to implement it")
}

func (r *TestResponse) Stream(int, func(contractshttp.StreamWriter) error) contractshttp.Response {
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
