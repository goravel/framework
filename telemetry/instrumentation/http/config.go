package http

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"

	"github.com/goravel/framework/contracts/http"
)

// Filter allows excluding specific requests from being traced.
type Filter func(ctx http.Context) bool

// SpanNameFormatter allows customizing the span name.
type SpanNameFormatter func(route string, ctx http.Context) string

// Option applies configuration to the server instrumentation.
type Option func(*ServerConfig)

// ServerConfig maps to "telemetry.instrumentation.http_server".
type ServerConfig struct {
	Enabled           bool                 `json:"enabled"`
	ExcludedPaths     []string             `json:"excluded_paths"`
	ExcludedMethods   []string             `json:"excluded_methods"`
	Filters           []Filter             `json:"-"`
	SpanNameFormatter SpanNameFormatter    `json:"-"`
	MetricAttributes  []attribute.KeyValue `json:"-"`
}

func WithFilter(f Filter) Option {
	return func(c *ServerConfig) {
		c.Filters = append(c.Filters, f)
	}
}

func WithSpanNameFormatter(f SpanNameFormatter) Option {
	return func(c *ServerConfig) {
		c.SpanNameFormatter = f
	}
}

func WithMetricAttributes(attrs ...attribute.KeyValue) Option {
	return func(c *ServerConfig) {
		c.MetricAttributes = append(c.MetricAttributes, attrs...)
	}
}

func defaultSpanNameFormatter(route string, ctx http.Context) string {
	return fmt.Sprintf("%s %s", ctx.Request().Method(), route)
}
