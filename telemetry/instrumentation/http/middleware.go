package http

import (
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/telemetry"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

const (
	instrumentationName = "github.com/goravel/framework/telemetry/instrumentation/http"

	metricRequestDuration  = "http.server.request.duration"
	metricRequestBodySize  = "http.server.request.body.size"
	metricResponseBodySize = "http.server.response.body.size"

	unitSeconds = "s"
	unitBytes   = "By"
)

func Telemetry(opts ...Option) http.Middleware {
	if telemetry.TelemetryFacade == nil {
		color.Warningln("[Telemetry] Facade not initialized. HTTP middleware disabled.")
		return func(ctx http.Context) { ctx.Request().Next() }
	}

	var cfg ServerConfig
	_ = telemetry.ConfigFacade.UnmarshalKey("telemetry.instrumentation.http_server", &cfg)

	for _, opt := range opts {
		opt(&cfg)
	}

	if !cfg.Enabled {
		return func(ctx http.Context) { ctx.Request().Next() }
	}

	if cfg.SpanNameFormatter == nil {
		cfg.SpanNameFormatter = defaultSpanNameFormatter
	}

	tracer := telemetry.TelemetryFacade.Tracer(instrumentationName)
	meter := telemetry.TelemetryFacade.Meter(instrumentationName)
	propagator := telemetry.TelemetryFacade.Propagator()

	durationHist, _ := meter.Float64Histogram(metricRequestDuration, metric.WithUnit(unitSeconds), metric.WithDescription("Duration of HTTP server requests"))
	requestSizeHist, _ := meter.Int64Histogram(metricRequestBodySize, metric.WithUnit(unitBytes), metric.WithDescription("Size of HTTP server request body"))
	responseSizeHist, _ := meter.Int64Histogram(metricResponseBodySize, metric.WithUnit(unitBytes), metric.WithDescription("Size of HTTP server response body"))

	excludedPaths := make(map[string]bool, len(cfg.ExcludedPaths))
	for _, p := range cfg.ExcludedPaths {
		excludedPaths[p] = true
	}
	excludedMethods := make(map[string]bool, len(cfg.ExcludedMethods))
	for _, m := range cfg.ExcludedMethods {
		excludedMethods[m] = true
	}

	return func(ctx http.Context) {
		req := ctx.Request()

		routePath := req.OriginPath()
		if routePath == "" {
			routePath = req.Path()
		}

		if excludedPaths[routePath] || excludedMethods[req.Method()] {
			req.Next()
			return
		}

		for _, f := range cfg.Filters {
			if !f(ctx) {
				req.Next()
				return
			}
		}

		start := time.Now()
		parentCtx := propagator.Extract(ctx.Context(), propagation.HeaderCarrier(req.Headers()))
		spanName := cfg.SpanNameFormatter(routePath, ctx)

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

		baseAttrs = append(baseAttrs, cfg.MetricAttributes...)

		spanCtx, span := tracer.Start(parentCtx, spanName,
			telemetry.WithAttributes(baseAttrs...),
			telemetry.WithSpanKind(telemetry.SpanKindServer),
		)

		ctx.WithContext(spanCtx)

		func() {
			defer func() {
				if r := recover(); r != nil {
					err := fmt.Errorf("panic: %v", r)
					span.RecordError(err)
					span.SetStatus(codes.Error, "Internal Server Error")

					metricAttrs := metric.WithAttributes(append(baseAttrs, semconv.HTTPResponseStatusCode(500))...)

					durationHist.Record(spanCtx, time.Since(start).Seconds(), metricAttrs)
					requestSizeHist.Record(spanCtx, getRequestSize(req), metricAttrs)
					responseSizeHist.Record(spanCtx, 0, metricAttrs)

					span.End()
					panic(r)
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

		durationHist.Record(spanCtx, time.Since(start).Seconds(), metricAttrs)
		requestSizeHist.Record(spanCtx, getRequestSize(req), metricAttrs)
		responseSizeHist.Record(spanCtx, int64(ctx.Response().Origin().Size()), metricAttrs)
	}
}

func getRequestSize(req http.ContextRequest) int64 {
	size := req.Origin().ContentLength
	if size < 0 {
		return 0
	}
	return size
}
