package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
)

// FileHandler is a slog.Handler that writes formatted log entries to a file.
type FileHandler struct {
	writer io.Writer
	config config.Config
	json   foundation.Json
	level  slog.Leveler
	attrs  []slog.Attr
	groups []string
}

// NewFileHandler creates a new file handler.
func NewFileHandler(w io.Writer, config config.Config, json foundation.Json, level slog.Leveler) *FileHandler {
	return &FileHandler{
		writer: w,
		config: config,
		json:   json,
		level:  level,
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *FileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle handles the Record.
func (h *FileHandler) Handle(ctx context.Context, r slog.Record) error {
	var b bytes.Buffer

	timestamp := carbon.FromStdTime(r.Time, carbon.DefaultTimezone()).ToDateTimeMilliString()
	levelStr := levelToString(log.Level(r.Level))
	env := h.config.GetString("app.env")

	fmt.Fprintf(&b, "[%s] %s.%s: %s\n", timestamp, env, levelStr, r.Message)

	// Format attributes
	formattedData, err := h.formatAttrs(r)
	if err != nil {
		return err
	}
	b.WriteString(formattedData)

	_, err = h.writer.Write(b.Bytes())
	return err
}

// WithAttrs returns a new handler with the given attributes.
func (h *FileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &FileHandler{
		writer: h.writer,
		config: h.config,
		json:   h.json,
		level:  h.level,
		attrs:  append(h.attrs, attrs...),
		groups: h.groups,
	}
}

// WithGroup returns a new handler with the given group.
func (h *FileHandler) WithGroup(name string) slog.Handler {
	return &FileHandler{
		writer: h.writer,
		config: h.config,
		json:   h.json,
		level:  h.level,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}

func (h *FileHandler) formatAttrs(r slog.Record) (string, error) {
	var builder strings.Builder
	var rootData map[string]any

	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "root" {
			rootData = extractGroupData(a.Value)
		}
		return true
	})

	if rootData == nil {
		return "", nil
	}

	for _, key := range []string{"hint", "tags", "owner", "context", "with", "domain", "code", "request", "response", "user"} {
		if value, exists := rootData[key]; exists && value != nil {
			builder.WriteString(fmt.Sprintf("[%s] %+v\n", str.Of(key).UcFirst().String(), value))
		}
	}

	if stackTraceValue, exists := rootData["stacktrace"]; exists && stackTraceValue != nil {
		traces, err := h.formatStackTraces(stackTraceValue)
		if err != nil {
			return "", err
		}
		builder.WriteString(traces)
	}

	return builder.String(), nil
}

type StackTrace struct {
	Root struct {
		Message string   `json:"message"`
		Stack   []string `json:"stack"`
	} `json:"root"`
	Wrap []struct {
		Message string `json:"message"`
		Stack   string `json:"stack"`
	} `json:"wrap"`
}

func (h *FileHandler) formatStackTraces(stackTraces any) (string, error) {
	var formattedTraces strings.Builder
	data, err := h.json.Marshal(stackTraces)

	if err != nil {
		return "", err
	}
	var traces StackTrace
	err = h.json.Unmarshal(data, &traces)
	if err != nil {
		return "", err
	}
	formattedTraces.WriteString("[Trace]\n")
	root := traces.Root
	if len(root.Stack) > 0 {
		for _, stackStr := range root.Stack {
			formattedTraces.WriteString(formatStackTrace(stackStr))
		}
	}

	return formattedTraces.String(), nil
}

func formatStackTrace(stackStr string) string {
	lastColon := strings.LastIndex(stackStr, ":")
	if lastColon > 0 && lastColon < len(stackStr)-1 {
		secondLastColon := strings.LastIndex(stackStr[:lastColon], ":")
		if secondLastColon > 0 {
			fileLine := stackStr[secondLastColon+1:]
			method := stackStr[:secondLastColon]
			return fmt.Sprintf("%s [%s]\n", fileLine, method)
		}
	}
	return fmt.Sprintf("%s\n", stackStr)
}

// extractGroupData extracts map data from a slog.Value.
func extractGroupData(v slog.Value) map[string]any {
	result := make(map[string]any)

	switch v.Kind() {
	case slog.KindGroup:
		for _, attr := range v.Group() {
			result[attr.Key] = extractValue(attr.Value)
		}
	case slog.KindAny:
		if m, ok := v.Any().(map[string]any); ok {
			return m
		}
	}

	return result
}

func extractValue(v slog.Value) any {
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindInt64:
		return v.Int64()
	case slog.KindFloat64:
		return v.Float64()
	case slog.KindBool:
		return v.Bool()
	case slog.KindTime:
		return v.Time()
	case slog.KindDuration:
		return v.Duration()
	case slog.KindGroup:
		result := make(map[string]any)
		for _, attr := range v.Group() {
			result[attr.Key] = extractValue(attr.Value)
		}
		return result
	case slog.KindAny:
		return v.Any()
	default:
		return v.Any()
	}
}

func levelToString(level log.Level) string {
	switch level {
	case log.DebugLevel:
		return "debug"
	case log.InfoLevel:
		return "info"
	case log.WarningLevel:
		return "warning"
	case log.ErrorLevel:
		return "error"
	case log.FatalLevel:
		return "fatal"
	case log.PanicLevel:
		return "panic"
	default:
		return "unknown"
	}
}

// ConsoleHandler is a slog.Handler that writes formatted log entries to stdout.
type ConsoleHandler struct {
	*FileHandler
}

// NewConsoleHandler creates a new console handler.
func NewConsoleHandler(config config.Config, json foundation.Json) *ConsoleHandler {
	return &ConsoleHandler{
		FileHandler: &FileHandler{
			writer: os.Stdout,
			config: config,
			json:   json,
			level:  log.DebugLevel,
		},
	}
}
