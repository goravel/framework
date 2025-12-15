package log

import (
	"context"
	"log/slog"
	"time"

	"github.com/goravel/framework/contracts/http"
)

const (
	StackDriver  = "stack"
	SingleDriver = "single"
	DailyDriver  = "daily"
	CustomDriver = "custom"
)

// Data is a map of key-value pairs for structured logging.
type Data map[string]any

// Log is the main logging interface that provides methods for writing logs
// to different channels and with different contexts.
type Log interface {
	// WithContext adds a context to the logger.
	WithContext(ctx context.Context) Writer
	// Channel returns a writer for a specific channel.
	Channel(channel string) Writer
	// Stack returns a writer for multiple channels.
	Stack(channels []string) Writer
	Writer
}

// Writer provides methods for writing log entries at different severity levels
// with support for structured data and metadata.
type Writer interface {
	// Debug logs a message at LevelDebug.
	Debug(args ...any)
	// Debugf is equivalent to Debug, but with support for fmt.Printf-style arguments.
	Debugf(format string, args ...any)
	// Info logs a message at LevelInfo.
	Info(args ...any)
	// Infof is equivalent to Info, but with support for fmt.Printf-style arguments.
	Infof(format string, args ...any)
	// Warning logs a message at LevelWarning.
	Warning(args ...any)
	// Warningf is equivalent to Warning, but with support for fmt.Printf-style arguments.
	Warningf(format string, args ...any)
	// Error logs a message at LevelError.
	Error(args ...any)
	// Errorf is equivalent to Error, but with support for fmt.Printf-style arguments.
	Errorf(format string, args ...any)
	// Fatal logs a message at LevelFatal.
	Fatal(args ...any)
	// Fatalf is equivalent to Fatal, but with support for fmt.Printf-style arguments.
	Fatalf(format string, args ...any)
	// Panic logs a message at LevelPanic.
	Panic(args ...any)
	// Panicf is equivalent to Panic, but with support for fmt.Printf-style arguments.
	Panicf(format string, args ...any)
	// Code sets a code or slug that describes the error.
	// Error messages are intended to be read by humans, but such code is expected to
	// be read by machines and even transported over different services.
	Code(code string) Writer
	// Hint sets a hint for faster debugging.
	Hint(hint string) Writer
	// In sets the feature category or domain in which the log entry is relevant.
	In(domain string) Writer
	// Owner sets the name/email of the colleague/team responsible for handling this error.
	// Useful for alerting purpose.
	Owner(owner any) Writer
	// Request supplies an http.Request.
	Request(req http.ContextRequest) Writer
	// Response supplies an http.Response.
	Response(res http.ContextResponse) Writer
	// Tags adds multiple tags, describing the feature returning an error.
	Tags(tags ...string) Writer
	// User sets the user associated with the log entry.
	User(user any) Writer
	// With adds key-value pairs to the context of the log entry.
	With(data map[string]any) Writer
	// WithTrace adds a stack trace to the log entry.
	WithTrace() Writer
}

// ToSlogHandler converts a framework Handler to a slog.Handler.
// This function allows framework handlers to be used with standard slog loggers.
func ToSlogHandler(h Handler) slog.Handler {
	return &slogHandlerAdapter{handler: h}
}

// FromSlogHandler converts a slog.Handler to a framework Handler.
// This function allows standard slog handlers to be used within the framework.
func FromSlogHandler(h slog.Handler) Handler {
	return &frameworkHandlerAdapter{handler: h}
}

// slogHandlerAdapter wraps a framework Handler to implement slog.Handler.
type slogHandlerAdapter struct {
	handler Handler
}

func (a *slogHandlerAdapter) Enabled(ctx context.Context, level slog.Level) bool {
	return a.handler.Enabled(ctx, Level(level))
}

func (a *slogHandlerAdapter) Handle(ctx context.Context, r slog.Record) error {
	return a.handler.Handle(ctx, &slogRecordAdapter{record: &r})
}

func (a *slogHandlerAdapter) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &slogHandlerAdapter{handler: a.handler.WithAttrs(attrs)}
}

func (a *slogHandlerAdapter) WithGroup(name string) slog.Handler {
	return &slogHandlerAdapter{handler: a.handler.WithGroup(name)}
}

// frameworkHandlerAdapter wraps a slog.Handler to implement framework Handler.
type frameworkHandlerAdapter struct {
	handler slog.Handler
}

func (a *frameworkHandlerAdapter) Enabled(ctx context.Context, level Level) bool {
	return a.handler.Enabled(ctx, level.SlogLevel())
}

func (a *frameworkHandlerAdapter) Handle(ctx context.Context, r Record) error {
	// Convert framework Record to slog.Record
	slogRecord := slog.NewRecord(r.Time(), r.Level().SlogLevel(), r.Message(), 0)
	r.Attrs(func(attr slog.Attr) bool {
		slogRecord.AddAttrs(attr)
		return true
	})
	return a.handler.Handle(ctx, slogRecord)
}

func (a *frameworkHandlerAdapter) WithAttrs(attrs []slog.Attr) Handler {
	return &frameworkHandlerAdapter{handler: a.handler.WithAttrs(attrs)}
}

func (a *frameworkHandlerAdapter) WithGroup(name string) Handler {
	return &frameworkHandlerAdapter{handler: a.handler.WithGroup(name)}
}

// slogRecordAdapter wraps a slog.Record to implement framework Record.
type slogRecordAdapter struct {
	record *slog.Record
}

func (a *slogRecordAdapter) Time() time.Time {
	return a.record.Time
}

func (a *slogRecordAdapter) Level() Level {
	return Level(a.record.Level)
}

func (a *slogRecordAdapter) Message() string {
	return a.record.Message
}

func (a *slogRecordAdapter) Context() context.Context {
	return context.Background()
}

func (a *slogRecordAdapter) Attrs(f func(slog.Attr) bool) {
	a.record.Attrs(f)
}

func (a *slogRecordAdapter) NumAttrs() int {
	return a.record.NumAttrs()
}

func (a *slogRecordAdapter) AddAttrs(attrs ...slog.Attr) {
	a.record.AddAttrs(attrs...)
}

func (a *slogRecordAdapter) Add(args ...any) {
	a.record.Add(args...)
}

func (a *slogRecordAdapter) Clone() Record {
	cloned := a.record.Clone()
	return &slogRecordAdapter{record: &cloned}
}
