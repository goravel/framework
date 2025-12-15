package log

import (
	"context"
	"log/slog"
	"time"
)

// Record holds information about a log event.
// This interface provides a stable abstraction over log records,
// allowing the framework to remain independent of specific implementations.
// It is designed to be compatible with slog.Record while providing
// additional fields for Goravel's extended logging features.
type Record interface {
	// Time returns the timestamp of the log event.
	Time() time.Time

	// Level returns the severity level of the log event.
	Level() Level

	// Message returns the log message.
	Message() string

	// Context returns the context associated with the log event.
	Context() context.Context

	// Attrs calls f on each Attr in the Record.
	// If f returns false, iteration stops.
	Attrs(f func(slog.Attr) bool)

	// NumAttrs returns the number of attributes in the Record.
	NumAttrs() int

	// AddAttrs appends the given Attrs to the Record's list of Attrs.
	AddAttrs(attrs ...slog.Attr)

	// Add converts the args to Attrs as described in Logger.Log,
	// then appends the Attrs to the Record's list of Attrs.
	Add(args ...any)

	// Clone returns a copy of the Record.
	Clone() Record
}
