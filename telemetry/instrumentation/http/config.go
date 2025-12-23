package http

import (
	"fmt"

	"github.com/goravel/framework/contracts/http"
)

type Filter func(ctx http.Context) bool

type SpanNameFormatter func(route string, ctx http.Context) string

type Option func(*ServerConfig)

func WithSpanNameFormatter(f SpanNameFormatter) Option {
	return func(c *ServerConfig) {
		c.SpanNameFormatter = f
	}
}

func WithFilter(f Filter) Option {
	return func(c *ServerConfig) {
		c.Filters = append(c.Filters, f)
	}
}

// ServerConfig maps to the "telemetry.instrumentation.http_server" key in the config file.
type ServerConfig struct {
	Enabled           bool              `mapstructure:"enabled"`
	Name              string            `mapstructure:"name"`
	ExcludedPaths     []string          `mapstructure:"excluded_paths"`
	ExcludedMethods   []string          `mapstructure:"excluded_methods"`
	SpanNameFormatter SpanNameFormatter `mapstructure:"span_name_formatter"`
	Filters           []Filter          `mapstructure:"filters"`
}

func defaultSpanNameFormatter(route string, ctx http.Context) string {
	return fmt.Sprintf("%s %s", ctx.Request().Method(), route)
}
