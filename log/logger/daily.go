package logger

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/goravel/file-rotatelogs/v2"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/log/formatter"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/carbon"
)

type Daily struct {
	config config.Config
	json   foundation.Json
}

func NewDaily(config config.Config, json foundation.Json) *Daily {
	return &Daily{
		config: config,
		json:   json,
	}
}

func (daily *Daily) Handle(channel string) (slog.Handler, error) {
	logPath := daily.config.GetString(channel + ".path")
	if logPath == "" {
		return nil, errors.LogEmptyLogFilePath
	}

	ext := filepath.Ext(logPath)
	logPath = strings.ReplaceAll(logPath, ext, "")
	logPath = filepath.Join(support.RelativePath, logPath)

	writer, err := rotatelogs.New(
		logPath+"-%Y-%m-%d"+ext,
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
		rotatelogs.WithRotationCount(uint(daily.config.GetInt(channel+".days"))),
		rotatelogs.WithClock(rotatelogs.NewClock(carbon.Now().StdTime())),
	)
	if err != nil {
		return nil, err
	}

	minLevel := getSlogLevel(daily.config.GetString(channel + ".level"))
	generalFormatter := formatter.NewGeneral(daily.config, daily.json)

	// Create a custom handler that uses our formatter
	handler := &dailyFormatterHandler{
		writer:    writer,
		formatter: generalFormatter,
		minLevel:  minLevel,
	}

	return handler, nil
}

// dailyFormatterHandler wraps our formatter as a slog.Handler for daily rotation
type dailyFormatterHandler struct {
	writer    *rotatelogs.RotateLogs
	formatter *formatter.General
	minLevel  slog.Level
	attrs     []slog.Attr
	groups    []string
}

func (h *dailyFormatterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func (h *dailyFormatterHandler) Handle(ctx context.Context, record slog.Record) error {
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

func (h *dailyFormatterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	
	return &dailyFormatterHandler{
		writer:    h.writer,
		formatter: h.formatter,
		minLevel:  h.minLevel,
		attrs:     newAttrs,
		groups:    h.groups,
	}
}

func (h *dailyFormatterHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	
	return &dailyFormatterHandler{
		writer:    h.writer,
		formatter: h.formatter,
		minLevel:  h.minLevel,
		attrs:     h.attrs,
		groups:    newGroups,
	}
}
