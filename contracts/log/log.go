package log

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/http"
)

const (
	StackDriver  = "stack"
	SingleDriver = "single"
	DailyDriver  = "daily"
	CustomDriver = "custom"
)

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarningLevel
	InfoLevel
	DebugLevel
)

//go:generate mockery --name=Log
type Log interface {
	// WithContext adds a context to the logger.
	WithContext(ctx context.Context) Writer
	Writer
}

//go:generate mockery --name=Writer
type Writer interface {
	// Debug logs a message at DebugLevel.
	Debug(args ...any)
	// Debugf is equivalent to Debug, but with support for fmt.Printf-style arguments.
	Debugf(format string, args ...any)
	// Info logs a message at InfoLevel.
	Info(args ...any)
	// Infof is equivalent to Info, but with support for fmt.Printf-style arguments.
	Infof(format string, args ...any)
	// Warning logs a message at WarningLevel.
	Warning(args ...any)
	// Warningf is equivalent to Warning, but with support for fmt.Printf-style arguments.
	Warningf(format string, args ...any)
	// Error logs a message at ErrorLevel.
	Error(args ...any)
	// Errorf is equivalent to Error, but with support for fmt.Printf-style arguments.
	Errorf(format string, args ...any)
	// Fatal logs a message at FatalLevel.
	Fatal(args ...any)
	// Fatalf is equivalent to Fatal, but with support for fmt.Printf-style arguments.
	Fatalf(format string, args ...any)
	// Panic logs a message at PanicLevel.
	Panic(args ...any)
	// Panicf is equivalent to Panic, but with support for fmt.Printf-style arguments.
	Panicf(format string, args ...any)
	// Code set a code or slug that describes the error.
	// Error messages are intended to be read by humans, but such code is expected to
	// be read by machines and even transported over different services.
	Code(code string) Writer
	// Hint set a hint for faster debugging.
	Hint(hint string) Writer
	// In sets the feature category or domain in which the log entry is relevant.
	In(domain string) Writer
	// Owner set the name/email of the colleague/team responsible for handling this error.
	// Useful for alerting purpose.
	Owner(owner any) Writer
	// Request supplies a http.Request.
	Request(req http.ContextRequest) Writer
	// Response supplies a http.Response.
	Response(res http.ContextResponse) Writer
	// Tags add multiple tags, describing the feature returning an error.
	Tags(tags ...string) Writer
	// User sets the user associated with the log entry.
	User(user any) Writer
	// With adds key-value pairs to the context of the log entry
	With(data map[string]any) Writer
}

//go:generate mockery --name=Logger
type Logger interface {
	// Handle pass a channel config path here
	Handle(channel string) (Hook, error)
}

//go:generate mockery --name=Hook
type Hook interface {
	// Levels monitoring level
	Levels() []Level
	// Fire executes logic when trigger
	Fire(Entry) error
}

//go:generate mockery --name=Entry
type Entry interface {
	// Context returns the context of the entry.
	Context() context.Context
	// Level returns the level of the entry.
	Level() Level
	// Time returns the timestamp of the entry.
	Time() time.Time
	// Message returns the message of the entry.
	Message() string
}
