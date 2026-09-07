package config

import (
	"time"
)

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
	// GetDuration get duration type config from application
	GetDuration(path string, defaultValue ...time.Duration) time.Duration
	// GetTime get time.Time type config from application.
	GetTime(path string, defaultValue ...time.Time) time.Time
	// GetStringSlice get []string type config from application.
	GetStringSlice(path string, defaultValue ...[]string) []string
	// GetIntSlice get []int type config from application.
	GetIntSlice(path string, defaultValue ...[]int) []int
	// GetStringMap get map[string]any type config from application.
	GetStringMap(path string, defaultValue ...map[string]any) map[string]any
	// GetStringMapString get map[string]string type config from application.
	GetStringMapString(path string, defaultValue ...map[string]string) map[string]string
	// Has reports whether the given path exists in the configuration.
	Has(path string) bool
	// Forget removes the given path from the configuration.
	Forget(path string)
	// All returns a copy of the entire configuration as a map.
	All() map[string]any
	// UnmarshalKey unmarshal a specific key from config into a struct.
	UnmarshalKey(key string, rawVal any) error
}
