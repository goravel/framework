package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/goravel/framework/facades"
)

type MDReaderWriter struct {
	metadata.MD
}

func (c MDReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range c.MD {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c MDReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	c.MD[key] = append(c.MD[key], val)
}

func OpentracingClient(tracer opentracing.Tracer, parentCtx context.Context) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		span, _ := opentracing.StartSpanFromContext(parentCtx,
			"call gRPC",
			opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
			ext.SpanKindRPCClient)

		defer span.Finish()

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}

		err := tracer.Inject(span.Context(), opentracing.TextMap, MDReaderWriter{md})
		if err != nil {
			span.LogFields(log.Error(errors.WithMessage(err, "inject-error")))
		}

		newCtx := metadata.NewOutgoingContext(ctx, md)
		err = invoker(newCtx, method, req, reply, cc, opts...)
		if err != nil {
			span.LogFields(log.Error(errors.WithMessage(err, "call-error")))
		}
		return err
	}
}

func OpentracingServer(tracer opentracing.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
		resp any, err error) {
		var parentCtx context.Context

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		spanContext, err := tracer.Extract(opentracing.TextMap, MDReaderWriter{md})
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			facades.Log.Error(fmt.Sprintf("extract from metadata error: %v", err))
		} else {
			span := tracer.StartSpan(
				info.FullMethod,
				ext.RPCServerOption(spanContext),
				opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
				ext.SpanKindRPCServer,
			)
			defer span.Finish()
			parentCtx = opentracing.ContextWithSpan(ctx, span)
		}
		return handler(parentCtx, req)
	}
}
