package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	OpentracingTracer = "opentracing_tracer"
	OpentracingCtx    = "opentracing_ctx"
)

func Opentracing(tracer opentracing.Tracer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var parentSpan opentracing.Span

		spCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(ctx.Request.Header))
		if err != nil {
			parentSpan = tracer.StartSpan(ctx.Request.URL.Path)
			defer parentSpan.Finish()
		} else {
			parentSpan = opentracing.StartSpan(
				ctx.Request.URL.Path,
				opentracing.ChildOf(spCtx),
				opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
				ext.SpanKindRPCServer,
			)
			defer parentSpan.Finish()
		}

		ctx.Set(OpentracingTracer, tracer)
		ctx.Set(OpentracingCtx, opentracing.ContextWithSpan(context.Background(), parentSpan))
		ctx.Next()
	}
}
