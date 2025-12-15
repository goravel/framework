package logger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
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
	
	// Ensure directory exists
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	
	// Open or create the log file
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	minLevel := getSlogLevel(single.config.GetString(channel + ".level"))
	generalFormatter := formatter.NewGeneral(single.config, single.json)
	
	// Create a custom handler that uses our formatter
	handler := &formatterHandler{
		writer:    file,
		formatter: generalFormatter,
		minLevel:  minLevel,
	}

	return handler, nil
}

func getSlogLevel(level string) slog.Level {
	switch level {
	case "panic":
		return slog.Level(12) // Higher than Error
	case "fatal":
		return slog.Level(12)
	case "error":
		return slog.LevelError
	case "warning":
		return slog.LevelWarn
	case "info":
		return slog.LevelInfo
	case "debug":
		return slog.LevelDebug
	default:
		return slog.LevelDebug
	}
}

// formatterHandler wraps our formatter as a slog.Handler
type formatterHandler struct {
	writer    *os.File
	formatter *formatter.General
	minLevel  slog.Level
	attrs     []slog.Attr
	groups    []string
}

func (h *formatterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func (h *formatterHandler) Handle(ctx context.Context, record slog.Record) error {
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
	
	formatted, err := h.formatter.Format(ctx, newRecord)
	if err != nil {
		return err
	}
	
	_, err = h.writer.Write(formatted)
	return err
}

func (h *formatterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	
	return &formatterHandler{
		writer:    h.writer,
		formatter: h.formatter,
		minLevel:  h.minLevel,
		attrs:     newAttrs,
		groups:    h.groups,
	}
}

func (h *formatterHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	
	return &formatterHandler{
		writer:    h.writer,
		formatter: h.formatter,
		minLevel:  h.minLevel,
		attrs:     h.attrs,
		groups:    newGroups,
	}
}
