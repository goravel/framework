package logger

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
)

// FileHandler is a log.Handler that writes formatted log entries to a file.
type FileHandler struct {
	writer io.Writer
	config config.Config
	json   foundation.Json
	level  slog.Leveler
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
func (h *FileHandler) Enabled(level log.Level) bool {
	return slog.Level(level) >= h.level.Level()
}

// Handle handles the Record.
func (h *FileHandler) Handle(entry log.Entry) error {
	var b bytes.Buffer

	timestamp := entry.Time().UnixMilli()
	env := h.config.GetString("app.env")

	_, _ = fmt.Fprintf(&b, "[%d] %s.%s: %s\n", timestamp, env, entry.Level().String(), entry.Message())

	// Format Entry
	if v := entry.Code(); v != "" {
		_, _ = fmt.Fprintf(&b, "[Code] %+v\n", v)
	}
	if v := entry.Context(); v != nil {
		values := make(map[any]any)
		getContextValues(v, values)
		if len(values) > 0 {
			_, _ = fmt.Fprintf(&b, "[Context] %+v\n", values)
		}
	}
	if v := entry.Domain(); v != "" {
		_, _ = fmt.Fprintf(&b, "[Domain] %+v\n", v)
	}
	if v := entry.Hint(); v != "" {
		_, _ = fmt.Fprintf(&b, "[Hint] %+v\n", v)
	}
	if v := entry.Owner(); v != nil {
		_, _ = fmt.Fprintf(&b, "[Owner] %+v\n", v)
	}
	if v := entry.Request(); v != nil {
		_, _ = fmt.Fprintf(&b, "[Request] %+v\n", v)
	}
	if v := entry.Response(); v != nil {
		_, _ = fmt.Fprintf(&b, "[Response] %+v\n", v)
	}
	if v := entry.Trace(); v != nil {
		traces, err := formatStackTraces(h.json, v)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(&b, "[Trace] %+v\n", traces)
	}
	if v := entry.Tags(); v != nil {
		_, _ = fmt.Fprintf(&b, "[Tags] %+v\n", v)
	}
	if v := entry.User(); v != nil {
		_, _ = fmt.Fprintf(&b, "[User] %+v\n", v)
	}
	if v := entry.With(); v != nil {
		_, _ = fmt.Fprintf(&b, "[With] %+v\n", v)
	}

	_, err := h.writer.Write(b.Bytes())
	return err
}

// ConsoleHandler is a log.Handler that writes formatted log entries to stdout.
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
			level:  log.LevelDebug,
		},
	}
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

func formatStackTraces(json foundation.Json, stackTraces any) (string, error) {
	var formattedTraces strings.Builder
	data, err := json.Marshal(stackTraces)

	if err != nil {
		return "", err
	}
	var traces StackTrace
	err = json.Unmarshal(data, &traces)
	if err != nil {
		return "", err
	}
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
