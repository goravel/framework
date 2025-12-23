package grpc

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/stats"
)

// Option is a type alias so users don't need to import the otelgrpc package.
type Option = otelgrpc.Option

type Event = otelgrpc.Event

const (
	ReceivedEvents = otelgrpc.ReceivedEvents
	SentEvents     = otelgrpc.SentEvents
)

// WithFilter configures a filter to ignore specific requests based on the RPC tag info.
func WithFilter(filter func(info *stats.RPCTagInfo) bool) Option {
	return otelgrpc.WithFilter(filter)
}

// WithMessageEvents configures the handler to record the specified events
// (span.AddEvent) on spans.
//
// Valid events are:
//   - ReceivedEvents: Record the number of bytes read.
//   - SentEvents: Record the number of bytes written.
func WithMessageEvents(events ...Event) Option {
	return otelgrpc.WithMessageEvents(events...)
}

// WithSpanAttributes configures custom attributes for the spans.
func WithSpanAttributes(attrs ...attribute.KeyValue) Option {
	return otelgrpc.WithSpanAttributes(attrs...)
}

// WithMetricAttributes configures custom attributes for the metrics.
func WithMetricAttributes(attrs ...attribute.KeyValue) Option {
	return otelgrpc.WithMetricAttributes(attrs...)
}