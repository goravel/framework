package log

import (
	"time"

	"github.com/sirupsen/logrus"
)

type Level logrus.Level

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

type Log interface {
	Testing(is bool) Log
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

type Logger interface {
	// Handle pass channel config path here
	Handle(channel string) (Hook, error)
}

type Hook interface {
	// Levels monitoring level
	Levels() []Level
	// Fire execute logic when trigger
	Fire(*Entry) error
}

type Entry struct {
	Level   Level
	Time    time.Time
	Message string
}

func (r *Entry) GetLevel() Level {
	return r.Level
}

func (r *Entry) GetTime() time.Time {
	return r.Time
}

func (r *Entry) GetMessage() string {
	return r.Message
}
