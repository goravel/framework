package middleware

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/goravel/framework/contracts/http"
)

const (
	OpentracingTracer = "opentracing_tracer"
	OpentracingCtx    = "opentracing_ctx"
)

func Opentracing(tracer opentracing.Tracer) http.Middleware {
	return func(request http.Request) {
		var parentSpan opentracing.Span

		spCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(request.Headers()))
		if err != nil {
			parentSpan = tracer.StartSpan(request.Path())
			defer parentSpan.Finish()
		} else {
			parentSpan = opentracing.StartSpan(
				request.Path(),
				opentracing.ChildOf(spCtx),
				opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
				ext.SpanKindRPCServer,
			)
			defer parentSpan.Finish()
		}

		context.WithValue(request.Context(), OpentracingTracer, tracer)
		context.WithValue(request.Context(), OpentracingCtx, opentracing.ContextWithSpan(context.Background(), parentSpan))
		request.Next()
	}
}
