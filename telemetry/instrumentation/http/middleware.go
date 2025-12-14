package http

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/telemetry"
)

const defaultInstrumentationName = "github.com/goravel/framework/telemetry/instrumentation/http"

func Telemetry(opts ...Option) http.Middleware {
	var cfg ServerConfig
	_ = telemetry.ConfigFacade.UnmarshalKey("telemetry.instrumentation.http_server", &cfg)
	for _, opt := range opts {
		opt(&cfg)
	}

	if !cfg.Enabled {
		return noopMiddleware()
	}

	if cfg.SpanNameFormatter == nil {
		cfg.SpanNameFormatter = defaultSpanNameFormatter
	}

	instrumentationName := cfg.Name
	if instrumentationName == "" {
		instrumentationName = defaultInstrumentationName
	}

	tracer := telemetry.TelemetryFacade.Tracer(instrumentationName)
	meter := telemetry.TelemetryFacade.Meter(instrumentationName)
	propagator := telemetry.TelemetryFacade.Propagator()

	durationHist, _ := meter.Float64Histogram("http.server.request.duration", metric.WithUnit("s"), metric.WithDescription("Duration of HTTP server requests"))
	requestSizeHist, _ := meter.Int64Histogram("http.server.request.body.size", metric.WithUnit("By"), metric.WithDescription("Size of HTTP server request body"))
	responseSizeHist, _ := meter.Int64Histogram("http.server.response.body.size", metric.WithUnit("By"), metric.WithDescription("Size of HTTP server response body"))

	excludedPathsMap := sliceToMap(cfg.ExcludedPaths)
	excludedMethodsMap := sliceToMap(cfg.ExcludedMethods)
	return func(ctx http.Context) {
		start := time.Now()
		req := ctx.Request()

		routePattern := req.OriginPath()
		if routePattern == "" {
			routePattern = req.Path()
		}
		if excludedPathsMap[routePattern] || excludedMethodsMap[req.Method()] {
			req.Next()
			return
		}

		for _, f := range cfg.Filters {
			if !f(ctx) {
				req.Next()
				return
			}
		}

		parentCtx := propagator.Extract(ctx.Context(), telemetry.PropagationHeaderCarrier(req.Headers()))

		spanName := cfg.SpanNameFormatter(routePattern, ctx)

		scheme := "http"
		if proto := req.Header("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		}

		baseAttrs := []telemetry.KeyValue{
			semconv.HTTPRequestMethodOriginal(req.Method()),
			semconv.HTTPRoute(routePattern),
			semconv.URLScheme(scheme),
			semconv.ServerAddress(req.Host()),
			semconv.ClientAddress(req.Ip()),
			semconv.UserAgentOriginal(req.Header("User-Agent")),
		}

		spanCtx, span := tracer.Start(parentCtx, spanName,
			telemetry.WithAttributes(baseAttrs...),
			telemetry.WithSpanKind(telemetry.SpanKindServer),
		)

		ctx.WithContext(spanCtx)

		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic: %v", r)
				span.RecordError(err)
				span.SetStatus(telemetry.CodeError, "Internal Server Error")
				span.End()

				metricAttrs := metric.WithAttributes(append(baseAttrs, semconv.HTTPResponseStatusCode(500))...)
				durationHist.Record(spanCtx, time.Since(start).Seconds(), metricAttrs)

				panic(r)
			} else {
				span.End()
			}
		}()

		req.Next()

		status := ctx.Response().Origin().Status()
		span.SetAttributes(semconv.HTTPResponseStatusCode(status))

		if status >= 500 {
			span.SetStatus(telemetry.CodeError, "")
		} else {
			span.SetStatus(telemetry.CodeOk, "")
		}

		if err := ctx.Err(); err != nil {
			span.RecordError(err)
		}

		metricAttrs := metric.WithAttributes(append(baseAttrs, semconv.HTTPResponseStatusCode(status))...)

		reqSize := req.Origin().ContentLength
		if reqSize < 0 {
			reqSize = 0
		}
		requestSizeHist.Record(spanCtx, reqSize, metricAttrs)

		responseSizeHist.Record(spanCtx, int64(ctx.Response().Origin().Size()), metricAttrs)

		durationHist.Record(spanCtx, time.Since(start).Seconds(), metricAttrs)
	}
}

func noopMiddleware() http.Middleware {
	return func(ctx http.Context) {
		ctx.Request().Next()
	}
}

func sliceToMap(slice []string) map[string]bool {
	m := make(map[string]bool, len(slice))
	for _, s := range slice {
		m[s] = true
	}
	return m
}
