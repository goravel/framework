package http

import (
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/telemetry"
)

const (
	instrumentationName = "github.com/goravel/framework/telemetry/instrumentation/http"

	metricRequestDuration  = "http.server.request.duration"
	metricRequestBodySize  = "http.server.request.body.size"
	metricResponseBodySize = "http.server.response.body.size"

	unitSeconds = "s"
	unitBytes   = "By"
)

// Telemetry creates HTTP server telemetry middleware that instruments incoming
// requests with tracing and metrics. The optional opts parameters allow
// customizing the server configuration (such as span naming and enabling or
// disabling instrumentation). It returns an http.Middleware that propagates
// context, records spans and metrics when telemetry is enabled, and otherwise
// transparently passes requests through when telemetry is disabled or not
// initialized.
func Telemetry(opts ...Option) http.Middleware {
	h := &MiddlewareHandler{
		opts: opts,
	}

	return h.Handle
}

type MiddlewareHandler struct {
	opts     []Option
	once     sync.Once
	disabled bool

	// Telemetry components
	tracer           trace.Tracer
	propagator       propagation.TextMapPropagator
	durationHist     metric.Float64Histogram
	requestSizeHist  metric.Int64Histogram
	responseSizeHist metric.Int64Histogram

	cfg             ServerConfig
	excludedPaths   map[string]bool
	excludedMethods map[string]bool
}

func (r *MiddlewareHandler) Handle(ctx http.Context) {
	r.once.Do(r.init)

	if r.disabled {
		ctx.Request().Next()
		return
	}

	req := ctx.Request()

	routePath := req.OriginPath()
	if routePath == "" {
		routePath = req.Path()
	}

	if r.excludedPaths[routePath] || r.excludedMethods[req.Method()] {
		req.Next()
		return
	}

	for _, f := range r.cfg.Filters {
		if !f(ctx) {
			req.Next()
			return
		}
	}

	start := time.Now()
	parentCtx := r.propagator.Extract(ctx.Context(), propagation.HeaderCarrier(req.Headers()))
	spanName := r.cfg.SpanNameFormatter(routePath, ctx)

	scheme := "http"
	if proto := req.Header("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}

	baseAttrs := []telemetry.KeyValue{
		semconv.HTTPRequestMethodKey.String(req.Method()),
		semconv.HTTPRoute(routePath),
		semconv.URLScheme(scheme),
		semconv.ServerAddress(req.Host()),
		semconv.ClientAddress(req.Ip()),
		semconv.UserAgentOriginal(req.Header("User-Agent")),
	}

	baseAttrs = append(baseAttrs, r.cfg.MetricAttributes...)

	spanCtx, span := r.tracer.Start(parentCtx, spanName,
		telemetry.WithAttributes(baseAttrs...),
		telemetry.WithSpanKind(telemetry.SpanKindServer),
	)

	ctx.WithContext(spanCtx)

	func() {
		defer func() {
			if rec := recover(); rec != nil {
				err := fmt.Errorf("panic: %v", rec)
				span.RecordError(err)
				span.SetStatus(codes.Error, "Internal Server Error")

				metricAttrs := metric.WithAttributes(append(baseAttrs, semconv.HTTPResponseStatusCode(500))...)

				r.durationHist.Record(spanCtx, time.Since(start).Seconds(), metricAttrs)
				r.requestSizeHist.Record(spanCtx, getRequestSize(req), metricAttrs)
				r.responseSizeHist.Record(spanCtx, 0, metricAttrs)

				span.End()
				panic(rec)
			}
		}()
		req.Next()
	}()

	status := ctx.Response().Origin().Status()

	span.SetAttributes(semconv.HTTPResponseStatusCode(status))

	if status >= 500 {
		span.SetStatus(codes.Error, "")
	} else {
		span.SetStatus(codes.Ok, "")
	}

	if err := ctx.Err(); err != nil {
		span.RecordError(err)
	}

	span.End()

	metricAttrs := metric.WithAttributes(append(baseAttrs, semconv.HTTPResponseStatusCode(status))...)

	r.durationHist.Record(spanCtx, time.Since(start).Seconds(), metricAttrs)
	r.requestSizeHist.Record(spanCtx, getRequestSize(req), metricAttrs)
	r.responseSizeHist.Record(spanCtx, int64(ctx.Response().Origin().Size()), metricAttrs)
}

func (r *MiddlewareHandler) init() {
	if telemetry.TelemetryFacade == nil || telemetry.ConfigFacade == nil {
		color.Warningln("[Telemetry] Facades not initialized. HTTP middleware disabled.")
		r.disabled = true
		return
	}

	if err := telemetry.ConfigFacade.UnmarshalKey("telemetry.instrumentation.http_server", &r.cfg); err != nil {
		color.Errorf("[Telemetry] Failed to load HTTP server config: %v. HTTP middleware disabled.", err)
		r.disabled = true
		return
	}

	for _, opt := range r.opts {
		opt(&r.cfg)
	}

	if !r.cfg.Enabled {
		r.disabled = true
		return
	}

	if r.cfg.SpanNameFormatter == nil {
		r.cfg.SpanNameFormatter = defaultSpanNameFormatter
	}

	r.tracer = telemetry.TelemetryFacade.Tracer(instrumentationName)
	r.propagator = telemetry.TelemetryFacade.Propagator()
	meter := telemetry.TelemetryFacade.Meter(instrumentationName)

	r.durationHist, _ = meter.Float64Histogram(metricRequestDuration, metric.WithUnit(unitSeconds), metric.WithDescription("Duration of HTTP server requests"))
	r.requestSizeHist, _ = meter.Int64Histogram(metricRequestBodySize, metric.WithUnit(unitBytes), metric.WithDescription("Size of HTTP server request body"))
	r.responseSizeHist, _ = meter.Int64Histogram(metricResponseBodySize, metric.WithUnit(unitBytes), metric.WithDescription("Size of HTTP server response body"))

	r.excludedPaths = make(map[string]bool, len(r.cfg.ExcludedPaths))
	for _, p := range r.cfg.ExcludedPaths {
		r.excludedPaths[p] = true
	}
	r.excludedMethods = make(map[string]bool, len(r.cfg.ExcludedMethods))
	for _, m := range r.cfg.ExcludedMethods {
		r.excludedMethods[m] = true
	}
}

func getRequestSize(req http.ContextRequest) int64 {
	size := req.Origin().ContentLength
	if size < 0 {
		return 0
	}
	return size
}
