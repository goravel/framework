package log

import (
	"fmt"
	"log/slog"
	"strings"
)

type Level int

// Convert the Level to a string. E.g. PanicLevel becomes "panic".
func (level Level) String() string {
	if b, err := level.MarshalText(); err == nil {
		return string(b)
	} else {
		return "unknown"
	}
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

// ToSlog converts our Level to slog.Level
func (level Level) ToSlog() slog.Level {
	// Map our levels to slog levels
	// Note: slog.Level values are: Debug=-4, Info=0, Warn=4, Error=8
	// Our levels are: Panic=0, Fatal=1, Error=2, Warning=3, Info=4, Debug=5
	// We use custom levels for Panic (16) and Fatal (12) to distinguish them
	switch level {
	case PanicLevel:
		return slog.Level(16) // Higher than Fatal
	case FatalLevel:
		return slog.Level(12) // Higher than Error
	case ErrorLevel:
		return slog.LevelError
	case WarningLevel:
		return slog.LevelWarn
	case InfoLevel:
		return slog.LevelInfo
	case DebugLevel:
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}

// FromSlog converts slog.Level to our Level
func FromSlog(level slog.Level) Level {
	switch {
	case level >= 16:
		return PanicLevel
	case level >= 12:
		return FatalLevel
	case level >= slog.LevelError:
		return ErrorLevel
	case level >= slog.LevelWarn:
		return WarningLevel
	case level >= slog.LevelInfo:
		return InfoLevel
	default:
		return DebugLevel
	}
}
