package log

import (
	"context"
	"io"
	"log/slog"
)

// DiscardHandler is a handler that discards all logs
type DiscardHandler struct{}

func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}

func (h *DiscardHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *DiscardHandler) Handle(ctx context.Context, record slog.Record) error {
	return nil
}

func (h *DiscardHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *DiscardHandler) WithGroup(name string) slog.Handler {
	return h
}

// LeveledHandler filters logs based on minimum level
type LeveledHandler struct {
	handler  slog.Handler
	minLevel slog.Level
}

func NewLeveledHandler(handler slog.Handler, minLevel slog.Level) *LeveledHandler {
	return &LeveledHandler{
		handler:  handler,
		minLevel: minLevel,
	}
}

func (h *LeveledHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.minLevel && h.handler.Enabled(ctx, level)
}

func (h *LeveledHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.Level >= h.minLevel {
		return h.handler.Handle(ctx, record)
	}
	return nil
}

func (h *LeveledHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LeveledHandler{
		handler:  h.handler.WithAttrs(attrs),
		minLevel: h.minLevel,
	}
}

func (h *LeveledHandler) WithGroup(name string) slog.Handler {
	return &LeveledHandler{
		handler:  h.handler.WithGroup(name),
		minLevel: h.minLevel,
	}
}

// WriterHandler wraps an io.Writer to use as a slog.Handler
type WriterHandler struct {
	writer    io.Writer
	formatter Formatter
	attrs     []slog.Attr
	groups    []string
}

func NewWriterHandler(w io.Writer, formatter Formatter) *WriterHandler {
	return &WriterHandler{
		writer:    w,
		formatter: formatter,
	}
}

func (h *WriterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *WriterHandler) Handle(ctx context.Context, record slog.Record) error {
	// Create a new record with accumulated attrs
	newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(a slog.Attr) bool {
		newRecord.AddAttrs(a)
		return true
	})
	for _, attr := range h.attrs {
		newRecord.AddAttrs(attr)
	}
	
	formatted, err := h.formatter.Format(ctx, newRecord)
	if err != nil {
		return err
	}
	_, err = h.writer.Write(formatted)
	return err
}

func (h *WriterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	
	return &WriterHandler{
		writer:    h.writer,
		formatter: h.formatter,
		attrs:     newAttrs,
		groups:    h.groups,
	}
}

func (h *WriterHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	
	return &WriterHandler{
		writer:    h.writer,
		formatter: h.formatter,
		attrs:     h.attrs,
		groups:    newGroups,
	}
}

// Formatter interface for formatting slog records
type Formatter interface {
	Format(ctx context.Context, record slog.Record) ([]byte, error)
}
