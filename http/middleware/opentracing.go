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
	return func(ctx http.Context) {
		var parentSpan opentracing.Span

		spCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(ctx.Request().Headers()))
		if err != nil {
			parentSpan = tracer.StartSpan(ctx.Request().Path())
			defer parentSpan.Finish()
		} else {
			parentSpan = opentracing.StartSpan(
				ctx.Request().Path(),
				opentracing.ChildOf(spCtx),
				opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
				ext.SpanKindRPCServer,
			)
			defer parentSpan.Finish()
		}

		ctx.WithValue(OpentracingTracer, tracer)
		ctx.WithValue(OpentracingCtx, opentracing.ContextWithSpan(context.Background(), parentSpan))
		ctx.Request().Next()
	}
}
