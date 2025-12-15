package log

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"

	"github.com/rotisserie/eris"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
)

// Writer implements the log.Writer interface using slog.
// The Writer is designed to be reused - all metadata (code, hint, domain, etc.)
// is reset after each log operation to ensure clean state for the next log entry.
type Writer struct {
	ctx          context.Context
	logger       *slog.Logger
	owner        any
	request      http.ContextRequest
	response     http.ContextResponse
	user         any
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
		ctx:          ctx,
		logger:       logger,
		code:         "",
		domain:       "",
		hint:         "",
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
	r.log(log.DebugLevel, fmt.Sprint(args...))
}

func (r *Writer) Debugf(format string, args ...any) {
	r.log(log.DebugLevel, fmt.Sprintf(format, args...))
}

func (r *Writer) Info(args ...any) {
	r.log(log.InfoLevel, fmt.Sprint(args...))
}

func (r *Writer) Infof(format string, args ...any) {
	r.log(log.InfoLevel, fmt.Sprintf(format, args...))
}

func (r *Writer) Warning(args ...any) {
	r.log(log.WarningLevel, fmt.Sprint(args...))
}

func (r *Writer) Warningf(format string, args ...any) {
	r.log(log.WarningLevel, fmt.Sprintf(format, args...))
}

func (r *Writer) Error(args ...any) {
	msg := fmt.Sprint(args...)
	r.withStackTrace(msg)
	r.log(log.ErrorLevel, msg)
}

func (r *Writer) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	r.withStackTrace(msg)
	r.log(log.ErrorLevel, msg)
}

func (r *Writer) Fatal(args ...any) {
	msg := fmt.Sprint(args...)
	r.withStackTrace(msg)
	r.log(log.FatalLevel, msg)
	os.Exit(1)
}

func (r *Writer) Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	r.withStackTrace(msg)
	r.log(log.FatalLevel, msg)
	os.Exit(1)
}

func (r *Writer) Panic(args ...any) {
	msg := fmt.Sprint(args...)
	r.withStackTrace(msg)
	r.log(log.PanicLevel, msg)
	panic(msg)
}

func (r *Writer) Panicf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	r.withStackTrace(msg)
	r.log(log.PanicLevel, msg)
	panic(msg)
}

// Code set a code or slug that describes the error.
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

func (r *Writer) log(level log.Level, msg string) {
	attrs := r.toAttrs()
	r.logger.LogAttrs(r.ctx, level.Level(), msg, attrs...)
	r.resetAll()
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

// resetAll resets all properties for a new log entry.
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

// toAttrs converts the writer state to slog attributes.
func (r *Writer) toAttrs() []slog.Attr {
	var attrs []slog.Attr

	// Build root group with all metadata
	rootAttrs := []any{}

	if code := r.code; code != "" {
		rootAttrs = append(rootAttrs, slog.String("code", code))
	}
	if ctx := r.ctx; ctx != nil {
		values := make(map[any]any)
		getContextValues(ctx, values)
		if len(values) > 0 {
			rootAttrs = append(rootAttrs, slog.Any("context", values))
		}
	}
	if domain := r.domain; domain != "" {
		rootAttrs = append(rootAttrs, slog.String("domain", domain))
	}
	if hint := r.hint; hint != "" {
		rootAttrs = append(rootAttrs, slog.String("hint", hint))
	}
	if message := r.message; message != "" {
		rootAttrs = append(rootAttrs, slog.String("message", message))
	}
	if owner := r.owner; owner != nil {
		rootAttrs = append(rootAttrs, slog.Any("owner", owner))
	}
	if req := r.request; req != nil {
		rootAttrs = append(rootAttrs, slog.Any("request", map[string]any{
			"method": req.Method(),
			"uri":    req.FullUrl(),
			"header": req.Headers(),
			"body":   req.All(),
		}))
	}
	if res := r.response; res != nil {
		rootAttrs = append(rootAttrs, slog.Any("response", map[string]any{
			"status": res.Origin().Status(),
			"header": res.Origin().Header(),
			"body":   res.Origin().Body(),
			"size":   res.Origin().Size(),
		}))
	}
	if stacktrace := r.stacktrace; stacktrace != nil || r.stackEnabled {
		rootAttrs = append(rootAttrs, slog.Any("stacktrace", stacktrace))
	}
	if tags := r.tags; len(tags) > 0 {
		rootAttrs = append(rootAttrs, slog.Any("tags", tags))
	}
	if r.user != nil {
		rootAttrs = append(rootAttrs, slog.Any("user", r.user))
	}
	if with := r.with; len(with) > 0 {
		rootAttrs = append(rootAttrs, slog.Any("with", with))
	}

	if len(rootAttrs) > 0 {
		attrs = append(attrs, slog.Group("root", rootAttrs...))
	}

	return attrs
}
