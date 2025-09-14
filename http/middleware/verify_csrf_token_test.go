package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractshttp "github.com/goravel/framework/contracts/http"
	contractsession "github.com/goravel/framework/contracts/session"
	foundationJson "github.com/goravel/framework/foundation/json"
	configmocks "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/session"
	sessionMiddleware "github.com/goravel/framework/session/middleware"
	"github.com/goravel/framework/support/path"
)

func testHttpSessionMiddleware(next nethttp.Handler, mockConfig *configmocks.Config) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		csrfMiddlewareHandler := func() {}
		isCSRFMiddlewareCalled := false
		ctx := NewTestContext(context.Background(), nethttp.HandlerFunc(func(w1 nethttp.ResponseWriter, r1 *nethttp.Request) {
			csrfMiddlewareHandler()
		}), w, r)
		csrfMiddleware := VerifyCsrfToken([]string{
			"unprotected",
			"unprotectedNested/*",
		})
		csrfMiddlewareHandler = func() {
			if !isCSRFMiddlewareCalled {
				isCSRFMiddlewareCalled = true
				csrfMiddleware(ctx)
			}
		}

		mockConfigFacade(mockConfig)
		sessionMiddleware.StartSession()(ctx)

	})
}

func mockConfigFacade(mockConfig *configmocks.Config) {
	mockConfig.EXPECT().GetString("session.default").Return("file").Once()
	mockConfig.EXPECT().GetInt("session.lifetime", 120).Return(120).Once()
	mockConfig.EXPECT().GetString("session.path").Return("/").Once()
	mockConfig.EXPECT().GetString("session.domain").Return("").Once()
	mockConfig.EXPECT().GetBool("session.secure").Return(false).Once()
	mockConfig.EXPECT().GetBool("session.http_only").Return(true).Once()
	mockConfig.EXPECT().GetString("session.same_site").Return("").Once()
}

func TestVerifyCsrfToken(t *testing.T) {
	mockConfig := configmocks.NewConfig(t)
	session.ConfigFacade = mockConfig

	// Setup all required mock expectations with EXPECT()
	mockConfig.EXPECT().GetString("session.default", "file").Return("file").Once()
	mockConfig.EXPECT().GetString("session.drivers.file.driver").Return("file").Once()
	mockConfig.EXPECT().GetInt("session.lifetime", 120).Return(120).Once()
	mockConfig.EXPECT().GetInt("session.gc_interval", 30).Return(30).Once()
	mockConfig.EXPECT().GetString("session.files").Return(path.Storage("framework/sessions")).Once()
	mockConfig.EXPECT().GetString("session.cookie").Return("goravel_session").Once()

	session.SessionFacade = session.NewManager(mockConfig, foundationJson.New())

	handler := nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {})

	server := httptest.NewServer(testHttpSessionMiddleware(handler, mockConfig))
	defer server.Close()

	client := &nethttp.Client{}

	unProtectedMethods := []string{contractshttp.MethodGet, contractshttp.MethodHead, contractshttp.MethodOptions}
	for _, method := range unProtectedMethods {
		req, err := nethttp.NewRequest(method, server.URL+"/unprotected", nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Header.Get(HeaderCsrfKey))
		assert.Equal(t, contractshttp.StatusOK, resp.StatusCode)
	}

	csrfToken := ""
	sessionCookie := nethttp.Cookie{}
	unProtectedNestedRoutes := []string{"/unprotected", "/unprotectedNested/nested/", "/unprotectedNested/nested2"}
	for _, route := range unProtectedNestedRoutes {
		req, err := nethttp.NewRequest(contractshttp.MethodPost, server.URL+route, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		assert.NoError(t, err)
		csrfToken = resp.Header.Get(HeaderCsrfKey)
		sessionCookie = *resp.Cookies()[0]
		assert.Equal(t, contractshttp.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, csrfToken)
	}

	protectedPaths := []string{"/protected", "/unprotectedNested/nested/2", "/protectedNested/nested/", "/protectedNested/nested2", "/unprotectedNested"}
	for _, path := range protectedPaths {
		req, err := nethttp.NewRequest(contractshttp.MethodPost, server.URL+path, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, contractshttp.StatusTokenMismatch, resp.StatusCode)
		assert.Empty(t, resp.Header.Get(HeaderCsrfKey))
	}

	for _, path := range protectedPaths {
		req, err := nethttp.NewRequest(contractshttp.MethodPost, server.URL+path, nil)
		require.NoError(t, err)
		req.Header.Add(HeaderCsrfKey, csrfToken)
		req.Header.Set("Cookie", sessionCookie.String())
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, contractshttp.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, resp.Header.Get(HeaderCsrfKey))
	}

	for _, path := range protectedPaths {
		body := struct {
			Token string `json:"_token"`
		}{
			Token: csrfToken,
		}
		bodyData, err := json.Marshal(body)
		assert.NoError(t, err)
		req, err := nethttp.NewRequest(contractshttp.MethodPost, server.URL+path, bytes.NewBuffer(bodyData))
		require.NoError(t, err)
		req.Header.Set("Cookie", sessionCookie.String())
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, contractshttp.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, resp.Header.Get(HeaderCsrfKey))
	}
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

func (r *TestContext) WithContext(context.Context) {
	panic("do not need to implement it")
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
	contractshttp.ContextRequest
	ctx *TestContext
}

func NewTestRequest(ctx *TestContext) *TestRequest {
	return &TestRequest{
		ctx: ctx,
	}
}

func (r *TestRequest) Path() string {
	if r.ctx != nil && r.ctx.request != nil {
		return r.ctx.request.URL.Path
	}
	return ""
}

func (r *TestRequest) Ip() string {
	return "127.0.0.1"
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

func (r *TestRequest) HasSession() bool {
	if r.ctx == nil {
		return false
	}
	session := r.Session()
	return session != nil
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

func (r *TestRequest) Abort(code ...int) {
	r.ctx.writer.WriteHeader(code[0])
}

func (r *TestRequest) Next() {
	if r.ctx != nil && r.ctx.next != nil {
		r.ctx.next.ServeHTTP(r.ctx.writer, r.ctx.request.WithContext(r.ctx.Context()))
	}
}

func (r *TestRequest) Method() string {
	return r.ctx.request.Method
}

func (r *TestRequest) Header(key string, defaultValue ...string) string {
	headerValue := r.ctx.request.Header.Get(key)
	if len(headerValue) > 0 {
		return headerValue
	} else if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

func (r *TestRequest) Input(key string, defualtVaue ...string) string {
	if body, err := io.ReadAll(r.ctx.request.Body); err != nil {
		if len(defualtVaue) > 0 {
			return defualtVaue[0]
		}
		return ""
	} else {
		data := map[string]any{}
		if err := json.Unmarshal(body, &data); err != nil {
			return ""
		}
		if data[key] == nil {
			return ""
		}
		return data[key].(string)
	}

}

type TestResponse struct {
	contractshttp.ContextResponse
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

func (r *TestResponse) Header(key string, value string) contractshttp.ContextResponse {
	r.ctx.writer.Header().Set(key, value)
	return r
}
