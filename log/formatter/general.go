package formatter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
)

type General struct {
	config config.Config
	json   foundation.Json
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

func NewGeneral(config config.Config, json foundation.Json) *General {
	return &General{
		config: config,
		json:   json,
	}
}

// GeneralHandler is a slog.Handler that formats logs in the Goravel general format.
type GeneralHandler struct {
	config config.Config
	json   foundation.Json
	writer io.Writer
	level  log.Level
	attrs  []slog.Attr
	groups []string
	mu     sync.Mutex
}

// NewGeneralHandler creates a new GeneralHandler.
func NewGeneralHandler(config config.Config, json foundation.Json, writer io.Writer, level log.Level) *GeneralHandler {
	return &GeneralHandler{
		config: config,
		json:   json,
		writer: writer,
		level:  level,
		attrs:  []slog.Attr{},
		groups: []string{},
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *GeneralHandler) Enabled(_ context.Context, level slog.Level) bool {
	return log.Level(level) >= h.level
}

// Handle handles the Record.
func (h *GeneralHandler) Handle(_ context.Context, record slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var b bytes.Buffer

	timestamp := carbon.FromStdTime(record.Time, carbon.DefaultTimezone()).ToDateTimeMilliString()
	levelStr := levelToString(log.Level(record.Level))
	fmt.Fprintf(&b, "[%s] %s.%s: %s\n", timestamp, h.config.GetString("app.env"), levelStr, record.Message)

	// Collect all attributes
	data := make(map[string]any)
	record.Attrs(func(attr slog.Attr) bool {
		data[attr.Key] = attr.Value.Any()
		return true
	})

	// Also add handler-level attrs
	for _, attr := range h.attrs {
		data[attr.Key] = attr.Value.Any()
	}

	if len(data) > 0 {
		formattedData, err := h.formatData(data)
		if err != nil {
			return err
		}
		b.WriteString(formattedData)
	}

	_, err := h.writer.Write(b.Bytes())
	return err
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
func (h *GeneralHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := &GeneralHandler{
		config: h.config,
		json:   h.json,
		writer: h.writer,
		level:  h.level,
		attrs:  make([]slog.Attr, len(h.attrs)+len(attrs)),
		groups: make([]string, len(h.groups)),
	}
	copy(newHandler.attrs, h.attrs)
	copy(newHandler.attrs[len(h.attrs):], attrs)
	copy(newHandler.groups, h.groups)
	return newHandler
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
func (h *GeneralHandler) WithGroup(name string) slog.Handler {
	newHandler := &GeneralHandler{
		config: h.config,
		json:   h.json,
		writer: h.writer,
		level:  h.level,
		attrs:  make([]slog.Attr, len(h.attrs)),
		groups: make([]string, len(h.groups)+1),
	}
	copy(newHandler.attrs, h.attrs)
	copy(newHandler.groups, h.groups)
	newHandler.groups[len(h.groups)] = name
	return newHandler
}

func (h *GeneralHandler) formatData(data map[string]any) (string, error) {
	var builder strings.Builder

	if len(data) > 0 {
		removedData := deleteKey(data, "root")
		if len(removedData) > 0 {
			removedDataStr, err := h.json.MarshalString(removedData)
			if err != nil {
				return "", err
			}

			builder.WriteString(fmt.Sprintf("fields: %s\n", removedDataStr))
		}

		root, err := cast.ToStringMapE(data["root"])
		if err != nil {
			// If root key doesn't exist or is not a map, just return what we have so far
			return builder.String(), nil
		}

		for _, key := range []string{"hint", "tags", "owner", "context", "with", "domain", "code", "request", "response", "user"} {
			if value, exists := root[key]; exists && value != nil {
				builder.WriteString(fmt.Sprintf("[%s] %+v\n", str.Of(key).UcFirst().String(), value))
			}
		}

		if stackTraceValue, exists := root["stacktrace"]; exists && stackTraceValue != nil {
			traces, err := h.formatStackTraces(stackTraceValue)
			if err != nil {
				return "", err
			}

			builder.WriteString(traces)
		}
	}

	return builder.String(), nil
}

func (h *GeneralHandler) formatStackTraces(stackTraces any) (string, error) {
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

func deleteKey(data map[string]any, keyToDelete string) map[string]any {
	dataCopy := make(map[string]any)
	for key, value := range data {
		if key != keyToDelete {
			dataCopy[key] = value
		}
	}

	return dataCopy
}

func levelToString(level log.Level) string {
	switch level {
	case log.LevelDebug:
		return "debug"
	case log.LevelInfo:
		return "info"
	case log.LevelWarning:
		return "warning"
	case log.LevelError:
		return "error"
	case log.LevelFatal:
		return "fatal"
	case log.LevelPanic:
		return "panic"
	default:
		return "unknown"
	}
}
