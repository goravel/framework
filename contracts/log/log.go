package log

import (
	"context"
	"time"
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
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warning(args ...interface{})
	Warningf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
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
	// DEPRECATED: use Level()
	GetLevel() Level
	// DEPRECATED: use Time()
	GetTime() time.Time
	// DEPRECATED: use Message()
	GetMessage() string
}
