package log

import (
	"context"
	"log/slog"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/log"
)

// LegacyHandlerAdapter adapts slog.Handler to work with our Entry interface.
// This is used internally to bridge custom handlers with our logging system.
type LegacyHandlerAdapter struct {
	handler  slog.Handler
	attrs    []slog.Attr
	groups   []string
}

func NewLegacyHandlerAdapter(handler slog.Handler) *LegacyHandlerAdapter {
	return &LegacyHandlerAdapter{
		handler: handler,
	}
}

func (h *LegacyHandlerAdapter) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *LegacyHandlerAdapter) Handle(ctx context.Context, record slog.Record) error {
	// Create a new record with accumulated attrs
	newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(a slog.Attr) bool {
		newRecord.AddAttrs(a)
		return true
	})
	for _, attr := range h.attrs {
		newRecord.AddAttrs(attr)
	}

	return h.handler.Handle(ctx, newRecord)
}

func (h *LegacyHandlerAdapter) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &LegacyHandlerAdapter{
		handler: h.handler,
		attrs:   newAttrs,
		groups:  h.groups,
	}
}

func (h *LegacyHandlerAdapter) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &LegacyHandlerAdapter{
		handler: h.handler,
		attrs:   h.attrs,
		groups:  newGroups,
	}
}

// createEntryFromRecord creates an Entry from a slog.Record for handlers that need it
func createEntryFromRecord(ctx context.Context, record slog.Record) *Entry {
	// Collect all attributes
	attrs := make(map[string]any)
	record.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	e := &Entry{
		ctx:     ctx,
		data:    attrs,
		level:   log.FromSlog(record.Level),
		message: record.Message,
		time:    record.Time,
	}

	if len(attrs) > 0 {
		root, err := cast.ToStringMapE(attrs["root"])
		if err == nil {
			if code, err := cast.ToStringE(root["code"]); err == nil {
				e.code = code
			}
			if domain, err := cast.ToStringE(root["domain"]); err == nil {
				e.domain = domain
			}
			if hint, err := cast.ToStringE(root["hint"]); err == nil {
				e.hint = hint
			}

			e.owner = root["owner"]

			if request, err := cast.ToStringMapE(root["request"]); err == nil {
				e.request = request
			}
			if response, err := cast.ToStringMapE(root["response"]); err == nil {
				e.response = response
			}
			if stacktrace, err := cast.ToStringMapE(root["stacktrace"]); err == nil {
				e.stacktrace = stacktrace
			}
			if tags, err := cast.ToStringSliceE(root["tags"]); err == nil {
				e.tags = tags
			}

			e.user = root["user"]

			if with, err := cast.ToStringMapE(root["with"]); err == nil {
				e.with = with
			}
		}
	}

	return e
}

