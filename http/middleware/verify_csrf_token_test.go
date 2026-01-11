package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	contractshttp "github.com/goravel/framework/contracts/http"
	contractsession "github.com/goravel/framework/contracts/session"
	mockhttp "github.com/goravel/framework/mocks/http"
	mocksession "github.com/goravel/framework/mocks/session"
)

func TestTokenMatch(t *testing.T) {
	tests := []struct {
		name          string
		hasSession    bool
		sessionToken  string
		headerToken   string
		formToken     string
		expectedMatch bool
	}{
		{
			name:          "no session returns false",
			hasSession:    false,
			expectedMatch: false,
		},
		{
			name:          "valid token in header",
			hasSession:    true,
			sessionToken:  "valid-token",
			headerToken:   "valid-token",
			expectedMatch: true,
		},
		{
			name:          "valid token in form",
			hasSession:    true,
			sessionToken:  "valid-token",
			formToken:     "valid-token",
			expectedMatch: true,
		},
		{
			name:          "invalid token",
			hasSession:    true,
			sessionToken:  "valid-token",
			headerToken:   "invalid-token",
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtx := mockhttp.NewContext(t)
			mockRequest := mockhttp.NewContextRequest(t)
			mockSession := mocksession.NewSession(t)

			mockRequest.EXPECT().HasSession().Return(tt.hasSession).Once()

			if tt.hasSession {
				mockRequest.EXPECT().Session().Return(mockSession).Once()
				mockSession.EXPECT().Token().Return(tt.sessionToken).Once()
				mockRequest.EXPECT().Header(HeaderCsrfKey).Return(tt.headerToken).Once()
				if tt.headerToken == "" {
					mockCtx.EXPECT().Request().Return(mockRequest).Times(4)
					mockRequest.EXPECT().Input("_token").Return(tt.formToken).Once()
				} else {
					mockCtx.EXPECT().Request().Return(mockRequest).Times(3)
				}
			} else {
				mockCtx.EXPECT().Request().Return(mockRequest).Once()
			}
			result := tokenMatch(mockCtx)
			assert.Equal(t, tt.expectedMatch, result)
		})
	}
}

func TestInExceptArray(t *testing.T) {
	tests := []struct {
		name        string
		excepts     []string
		currentPath string
		expected    bool
	}{
		{
			name:        "exact match",
			excepts:     []string{"api/users"},
			currentPath: "api/users",
			expected:    true,
		},
		{
			name:        "wildcard match",
			excepts:     []string{"api/*"},
			currentPath: "api/users",
			expected:    true,
		},
		{
			name:        "no match",
			excepts:     []string{"api/*"},
			currentPath: "web/users",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inExceptArray(tt.excepts, tt.currentPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsReading(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected bool
	}{
		{"GET method", contractshttp.MethodGet, true},
		{"HEAD method", contractshttp.MethodHead, true},
		{"OPTIONS method", contractshttp.MethodOptions, true},
		{"POST method", contractshttp.MethodPost, false},
		{"PUT method", contractshttp.MethodPut, false},
		{"DELETE method", contractshttp.MethodDelete, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isReading(tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseExceptPaths(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []string
		expected []string
	}{
		{
			name:     "simple paths",
			inputs:   []string{"/api/users", "web/posts/"},
			expected: []string{"api/users", "web/posts"},
		},
		{
			name:     "with query parameters",
			inputs:   []string{"/api/users?page=1", "web/posts?sort=desc"},
			expected: []string{"api/users", "web/posts"},
		},
		{
			name:     "with wildcards",
			inputs:   []string{"/api/*", "web/*/comments"},
			expected: []string{"api/*", "web/*/comments"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseExceptPaths(tt.inputs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

type TestContext struct {
	ctx     context.Context
	next    http.Handler
	request *http.Request
	writer  http.ResponseWriter
}

func NewTestContext(ctx context.Context, next http.Handler, w http.ResponseWriter, r *http.Request) *TestContext {
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
	http.SetCookie(r.ctx.writer, &http.Cookie{
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
