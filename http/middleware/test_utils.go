package middleware

import (
	"context"
	"encoding/json"
	"io"
	nethttp "net/http"
	"net/url"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	contractsession "github.com/goravel/framework/contracts/session"
)

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
