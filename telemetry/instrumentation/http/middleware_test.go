package http

import (
	"bytes"
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	contractshttp "github.com/goravel/framework/contracts/http"
	contractstelemetry "github.com/goravel/framework/contracts/telemetry"
	configmocks "github.com/goravel/framework/mocks/config"
	telemetrymocks "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/telemetry"
)

type MiddlewareTestSuite struct {
	suite.Suite
	originalTelemetry contractstelemetry.Telemetry
}

func (s *MiddlewareTestSuite) SetupTest() {
	s.originalTelemetry = telemetry.TelemetryFacade
}

func (s *MiddlewareTestSuite) TearDownTest() {
	telemetry.TelemetryFacade = s.originalTelemetry
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

func (s *MiddlewareTestSuite) TestTelemetry() {
	defaultTelemetrySetup := func(mockTelemetry *telemetrymocks.Telemetry) {
		mockTelemetry.EXPECT().Tracer(instrumentationName).Return(tracenoop.NewTracerProvider().Tracer("test")).Once()
		mockTelemetry.EXPECT().Meter(instrumentationName).Return(metricnoop.NewMeterProvider().Meter("test")).Once()
		mockTelemetry.EXPECT().Propagator().Return(propagation.NewCompositeTextMapPropagator()).Once()
	}

	tests := []struct {
		name           string
		configSetup    func(*configmocks.Config)
		telemetrySetup func(*telemetrymocks.Telemetry)
		handler        nethttp.HandlerFunc
		requestPath    string
		expectPanic    bool
	}{
		{
			name:        "Success: Request is traced and metrics recorded",
			requestPath: "/users",
			configSetup: func(mockConfig *configmocks.Config) {
				mockConfig.EXPECT().UnmarshalKey("telemetry.instrumentation.http_server", mock.Anything).
					Run(func(_ string, cfg any) {
						c := cfg.(*ServerConfig)
						c.Enabled = true
					}).Return(nil).Once()
			},
			telemetrySetup: defaultTelemetrySetup,
			handler: func(w nethttp.ResponseWriter, r *nethttp.Request) {
				w.WriteHeader(nethttp.StatusOK)
				_, _ = w.Write([]byte("OK"))
			},
		},
		{
			name:        "Ignored: Excluded path is skipped",
			requestPath: "/health",
			configSetup: func(mockConfig *configmocks.Config) {
				mockConfig.EXPECT().UnmarshalKey("telemetry.instrumentation.http_server", mock.Anything).
					Run(func(_ string, cfg interface{}) {
						c := cfg.(*ServerConfig)
						c.Enabled = true
						c.ExcludedPaths = []string{"/health"}
					}).Return(nil).Once()
			},
			telemetrySetup: defaultTelemetrySetup,
			handler: func(w nethttp.ResponseWriter, r *nethttp.Request) {
				w.WriteHeader(nethttp.StatusOK)
			},
		},
		{
			name:        "Ignored: Disabled via config",
			requestPath: "/users",
			configSetup: func(mockConfig *configmocks.Config) {
				mockConfig.EXPECT().UnmarshalKey("telemetry.instrumentation.http_server", mock.Anything).
					Run(func(_ string, cfg interface{}) {
						c := cfg.(*ServerConfig)
						c.Enabled = false
					}).Return(nil).Once()
			},
			telemetrySetup: func(mockTelemetry *telemetrymocks.Telemetry) {
				// If disabled, Tracer/Meter should NOT be initialized
			},
			handler: func(w nethttp.ResponseWriter, r *nethttp.Request) {
				w.WriteHeader(nethttp.StatusOK)
			},
		},
		{
			name:        "Panic: Metrics recorded as 500 and panic re-thrown",
			requestPath: "/crash",
			expectPanic: true,
			configSetup: func(mockConfig *configmocks.Config) {
				mockConfig.EXPECT().UnmarshalKey("telemetry.instrumentation.http_server", mock.Anything).
					Run(func(_ string, cfg any) {
						c := cfg.(*ServerConfig)
						c.Enabled = true
					}).Return(nil).Once()
			},
			telemetrySetup: defaultTelemetrySetup,
			handler: func(w nethttp.ResponseWriter, r *nethttp.Request) {
				panic("server crash")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockConfig := configmocks.NewConfig(s.T())
			mockTelemetry := telemetrymocks.NewTelemetry(s.T())

			telemetry.ConfigFacade = mockConfig
			telemetry.TelemetryFacade = mockTelemetry

			tt.configSetup(mockConfig)
			tt.telemetrySetup(mockTelemetry)

			handler := testMiddleware(tt.handler)
			if tt.expectPanic {
				req := httptest.NewRequest("GET", tt.requestPath, nil)
				w := httptest.NewRecorder()
				s.Panics(func() {
					handler.ServeHTTP(w, req)
				})
			} else {
				server := httptest.NewServer(handler)
				defer server.Close()

				client := &nethttp.Client{}
				resp, err := client.Get(server.URL + tt.requestPath)
				s.NoError(err)
				if resp != nil {
					s.NoError(resp.Body.Close())
				}
			}
		})
	}
}

func testMiddleware(next nethttp.Handler) nethttp.Handler {
	mw := Telemetry()
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		ctx := NewTestContext(r.Context(), next, w, r)
		mw(ctx)
	})
}

type TestContext struct {
	ctx     context.Context
	next    nethttp.Handler
	request *nethttp.Request
	writer  *TestResponseWriter
}

func NewTestContext(ctx context.Context, next nethttp.Handler, w nethttp.ResponseWriter, r *nethttp.Request) *TestContext {
	return &TestContext{
		ctx:     ctx,
		next:    next,
		request: r,
		writer:  &TestResponseWriter{ResponseWriter: w, status: 200},
	}
}

func (c *TestContext) Request() contractshttp.ContextRequest {
	return NewTestRequest(c)
}

func (c *TestContext) Response() contractshttp.ContextResponse {
	return NewTestResponse(c)
}

func (c *TestContext) WithContext(ctx context.Context) {
	c.ctx = ctx
	c.request = c.request.WithContext(ctx)
}

func (c *TestContext) Context() context.Context {
	return c.ctx
}

func (c *TestContext) Err() error {
	return c.ctx.Err()
}

func (c *TestContext) Deadline() (deadline time.Time, ok bool) { return c.ctx.Deadline() }
func (c *TestContext) Done() <-chan struct{}                   { return c.ctx.Done() }
func (c *TestContext) Value(key any) any                       { return c.ctx.Value(key) }
func (c *TestContext) WithValue(key any, value any)            { c.ctx = context.WithValue(c.ctx, key, value) }

type TestRequest struct {
	contractshttp.ContextRequest
	ctx *TestContext
}

func NewTestRequest(ctx *TestContext) *TestRequest {
	return &TestRequest{ctx: ctx}
}

func (r *TestRequest) Next() {
	r.ctx.next.ServeHTTP(r.ctx.writer, r.ctx.request)
}

func (r *TestRequest) Method() string {
	return r.ctx.request.Method
}

func (r *TestRequest) Path() string {
	return r.ctx.request.URL.Path
}

func (r *TestRequest) OriginPath() string {
	return r.ctx.request.URL.Path
}

func (r *TestRequest) Headers() nethttp.Header {
	return r.ctx.request.Header
}

func (r *TestRequest) Header(key string, defaultValue ...string) string {
	val := r.ctx.request.Header.Get(key)
	if val == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return val
}

func (r *TestRequest) Host() string {
	return r.ctx.request.Host
}

func (r *TestRequest) Ip() string {
	return "127.0.0.1"
}

func (r *TestRequest) Origin() *nethttp.Request {
	return r.ctx.request
}

type TestResponse struct {
	contractshttp.ContextResponse
	ctx *TestContext
}

func NewTestResponse(ctx *TestContext) *TestResponse {
	return &TestResponse{ctx: ctx}
}

func (r *TestResponse) Origin() contractshttp.ResponseOrigin {
	return r.ctx.writer
}

type TestResponseWriter struct {
	nethttp.ResponseWriter
	status int
	size   int
}

func (w *TestResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *TestResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

func (w *TestResponseWriter) Status() int {
	return w.status
}

func (w *TestResponseWriter) Size() int {
	return w.size
}

func (w *TestResponseWriter) Header() nethttp.Header {
	return w.ResponseWriter.Header()
}

func (w *TestResponseWriter) Body() *bytes.Buffer {
	return nil
}
