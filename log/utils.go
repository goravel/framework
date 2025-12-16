package log

import (
	"context"
	"log/slog"

	"github.com/goravel/framework/contracts/log"
)

// slogAdapter wraps a log.Handler to implement slog.Handler
type slogAdapter struct {
	handler log.Handler
}

func (h *slogAdapter) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(log.Level(level))
}

func (h *slogAdapter) Handle(ctx context.Context, record slog.Record) error {
	entry := FromSlogRecord(record)
	return h.handler.Handle(entry)
}

func (h *slogAdapter) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *slogAdapter) WithGroup(name string) slog.Handler {
	return h
}

func HandlerToSlogHandler(handler log.Handler) slog.Handler {
	return &slogAdapter{handler: handler}
}
