package log

import (
	"context"
	"io"
	"log/slog"
)

// MultiHandler wraps multiple slog.Handler instances to support multiple channels/hooks
type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			if err := handler.Handle(ctx, record); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (h *MultiHandler) AddHandler(handler slog.Handler) {
	h.handlers = append(h.handlers, handler)
}

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

// NewSlogLogger creates a new slog.Logger with a discard handler by default
func NewSlogLogger() *slog.Logger {
	return slog.New(NewDiscardHandler())
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
