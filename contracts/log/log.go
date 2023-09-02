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
	WithContext(ctx context.Context) Writer
	Writer
}

//go:generate mockery --name=Writer
type Writer interface {
	Debug(args ...any)
	Debugf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Warning(args ...any)
	Warningf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Panic(args ...any)
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
	Request(req http.Request) Writer
	// Response supplies a http.Response.
	Response(res http.Response) Writer
	// Tags add multiple tags, describing the feature returning an error.
	Tags(tags ...string) Writer
	// User sets the user associated with the log entry.
	User(user any) Writer
	// With adds key-value pairs to the context of the log entry
	With(data map[string]any) Writer
}

//go:generate mockery --name=Logger
type Logger interface {
	// Handle pass channel config path here
	Handle(channel string) (Hook, error)
}

//go:generate mockery --name=Hook
type Hook interface {
	// Levels monitoring level
	Levels() []Level
	// Fire execute logic when trigger
	Fire(Entry) error
}

//go:generate mockery --name=Entry
type Entry interface {
	Context() context.Context
	Level() Level
	Time() time.Time
	Message() string
}
