package log

import (
	"context"
	"log/slog"
)

// Handler is an interface that mirrors slog.Handler for handling log records.
// This abstraction allows the framework to remain independent of the specific
// logging implementation while following slog's explicit Handler pipeline model.
//
// In this model:
//   - Handle is equivalent to logrus's Fire method
//   - Enabled replaces the Levels concept
//   - Records are immutable structures (safer than mutable entries)
//
// The pipeline pattern is: Logger → Handler → Handler → Handler → Output
type Handler interface {
	// Enabled reports whether the handler handles records at the given level.
	// The handler ignores records whose level is lower.
	Enabled(context.Context, Level) bool

	// Handle handles the Record.
	// It will only be called when Enabled returns true.
	// Handle methods that produce output should observe the following rules:
	//   - If r.Time is the zero time, ignore the time.
	//   - If an Attr's key and value are both the zero value, ignore the Attr.
	Handle(context.Context, Record) error

	// WithAttrs returns a new Handler whose attributes consist of
	// both the receiver's attributes and the arguments.
	WithAttrs(attrs []slog.Attr) Handler

	// WithGroup returns a new Handler with the given group appended to
	// the receiver's existing groups.
	WithGroup(name string) Handler
}

// Logger is an interface for creating log handlers for a specific channel.
// It returns a Handler instead of the deprecated Hook interface.
type Logger interface {
	// Handle creates and returns a Handler for the specified channel configuration path.
	Handle(channel string) (Handler, error)
}
