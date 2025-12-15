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
	// LevelDebug level. Usually only enabled when debugging. Very verbose logging.
	LevelDebug Level = Level(slog.LevelDebug) // -4
	// LevelInfo level. General operational entries about what's going on inside the application.
	LevelInfo Level = Level(slog.LevelInfo) // 0
	// LevelWarning level. Non-critical entries that deserve eyes.
	LevelWarning Level = Level(slog.LevelWarn) // 4
	// LevelError level. Used for errors that should definitely be noted.
	LevelError Level = Level(slog.LevelError) // 8
	// LevelFatal level. Logs and then calls `os.Exit(1)`.
	LevelFatal Level = Level(slog.LevelError + 4) // 12
	// LevelPanic level. Highest level of severity. Logs and then calls panic.
	LevelPanic Level = Level(slog.LevelError + 8) // 16
)

// String converts the Level to a string. E.g. LevelPanic becomes "panic".
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

// Level implements the slog.Leveler interface.
func (level Level) Level() slog.Level {
	return slog.Level(level)
}
