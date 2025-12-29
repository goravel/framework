package log

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"

	"github.com/dromara/carbon/v2"
	"github.com/rotisserie/eris"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
)

type Writer struct {
	logger *slog.Logger
	ctx    context.Context
	entry  *Entry // nil for base writer, only set when fluent methods are called
}

func NewWriter(logger *slog.Logger, ctx context.Context) log.Writer {
	return &Writer{
		logger: logger,
		ctx:    ctx,
		entry:  nil,
	}
}

func (w *Writer) Debug(args ...any) {
	w.log(log.LevelDebug, fmt.Sprint(args...))
}

func (w *Writer) Debugf(format string, args ...any) {
	w.log(log.LevelDebug, fmt.Sprintf(format, args...))
}

func (w *Writer) Info(args ...any) {
	w.log(log.LevelInfo, fmt.Sprint(args...))
}

func (w *Writer) Infof(format string, args ...any) {
	w.log(log.LevelInfo, fmt.Sprintf(format, args...))
}

func (w *Writer) Warning(args ...any) {
	w.log(log.LevelWarning, fmt.Sprint(args...))
}

func (w *Writer) Warningf(format string, args ...any) {
	w.log(log.LevelWarning, fmt.Sprintf(format, args...))
}

func (w *Writer) Error(args ...any) {
	msg := fmt.Sprint(args...)
	nw := w.ensureEntry()
	nw.withStackTrace(msg)
	nw.log(log.LevelError, msg)
}

func (w *Writer) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	nw := w.ensureEntry()
	nw.withStackTrace(msg)
	nw.log(log.LevelError, msg)
}

func (w *Writer) Fatal(args ...any) {
	msg := fmt.Sprint(args...)
	nw := w.ensureEntry()
	nw.withStackTrace(msg)
	nw.log(log.LevelFatal, msg)
	os.Exit(1)
}

func (w *Writer) Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	nw := w.ensureEntry()
	nw.withStackTrace(msg)
	nw.log(log.LevelFatal, msg)
	os.Exit(1)
}

func (w *Writer) Panic(args ...any) {
	msg := fmt.Sprint(args...)
	nw := w.ensureEntry()
	nw.withStackTrace(msg)
	nw.log(log.LevelPanic, msg)
	panic(msg)
}

func (w *Writer) Panicf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	nw := w.ensureEntry()
	nw.withStackTrace(msg)
	nw.log(log.LevelPanic, msg)
	panic(msg)
}

// Code set a code or slug that describes the error.
func (w *Writer) Code(code string) log.Writer {
	nw := w.clone()
	nw.entry.code = code
	return nw
}

// Hint set a hint for faster debugging.
func (w *Writer) Hint(hint string) log.Writer {
	nw := w.clone()
	nw.entry.hint = hint
	return nw
}

// In sets the feature category or domain in which the log entry is relevant.
func (w *Writer) In(domain string) log.Writer {
	nw := w.clone()
	nw.entry.domain = domain
	return nw
}

// Owner set the name/email of the colleague/team responsible for handling this error.
func (w *Writer) Owner(owner any) log.Writer {
	nw := w.clone()
	nw.entry.owner = owner
	return nw
}

// Request supplies a http.Request.
func (w *Writer) Request(req http.ContextRequest) log.Writer {
	nw := w.clone()
	if req != nil {
		nw.entry.request = map[string]any{
			"method": req.Method(),
			"uri":    req.FullUrl(),
			"header": req.Headers(),
			"body":   req.All(),
		}
	}
	return nw
}

// Response supplies a http.Response.
func (w *Writer) Response(res http.ContextResponse) log.Writer {
	nw := w.clone()
	if res != nil {
		nw.entry.response = map[string]any{
			"status": res.Origin().Status(),
			"header": res.Origin().Header(),
			"body":   res.Origin().Body(),
			"size":   res.Origin().Size(),
		}
	}
	return nw
}

// Tags add multiple tags, describing the feature returning an error.
func (w *Writer) Tags(tags ...string) log.Writer {
	nw := w.clone()
	nw.entry.tags = append(nw.entry.tags, tags...)
	return nw
}

// User sets the user associated with the log entry.
func (w *Writer) User(user any) log.Writer {
	nw := w.clone()
	nw.entry.user = user
	return nw
}

// With adds key-value pairs to the context of the log entry.
func (w *Writer) With(data map[string]any) log.Writer {
	nw := w.clone()
	maps.Copy(nw.entry.with, data)
	return nw
}

// WithTrace adds a stack trace to the log entry.
func (w *Writer) WithTrace() log.Writer {
	nw := w.clone()
	nw.withStackTrace("")
	return nw
}

func (w *Writer) log(level log.Level, msg string) {
	entry := w.entry
	if entry == nil {
		// For direct log calls without fluent methods, acquire a fresh entry
		entry = acquireEntry()
		entry.ctx = w.ctx
	}

	entry.time = carbon.Now().StdTime()
	entry.message = msg
	entry.level = level

	_ = w.logger.Handler().Handle(entry.ctx, entry.ToSlogRecord())
	releaseEntry(entry)
}

func (w *Writer) withStackTrace(message string) {
	erisNew := eris.New(message)
	if erisNew == nil {
		return
	}

	format := eris.NewDefaultJSONFormat(eris.FormatOptions{
		InvertOutput: true,
		WithTrace:    true,
		InvertTrace:  true,
	})
	w.entry.stacktrace = eris.ToCustomJSON(erisNew, format)
}

func (w *Writer) getEntry() *Entry {
	if w.entry == nil {
		entry := acquireEntry()
		entry.ctx = w.ctx
		return entry
	}
	return w.entry
}

func (w *Writer) clone() *Writer {
	entry := w.getEntry()
	return &Writer{
		logger: w.logger,
		ctx:    w.ctx,
		entry:  entry,
	}
}

func (w *Writer) ensureEntry() *Writer {
	if w.entry != nil {
		return w
	}
	return w.clone()
}
