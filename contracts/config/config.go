package config

import "time"

type Config interface {
	// Env get config from env.
	Env(envName string, defaultValue ...any) any
	// EnvString get string value from env with optional default.
	EnvString(envName string, defaultValue ...string) string
	// EnvBool get bool value from env with optional default.
	EnvBool(envName string, defaultValue ...bool) bool
	// Add config to application.
	Add(name string, configuration any)
	// Get config from application.
	Get(path string, defaultValue ...any) any
	// GetString get string type config from application.
	GetString(path string, defaultValue ...string) string
	// GetInt get int type config from application.
	GetInt(path string, defaultValue ...int) int
	// GetInt64 get int64 type config from application.
	GetInt64(path string, defaultValue ...int64) int64
	// GetFloat64 get float64 type config from application.
	GetFloat64(path string, defaultValue ...float64) float64
	// GetBool get bool type config from application.
	GetBool(path string, defaultValue ...bool) bool
	// GetDuration get duration type config from application.
	GetDuration(path string, defaultValue ...time.Duration) time.Duration
	// GetTime get time.Time type config from application. String values are parsed with
	// time.RFC3339, "2006-01-02 15:04:05" and "2006-01-02"; the zero time is returned
	// when the value is set but can not be parsed.
	GetTime(path string, defaultValue ...time.Time) time.Time
	// GetStringSlice get []string type config from application, the stored slice is returned
	// as is, not a copy.
	GetStringSlice(path string, defaultValue ...[]string) []string
	// GetIntSlice get []int type config from application, the stored slice is returned as is,
	// not a copy.
	GetIntSlice(path string, defaultValue ...[]int) []int
	// GetStringMap get map[string]any type config from application, the stored map is returned
	// as is, not a copy.
	GetStringMap(path string, defaultValue ...map[string]any) map[string]any
	// GetStringMapString get map[string]string type config from application, the stored map is
	// returned as is, not a copy.
	GetStringMapString(path string, defaultValue ...map[string]string) map[string]string
	// Has reports whether the given path has a value in the configuration.
	Has(path string) bool
	// Forget treats the given path, and everything nested under it, as not set. Has, the
	// Get methods and All observe the path as missing until Add is called for the path or
	// one of its parents.
	Forget(path string)
	// All returns a copy of the configuration as a map[string]any. Nested maps and slices are
	// copied; other values, such as structs and pointers, are shared.
	All() map[string]any
	// UnmarshalKey unmarshal a specific key from config into a struct.
	UnmarshalKey(key string, rawVal any) error
}
