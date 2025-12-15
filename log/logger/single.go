package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/log/formatter"
	"github.com/goravel/framework/support"
)

type Single struct {
	config config.Config
	json   foundation.Json
}

func NewSingle(config config.Config, json foundation.Json) *Single {
	return &Single{
		config: config,
		json:   json,
	}
}

func (single *Single) Handle(channel string) (slog.Handler, error) {
	logPath := single.config.GetString(channel + ".path")
	if logPath == "" {
		return nil, errors.LogEmptyLogFilePath
	}

	logPath = filepath.Join(support.RelativePath, logPath)
	level := getLevelFromString(single.config.GetString(channel + ".level"))

	// Ensure parent directory exists
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Open log file for appending
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	var writers []io.Writer
	writers = append(writers, file)

	if single.config.GetBool(channel + ".print") {
		writers = append(writers, os.Stdout)
	}

	// Create the slog handler
	handler := formatter.NewGeneralHandler(single.config, single.json, io.MultiWriter(writers...), level)
	return handler, nil
}

func getLevelFromString(level string) log.Level {
	switch level {
	case "panic":
		return log.LevelPanic
	case "fatal":
		return log.LevelFatal
	case "error":
		return log.LevelError
	case "warning", "warn":
		return log.LevelWarning
	case "info":
		return log.LevelInfo
	case "debug":
		return log.LevelDebug
	default:
		return log.LevelDebug
	}
}

// levelHandler is a slog.Handler that filters by level
type levelHandler struct {
	level   log.Level
	handler slog.Handler
}

func (h *levelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return log.Level(level) >= h.level
}

func (h *levelHandler) Handle(ctx context.Context, record slog.Record) error {
	return h.handler.Handle(ctx, record)
}

func (h *levelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelHandler{level: h.level, handler: h.handler.WithAttrs(attrs)}
}

func (h *levelHandler) WithGroup(name string) slog.Handler {
	return &levelHandler{level: h.level, handler: h.handler.WithGroup(name)}
}
