package log

import (
	"context"
	"log/slog"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/log"
)

// Hook wraps a custom log.Hook as a logrus-compatible hook (legacy support)
type Hook struct {
	instance log.Hook
}

func (h *Hook) Levels() []log.Level {
	return h.instance.Levels()
}

func (h *Hook) Fire(ctx context.Context, record slog.Record) error {
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
		if err != nil {
			return err
		}

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

	return h.instance.Fire(e)
}

// HookHandler wraps a custom log.Hook as a slog.Handler
type HookHandler struct {
	hook   log.Hook
	attrs  []slog.Attr
	groups []string
}

func (h *HookHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Check if the hook supports this level
	ourLevel := log.FromSlog(level)
	for _, hookLevel := range h.hook.Levels() {
		if hookLevel == ourLevel {
			return true
		}
	}
	return false
}

func (h *HookHandler) Handle(ctx context.Context, record slog.Record) error {
	if !h.Enabled(ctx, record.Level) {
		return nil
	}

	// Create a new record with accumulated attrs
	newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(a slog.Attr) bool {
		newRecord.AddAttrs(a)
		return true
	})
	for _, attr := range h.attrs {
		newRecord.AddAttrs(attr)
	}

	// Collect all attributes
	attrs := make(map[string]any)
	newRecord.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	e := &Entry{
		ctx:     ctx,
		data:    attrs,
		level:   log.FromSlog(newRecord.Level),
		message: newRecord.Message,
		time:    newRecord.Time,
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

	return h.hook.Fire(e)
}

func (h *HookHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &HookHandler{
		hook:   h.hook,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

func (h *HookHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &HookHandler{
		hook:   h.hook,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

