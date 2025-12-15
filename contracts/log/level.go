package log

import (
	"fmt"
	"log/slog"
	"strings"
)

// Level represents a logging severity level.
// The values are designed to be compatible with slog levels, with custom extensions
// for Fatal and Panic levels. Higher values indicate more severe log events.
type Level int

const (
	// LevelDebug is the debug level, used for detailed troubleshooting information.
	LevelDebug Level = Level(slog.LevelDebug) // -4
	// LevelInfo is the info level, used for general operational information.
	LevelInfo Level = Level(slog.LevelInfo) // 0
	// LevelWarning is the warning level, used for potentially harmful situations.
	LevelWarning Level = Level(slog.LevelWarn) // 4
	// LevelError is the error level, used for error events.
	LevelError Level = Level(slog.LevelError) // 8
	// LevelFatal is the fatal level, used for severe errors that cause application exit.
	LevelFatal Level = 12
	// LevelPanic is the panic level, used for severe errors that cause panic.
	LevelPanic Level = 16
)

// String returns the string representation of the Level.
// E.g., LevelPanic becomes "panic".
func (level Level) String() string {
	if b, err := level.MarshalText(); err == nil {
		return string(b)
	}
	return "unknown"
}

// ParseLevel takes a string level and returns the corresponding Level constant.
func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "panic":
		return LevelPanic, nil
	case "fatal":
		return LevelFatal, nil
	case "error":
		return LevelError, nil
	case "warn", "warning":
		return LevelWarning, nil
	case "info":
		return LevelInfo, nil
	case "debug":
		return LevelDebug, nil
	}

	var l Level
	return l, fmt.Errorf("not a valid log Level: %q", lvl)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (level *Level) UnmarshalText(text []byte) error {
	l, err := ParseLevel(string(text))
	if err != nil {
		return err
	}

	*level = l
	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (level Level) MarshalText() ([]byte, error) {
	switch level {
	case LevelDebug:
		return []byte("debug"), nil
	case LevelInfo:
		return []byte("info"), nil
	case LevelWarning:
		return []byte("warning"), nil
	case LevelError:
		return []byte("error"), nil
	case LevelFatal:
		return []byte("fatal"), nil
	case LevelPanic:
		return []byte("panic"), nil
	}

	return nil, fmt.Errorf("not a valid log level %d", level)
}

// SlogLevel returns the corresponding slog.Level for this Level.
// For custom levels (Fatal, Panic), it returns the closest slog equivalent
// with appropriate offset.
func (level Level) SlogLevel() slog.Level {
	return slog.Level(level)
}
