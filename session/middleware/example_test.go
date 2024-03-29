package middleware

import (
	"context"
	"fmt"
	"io"
	nethttp "net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/contracts/filesystem"
	contractshttp "github.com/goravel/framework/contracts/http"
	contractsession "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/contracts/validation"
)

func StartSession1() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		fmt.Println("start")
		ctx.Request().Next()
		fmt.Println("end")
	}
}

func testHttpSessionMiddleware(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		// TODO change StartSession1 to StartSession, instantiate the Facade of service_provider.go and create a ConfigFacade mock
		StartSession1()(NewTestContext1(context.Background(), next, w, r))
	})
}

func TestStartSession1(t *testing.T) {
	go func() {
		nethttp.Handle("/test", testHttpSessionMiddleware(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			fmt.Println("processing")

			w.Write([]byte("Hello, World!"))
		})))
		nethttp.ListenAndServe(":8080", nil)

		select {}
	}()

	time.Sleep(1 * time.Second)

	req, err := nethttp.NewRequest("GET", "http://127.0.0.1:8080/test", nil)
	require.NoError(t, err)

	client := &nethttp.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	fmt.Println("Response:", string(body))
}

type TestContext1 struct {
	ctx      context.Context
	next     nethttp.Handler
	request  *nethttp.Request
	response contractshttp.ContextResponse
	writer   nethttp.ResponseWriter
}

func NewTestContext1(ctx context.Context, next nethttp.Handler, w nethttp.ResponseWriter, r *nethttp.Request) *TestContext1 {
	return &TestContext1{
		ctx:     ctx,
		next:    next,
		request: r,
		writer:  w,
	}
}

func (r *TestContext1) Deadline() (deadline time.Time, ok bool) {

	panic("do not need to implement it")
}

func (r *TestContext1) Done() <-chan struct{} {
	panic("do not need to implement it")
}

func (r *TestContext1) Err() error {
	panic("do not need to implement it")
}

func (r *TestContext1) Value(key any) any {
	return r.ctx.Value(key)
}

func (r *TestContext1) Context() context.Context {
	panic("do not need to implement it")
}

func (r *TestContext1) WithValue(key string, value any) {
	r.ctx = context.WithValue(r.ctx, key, value)
}

func (r *TestContext1) Request() contractshttp.ContextRequest {
	return NewTestRequest(r)
}

func (r *TestContext1) Response() contractshttp.ContextResponse {
	return NewTestResponse(r)
}

type TestRequest struct {
	ctx *TestContext1
}

func NewTestRequest(ctx *TestContext1) *TestRequest {
	return &TestRequest{
		ctx: ctx,
	}
}

func (r *TestRequest) Header(key string, defaultValue ...string) string {
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

	return r
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

func (r *TestRequest) Next() {
	r.ctx.next.ServeHTTP(r.ctx.writer, r.ctx.request)
}

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
	ctx *TestContext1
}

func NewTestResponse(ctx *TestContext1) *TestResponse {
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
	return r
}

func (r *TestResponse) Json(code int, obj any) contractshttp.Response {
	panic("do not need to implement it")
}

func (r *TestResponse) Origin() contractshttp.ResponseOrigin {
	panic("do not need to implement it")
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
