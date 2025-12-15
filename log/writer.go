package log

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"

	"github.com/rotisserie/eris"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/log/formatter"
	"github.com/goravel/framework/log/logger"
)

type Writer struct {
	owner        any
	request      http.ContextRequest
	response     http.ContextResponse
	user         any
	instance     *slog.Logger
	ctx          context.Context
	stacktrace   map[string]any
	with         map[string]any
	code         string
	domain       string
	hint         string
	message      string
	tags         []string
	stackEnabled bool
}

func NewWriter(instance *slog.Logger, ctx context.Context) log.Writer {
	return &Writer{
		code:         "",
		domain:       "",
		hint:         "",
		instance:     instance,
		ctx:          ctx,
		message:      "",
		owner:        nil,
		request:      nil,
		response:     nil,
		stackEnabled: false,
		stacktrace:   nil,
		tags:         []string{},
		user:         nil,
		with:         map[string]any{},
	}
}

func (r *Writer) Debug(args ...any) {
	r.instance.LogAttrs(r.ctx, slog.LevelDebug, fmt.Sprint(args...), slog.Any("root", r.toMap()))
}

func (r *Writer) Debugf(format string, args ...any) {
	r.instance.LogAttrs(r.ctx, slog.LevelDebug, fmt.Sprintf(format, args...), slog.Any("root", r.toMap()))
}

func (r *Writer) Info(args ...any) {
	r.instance.LogAttrs(r.ctx, slog.LevelInfo, fmt.Sprint(args...), slog.Any("root", r.toMap()))
}

func (r *Writer) Infof(format string, args ...any) {
	r.instance.LogAttrs(r.ctx, slog.LevelInfo, fmt.Sprintf(format, args...), slog.Any("root", r.toMap()))
}

func (r *Writer) Warning(args ...any) {
	r.instance.LogAttrs(r.ctx, slog.LevelWarn, fmt.Sprint(args...), slog.Any("root", r.toMap()))
}

func (r *Writer) Warningf(format string, args ...any) {
	r.instance.LogAttrs(r.ctx, slog.LevelWarn, fmt.Sprintf(format, args...), slog.Any("root", r.toMap()))
}

func (r *Writer) Error(args ...any) {
	r.withStackTrace(fmt.Sprint(args...))
	r.instance.LogAttrs(r.ctx, slog.LevelError, fmt.Sprint(args...), slog.Any("root", r.toMap()))
}

func (r *Writer) Errorf(format string, args ...any) {
	r.withStackTrace(fmt.Sprintf(format, args...))
	r.instance.LogAttrs(r.ctx, slog.LevelError, fmt.Sprintf(format, args...), slog.Any("root", r.toMap()))
}

func (r *Writer) Fatal(args ...any) {
	r.withStackTrace(fmt.Sprint(args...))
	r.instance.LogAttrs(r.ctx, slog.Level(12), fmt.Sprint(args...), slog.Any("root", r.toMap()))
	os.Exit(1)
}

func (r *Writer) Fatalf(format string, args ...any) {
	r.withStackTrace(fmt.Sprintf(format, args...))
	r.instance.LogAttrs(r.ctx, slog.Level(12), fmt.Sprintf(format, args...), slog.Any("root", r.toMap()))
	os.Exit(1)
}

func (r *Writer) Panic(args ...any) {
	r.withStackTrace(fmt.Sprint(args...))
	msg := fmt.Sprint(args...)
	r.instance.LogAttrs(r.ctx, slog.Level(12), msg, slog.Any("root", r.toMap()))
	panic(msg)
}

func (r *Writer) Panicf(format string, args ...any) {
	r.withStackTrace(fmt.Sprintf(format, args...))
	msg := fmt.Sprintf(format, args...)
	r.instance.LogAttrs(r.ctx, slog.Level(12), msg, slog.Any("root", r.toMap()))
	panic(msg)
}

// Code set a code or slug that describes the error.
// Error messages are intended to be read by humans, but such code is expected to
// be read by machines and even transported over different services.
func (r *Writer) Code(code string) log.Writer {
	r.code = code
	return r
}

// Hint set a hint for faster debugging.
func (r *Writer) Hint(hint string) log.Writer {
	r.hint = hint

	return r
}

// In sets the feature category or domain in which the log entry is relevant.
func (r *Writer) In(domain string) log.Writer {
	r.domain = domain

	return r
}

// Owner set the name/email of the colleague/team responsible for handling this error.
// Useful for alerting purpose.
func (r *Writer) Owner(owner any) log.Writer {
	r.owner = owner

	return r
}

// Request supplies a http.Request.
func (r *Writer) Request(req http.ContextRequest) log.Writer {
	r.request = req

	return r
}

// Response supplies a http.Response.
func (r *Writer) Response(res http.ContextResponse) log.Writer {
	r.response = res

	return r
}

// Tags add multiple tags, describing the feature returning an error.
func (r *Writer) Tags(tags ...string) log.Writer {
	r.tags = append(r.tags, tags...)

	return r
}

// User sets the user associated with the log entry.
func (r *Writer) User(user any) log.Writer {
	r.user = user
	return r
}

// With adds key-value pairs to the context of the log entry
func (r *Writer) With(data map[string]any) log.Writer {
	maps.Copy(r.with, data)

	return r
}

// WithTrace adds a stack trace to the log entry.
func (r *Writer) WithTrace() log.Writer {
	r.withStackTrace("")
	return r
}

func (r *Writer) withStackTrace(message string) {
	erisNew := eris.New(message)
	if erisNew == nil {
		return
	}

	r.message = erisNew.Error()
	format := eris.NewDefaultJSONFormat(eris.FormatOptions{
		InvertOutput: true,
		WithTrace:    true,
		InvertTrace:  true,
	})
	r.stacktrace = eris.ToCustomJSON(erisNew, format)
	r.stackEnabled = true
}

// resetAll resets all properties of the log.Writer for a new log entry.
func (r *Writer) resetAll() {
	r.code = ""
	r.domain = ""
	r.hint = ""
	r.message = ""
	r.owner = nil
	r.request = nil
	r.response = nil
	r.stacktrace = nil
	r.stackEnabled = false
	r.tags = []string{}
	r.user = nil
	r.with = map[string]any{}
}

// toMap returns a map representation of the error.
func (r *Writer) toMap() map[string]any {
	payload := map[string]any{}

	if code := r.code; code != "" {
		payload["code"] = code
	}
	if r.ctx != nil {
		values := make(map[any]any)
		getContextValues(r.ctx, values)
		if len(values) > 0 {
			payload["context"] = values
		}
	}
	if domain := r.domain; domain != "" {
		payload["domain"] = domain
	}
	if hint := r.hint; hint != "" {
		payload["hint"] = hint
	}
	if message := r.message; message != "" {
		payload["message"] = message
	}
	if owner := r.owner; owner != nil {
		payload["owner"] = owner
	}
	if req := r.request; req != nil {
		payload["request"] = map[string]any{
			"method": req.Method(),
			"uri":    req.FullUrl(),
			"header": req.Headers(),
			"body":   req.All(),
		}
	}
	if res := r.response; res != nil {
		payload["response"] = map[string]any{
			"status": res.Origin().Status(),
			"header": res.Origin().Header(),
			"body":   res.Origin().Body(),
			"size":   res.Origin().Size(),
		}
	}
	if stacktrace := r.stacktrace; stacktrace != nil || r.stackEnabled {
		payload["stacktrace"] = stacktrace
	}
	if tags := r.tags; len(tags) > 0 {
		payload["tags"] = tags
	}
	if r.user != nil {
		payload["user"] = r.user
	}
	if with := r.with; len(with) > 0 {
		payload["with"] = with
	}

	// reset all properties for a new log entry
	r.resetAll()

	return payload
}

func registerHandler(config config.Config, json foundation.Json, multiHandler *MultiHandler, channel string) error {
	var (
		handler slog.Handler
		err     error

		channelPath = "logging.channels." + channel
		driver      = config.GetString(channelPath + ".driver")
	)

	switch driver {
	case log.StackDriver:
		for _, stackChannel := range config.Get(channelPath + ".channels").([]string) {
			if stackChannel == channel {
				return errors.LogDriverCircularReference.Args("stack")
			}

			if err := registerHandler(config, json, multiHandler, stackChannel); err != nil {
				return err
			}
		}

		return nil
	case log.SingleDriver:
		logLogger := logger.NewSingle(config, json)
		handler, err = logLogger.Handle(channelPath)
		if err != nil {
			return err
		}

		if config.GetBool(channelPath + ".print") {
			// Add console handler for print mode
			generalFormatter := formatter.NewGeneral(config, json)
			consoleHandler := &formatterHandler{
				writer:    os.Stdout,
				formatter: generalFormatter,
				minLevel:  slog.LevelDebug,
			}
			multiHandler.AddHandler(consoleHandler)
		}
	case log.DailyDriver:
		logLogger := logger.NewDaily(config, json)
		handler, err = logLogger.Handle(channelPath)
		if err != nil {
			return err
		}

		if config.GetBool(channelPath + ".print") {
			// Add console handler for print mode
			generalFormatter := formatter.NewGeneral(config, json)
			consoleHandler := &formatterHandler{
				writer:    os.Stdout,
				formatter: generalFormatter,
				minLevel:  slog.LevelDebug,
			}
			multiHandler.AddHandler(consoleHandler)
		}
	case log.CustomDriver:
		logLogger := config.Get(channelPath + ".via").(log.Logger)
		logHook, err := logLogger.Handle(channelPath)
		if err != nil {
			return err
		}

		handler = &HookHandler{hook: logHook}
	default:
		return errors.LogDriverNotSupported.Args(channel)
	}

	multiHandler.AddHandler(handler)

	return nil
}

// formatterHandler for console output
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
