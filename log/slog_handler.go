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
	formatted, err := h.formatter.Format(ctx, record)
	if err != nil {
		return err
	}
	_, err = h.writer.Write(formatted)
	return err
}

func (h *WriterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *WriterHandler) WithGroup(name string) slog.Handler {
	return h
}

// Formatter interface for formatting slog records
type Formatter interface {
	Format(ctx context.Context, record slog.Record) ([]byte, error)
}
