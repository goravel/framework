package config

import "time"

// Config is the application's configuration: a Repository over the whole config,
// plus the env helpers and per-module views.
type Config interface {
	Repository

	// Env get config from env.
	Env(envName string, defaultValue ...any) any
	// EnvString get string value from env with optional default.
	EnvString(envName string, defaultValue ...string) string
	// EnvBool get bool value from env with optional default.
	EnvBool(envName string, defaultValue ...bool) bool
	// Module returns the config of the given module: paths are read under "modules.<name>",
	// so Module("blog").GetInt("per_page") is GetInt("modules.blog.per_page"). An empty
	// path is the module's whole config.
	Module(name string) Repository
}

// Repository holds the config operations shared by the application's configuration
// and a module's view of it.
type Repository interface {
	// Add config.
	Add(name string, configuration any)
	// Get config.
	Get(path string, defaultValue ...any) any
	// GetString get string type config.
	GetString(path string, defaultValue ...string) string
	// GetInt get int type config.
	GetInt(path string, defaultValue ...int) int
	// GetInt64 get int64 type config.
	GetInt64(path string, defaultValue ...int64) int64
	// GetFloat64 get float64 type config.
	GetFloat64(path string, defaultValue ...float64) float64
	// GetBool get bool type config.
	GetBool(path string, defaultValue ...bool) bool
	// GetDuration get duration type config.
	GetDuration(path string, defaultValue ...time.Duration) time.Duration
	// GetTime get time.Time type config. String values are parsed with
	// time.RFC3339, "2006-01-02 15:04:05" and "2006-01-02"; the zero time is returned
	// when the value is set but can not be parsed.
	GetTime(path string, defaultValue ...time.Time) time.Time
	// GetStringSlice get []string type config, the stored slice is returned
	// as is, not a copy.
	GetStringSlice(path string, defaultValue ...[]string) []string
	// GetIntSlice get []int type config, the stored slice is returned as is,
	// not a copy.
	GetIntSlice(path string, defaultValue ...[]int) []int
	// GetStringMap get map[string]any type config, the stored map is returned
	// as is, not a copy.
	GetStringMap(path string, defaultValue ...map[string]any) map[string]any
	// GetStringMapString get map[string]string type config, the stored map is
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
