package log

import (
	"fmt"
	"log/slog"
	"strings"
)

// Level defines custom log levels for the logging system.
// We define custom levels that extend slog's built-in levels to support
// Panic and Fatal levels which are not part of the standard slog package.
type Level slog.Level

const (
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel Level = Level(slog.LevelDebug) // -4
	// InfoLevel level. General operational entries about what's going on inside the application.
	InfoLevel Level = Level(slog.LevelInfo) // 0
	// WarningLevel level. Non-critical entries that deserve eyes.
	WarningLevel Level = Level(slog.LevelWarn) // 4
	// ErrorLevel level. Used for errors that should definitely be noted.
	ErrorLevel Level = Level(slog.LevelError) // 8
	// FatalLevel level. Logs and then calls `os.Exit(1)`.
	FatalLevel Level = Level(slog.LevelError + 4) // 12
	// PanicLevel level. Highest level of severity. Logs and then calls panic.
	PanicLevel Level = Level(slog.LevelError + 8) // 16
)

// String converts the Level to a string. E.g. PanicLevel becomes "panic".
func (level Level) String() string {
	if b, err := level.MarshalText(); err == nil {
		return string(b)
	}
	return "unknown"
}

// ParseLevel takes a string level and returns the log level constant.
func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "panic":
		return PanicLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarningLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
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
	case DebugLevel:
		return []byte("debug"), nil
	case InfoLevel:
		return []byte("info"), nil
	case WarningLevel:
		return []byte("warning"), nil
	case ErrorLevel:
		return []byte("error"), nil
	case FatalLevel:
		return []byte("fatal"), nil
	case PanicLevel:
		return []byte("panic"), nil
	}

	return nil, fmt.Errorf("not a valid log level %d", level)
}

// SlogLevel converts the custom Level to slog.Level for use with slog handlers.
func (level Level) SlogLevel() slog.Level {
	return slog.Level(level)
}

// Level implements the slog.Leveler interface.
func (level Level) Level() slog.Level {
	return slog.Level(level)
}
