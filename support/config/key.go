package config

import "fmt"

// Key represents a configuration key that may contain format placeholders.
// Use String() for keys without placeholders, With() for keys with placeholders.
//
// Example:
//
//	const (
//	    ConfigServiceName configKey = "app.name"                    // no placeholder
//	    ConfigExporter    configKey = "telemetry.exporters.%s.driver" // with placeholder
//	)
//
//	cfg.GetString(ConfigServiceName.String())
//	cfg.GetString(ConfigExporter.With("otlp"))
type Key string

// String returns the key as a string. Use for keys without placeholders.
func (k Key) String() string {
	return string(k)
}

// With formats the key with the provided arguments using fmt.Sprintf.
// Use for keys containing format placeholders like %s, %d, etc.
func (k Key) With(args ...any) string {
	return fmt.Sprintf(string(k), args...)
}
