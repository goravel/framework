package logger

import (
	"context"

	"github.com/goravel/framework/support/carbon"
)

// Level log level
type Level int

const (
	Silent Level = iota + 1
	Error
	Warn
	Info
)

type Logger interface {
	Mode(Level) Logger
	Info(context.Context, string, ...any)
	Warn(context.Context, string, ...any)
	Error(context.Context, string, ...any)
	Trace(ctx context.Context, begin carbon.Carbon, sql string, rowsAffected int64, err error)
}
