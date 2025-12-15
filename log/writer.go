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
	"github.com/goravel/framework/log/logger"
)

type Writer struct {
	owner        any
	request      http.ContextRequest
	response     http.ContextResponse
	user         any
	logger       *slog.Logger
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

func NewWriter(logger *slog.Logger, ctx context.Context) log.Writer {
	return &Writer{
		code:         "",
		domain:       "",
		hint:         "",
		logger:       logger,
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
	r.logger.LogAttrs(r.ctx, slog.LevelDebug, fmt.Sprint(args...), r.toAttrs()...)
	r.resetAll()
}

func (r *Writer) Debugf(format string, args ...any) {
	r.logger.LogAttrs(r.ctx, slog.LevelDebug, fmt.Sprintf(format, args...), r.toAttrs()...)
	r.resetAll()
}

func (r *Writer) Info(args ...any) {
	r.logger.LogAttrs(r.ctx, slog.LevelInfo, fmt.Sprint(args...), r.toAttrs()...)
	r.resetAll()
}

func (r *Writer) Infof(format string, args ...any) {
	r.logger.LogAttrs(r.ctx, slog.LevelInfo, fmt.Sprintf(format, args...), r.toAttrs()...)
	r.resetAll()
}

func (r *Writer) Warning(args ...any) {
	r.logger.LogAttrs(r.ctx, slog.LevelWarn, fmt.Sprint(args...), r.toAttrs()...)
	r.resetAll()
}

func (r *Writer) Warningf(format string, args ...any) {
	r.logger.LogAttrs(r.ctx, slog.LevelWarn, fmt.Sprintf(format, args...), r.toAttrs()...)
	r.resetAll()
}

func (r *Writer) Error(args ...any) {
	r.withStackTrace(fmt.Sprint(args...))
	r.logger.LogAttrs(r.ctx, slog.LevelError, fmt.Sprint(args...), r.toAttrs()...)
	r.resetAll()
}

func (r *Writer) Errorf(format string, args ...any) {
	r.withStackTrace(fmt.Sprintf(format, args...))
	r.logger.LogAttrs(r.ctx, slog.LevelError, fmt.Sprintf(format, args...), r.toAttrs()...)
	r.resetAll()
}

func (r *Writer) Fatal(args ...any) {
	r.withStackTrace(fmt.Sprint(args...))
	r.logger.LogAttrs(r.ctx, slog.Level(log.LevelFatal), fmt.Sprint(args...), r.toAttrs()...)
	r.resetAll()
	os.Exit(1)
}

func (r *Writer) Fatalf(format string, args ...any) {
	r.withStackTrace(fmt.Sprintf(format, args...))
	r.logger.LogAttrs(r.ctx, slog.Level(log.LevelFatal), fmt.Sprintf(format, args...), r.toAttrs()...)
	r.resetAll()
	os.Exit(1)
}

func (r *Writer) Panic(args ...any) {
	r.withStackTrace(fmt.Sprint(args...))
	msg := fmt.Sprint(args...)
	r.logger.LogAttrs(r.ctx, slog.Level(log.LevelPanic), msg, r.toAttrs()...)
	r.resetAll()
	panic(msg)
}

func (r *Writer) Panicf(format string, args ...any) {
	r.withStackTrace(fmt.Sprintf(format, args...))
	msg := fmt.Sprintf(format, args...)
	r.logger.LogAttrs(r.ctx, slog.Level(log.LevelPanic), msg, r.toAttrs()...)
	r.resetAll()
	panic(msg)
}

// Code sets a code or slug that describes the error.
// Error messages are intended to be read by humans, but such code is expected to
// be read by machines and even transported over different services.
func (r *Writer) Code(code string) log.Writer {
	r.code = code
	return r
}

// Hint sets a hint for faster debugging.
func (r *Writer) Hint(hint string) log.Writer {
	r.hint = hint

	return r
}

// In sets the feature category or domain in which the log entry is relevant.
func (r *Writer) In(domain string) log.Writer {
	r.domain = domain

	return r
}

// Owner sets the name/email of the colleague/team responsible for handling this error.
// Useful for alerting purpose.
func (r *Writer) Owner(owner any) log.Writer {
	r.owner = owner

	return r
}

// Request supplies an http.Request.
func (r *Writer) Request(req http.ContextRequest) log.Writer {
	r.request = req

	return r
}

// Response supplies an http.Response.
func (r *Writer) Response(res http.ContextResponse) log.Writer {
	r.response = res

	return r
}

// Tags adds multiple tags, describing the feature returning an error.
func (r *Writer) Tags(tags ...string) log.Writer {
	r.tags = append(r.tags, tags...)

	return r
}

// User sets the user associated with the log entry.
func (r *Writer) User(user any) log.Writer {
	r.user = user
	return r
}

// With adds key-value pairs to the context of the log entry.
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

// toAttrs returns the attributes for the log entry.
func (r *Writer) toAttrs() []slog.Attr {
	var attrs []slog.Attr

	root := map[string]any{}

	if code := r.code; code != "" {
		root["code"] = code
	}
	if ctx := r.ctx; ctx != nil {
		values := make(map[any]any)
		getContextValues(ctx, values)
		if len(values) > 0 {
			root["context"] = values
		}
	}
	if domain := r.domain; domain != "" {
		root["domain"] = domain
	}
	if hint := r.hint; hint != "" {
		root["hint"] = hint
	}
	if message := r.message; message != "" {
		root["message"] = message
	}
	if owner := r.owner; owner != nil {
		root["owner"] = owner
	}
	if req := r.request; req != nil {
		root["request"] = map[string]any{
			"method": req.Method(),
			"uri":    req.FullUrl(),
			"header": req.Headers(),
			"body":   req.All(),
		}
	}
	if res := r.response; res != nil {
		root["response"] = map[string]any{
			"status": res.Origin().Status(),
			"header": res.Origin().Header(),
			"body":   res.Origin().Body(),
			"size":   res.Origin().Size(),
		}
	}
	if stacktrace := r.stacktrace; stacktrace != nil || r.stackEnabled {
		root["stacktrace"] = stacktrace
	}
	if tags := r.tags; len(tags) > 0 {
		root["tags"] = tags
	}
	if r.user != nil {
		root["user"] = r.user
	}
	if with := r.with; len(with) > 0 {
		root["with"] = with
	}

	if len(root) > 0 {
		attrs = append(attrs, slog.Any("root", root))
	}

	return attrs
}

// getHandlers returns the slog handlers for the specified channel.
func getHandlers(config config.Config, json foundation.Json, channel string) ([]slog.Handler, error) {
	var handlers []slog.Handler

	channelPath := "logging.channels." + channel
	driver := config.GetString(channelPath + ".driver")

	switch driver {
	case log.StackDriver:
		for _, stackChannel := range config.Get(channelPath + ".channels").([]string) {
			if stackChannel == channel {
				return nil, errors.LogDriverCircularReference.Args("stack")
			}

			h, err := getHandlers(config, json, stackChannel)
			if err != nil {
				return nil, err
			}
			handlers = append(handlers, h...)
		}

		return handlers, nil
	case log.SingleDriver:
		logLogger := logger.NewSingle(config, json)
		handler, err := logLogger.Handle(channelPath)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, handler)
	case log.DailyDriver:
		logLogger := logger.NewDaily(config, json)
		handler, err := logLogger.Handle(channelPath)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, handler)
	case log.CustomDriver:
		logLogger := config.Get(channelPath + ".via").(log.Logger)
		handler, err := logLogger.Handle(channelPath)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, log.ToSlogHandler(handler))
	default:
		return nil, errors.LogDriverNotSupported.Args(channel)
	}

	return handlers, nil
}
