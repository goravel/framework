package log

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"time"

	"github.com/rotisserie/eris"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
)

// Writer implements the log.Writer interface using slog.
// The Writer is designed to be reused - all metadata (code, hint, domain, etc.)
// is reset after each log operation to ensure clean state for the next log entry.
type Writer struct {
	logger *slog.Logger
	entry  *Entry
}

func NewWriter(logger *slog.Logger, ctx context.Context) log.Writer {
	entry := acquireEntry()
	entry.time = time.Now()
	entry.ctx = ctx
	return &Writer{
		logger: logger,
		entry:  entry,
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
	w.withStackTrace(msg)
	w.log(log.LevelError, msg)
}

func (w *Writer) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	w.withStackTrace(msg)
	w.log(log.LevelError, msg)
}

func (w *Writer) Fatal(args ...any) {
	msg := fmt.Sprint(args...)
	w.withStackTrace(msg)
	w.log(log.LevelFatal, msg)
	os.Exit(1)
}

func (w *Writer) Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	w.withStackTrace(msg)
	w.log(log.LevelFatal, msg)
	os.Exit(1)
}

func (w *Writer) Panic(args ...any) {
	msg := fmt.Sprint(args...)
	w.withStackTrace(msg)
	w.log(log.LevelPanic, msg)
	panic(msg)
}

func (w *Writer) Panicf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	w.withStackTrace(msg)
	w.log(log.LevelPanic, msg)
	panic(msg)
}

// Code set a code or slug that describes the error.
func (w *Writer) Code(code string) log.Writer {
	w.entry.code = code
	return w
}

// Hint set a hint for faster debugging.
func (w *Writer) Hint(hint string) log.Writer {
	w.entry.hint = hint
	return w
}

// In sets the feature category or domain in which the log entry is relevant.
func (w *Writer) In(domain string) log.Writer {
	w.entry.domain = domain
	return w
}

// Owner set the name/email of the colleague/team responsible for handling this error.
func (w *Writer) Owner(owner any) log.Writer {
	w.entry.owner = owner
	return w
}

// Request supplies a http.Request.
func (w *Writer) Request(req http.ContextRequest) log.Writer {
	if req != nil {
		w.entry.request = map[string]any{
			"method": req.Method(),
			"uri":    req.FullUrl(),
			"header": req.Headers(),
			"body":   req.All(),
		}
	}
	return w
}

// Response supplies a http.Response.
func (w *Writer) Response(res http.ContextResponse) log.Writer {
	if res != nil {
		w.entry.response = map[string]any{
			"status": res.Origin().Status(),
			"header": res.Origin().Header(),
			"body":   res.Origin().Body(),
			"size":   res.Origin().Size(),
		}
	}
	return w
}

// Tags add multiple tags, describing the feature returning an error.
func (w *Writer) Tags(tags ...string) log.Writer {
	w.entry.tags = append(w.entry.tags, tags...)
	return w
}

// User sets the user associated with the log entry.
func (w *Writer) User(user any) log.Writer {
	w.entry.user = user
	return w
}

// With adds key-value pairs to the context of the log entry.
func (w *Writer) With(data map[string]any) log.Writer {
	maps.Copy(w.entry.with, data)
	return w
}

// WithTrace adds a stack trace to the log entry.
func (w *Writer) WithTrace() log.Writer {
	w.withStackTrace("")
	return w
}

func (w *Writer) log(level log.Level, msg string) {
	defer releaseEntry(w.entry)

	w.entry.message = msg
	w.entry.level = level

	_ = w.logger.Handler().Handle(w.entry.ctx, w.entry.ToSlogRecord())
}

func (w *Writer) withStackTrace(message string) {
	erisNew := eris.New(message)
	if erisNew == nil {
		return
	}

	w.entry.message = erisNew.Error()
	format := eris.NewDefaultJSONFormat(eris.FormatOptions{
		InvertOutput: true,
		WithTrace:    true,
		InvertTrace:  true,
	})
	w.entry.stacktrace = eris.ToCustomJSON(erisNew, format)
}
