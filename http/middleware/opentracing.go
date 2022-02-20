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
	return func(c *gin.Context) {
		var parentSpan opentracing.Span

		spCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		if err != nil {
			parentSpan = tracer.StartSpan(c.Request.URL.Path)
			defer parentSpan.Finish()
		} else {
			parentSpan = opentracing.StartSpan(
				c.Request.URL.Path,
				opentracing.ChildOf(spCtx),
				opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
				ext.SpanKindRPCServer,
			)
			defer parentSpan.Finish()
		}

		c.Set(OpentracingTracer, tracer)
		c.Set(OpentracingCtx, opentracing.ContextWithSpan(context.Background(), parentSpan))
		c.Next()
	}
}
