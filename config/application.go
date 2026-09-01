package config

import (
	"os"
	"reflect"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/convert"
	"github.com/goravel/framework/support/file"
)

var _ config.Config = &Application{}

type Application struct {
	vip *viper.Viper
}

func NewApplication(envFilePath string) *Application {
	app := &Application{}
	app.vip = viper.New()
	app.vip.AutomaticEnv()

	if file.Exists(envFilePath) {
		app.vip.SetConfigType("env")
		app.vip.SetConfigFile(envFilePath)

		if err := app.vip.ReadInConfig(); err != nil {
			color.Errorln("Invalid Config error: " + err.Error())
			os.Exit(0)
		}
	}

	appKey := app.Env("APP_KEY")
	if !support.DontVerifyAppKey {
		if appKey == nil {
			color.Errorln("Please initialize APP_KEY first.")
			color.Default().Println("Create a .env file and run command: go run . artisan key:generate")
			color.Default().Println("Or set a system variable: APP_KEY={32-bit number} go run .")
			os.Exit(0)
		}

		if len(appKey.(string)) != 32 {
			color.Errorln("Invalid APP_KEY, the length must be 32, please reset it.")
			color.Warningln("Example command: \ngo run . artisan key:generate")
			os.Exit(0)
		}
	}

	return app
}

// Env Get config from env.
func (app *Application) Env(envName string, defaultValue ...any) any {
	value := app.Get(envName, defaultValue...)
	if cast.ToString(value) == "" {
		return convert.Default(defaultValue...)
	}

	return value
}

// EnvString get string value from env with optional default.
func (app *Application) EnvString(envName string, defaultValue ...string) string {
	value := app.Env(envName)
	if cast.ToString(value) == "" {
		return convert.Default(defaultValue...)
	}
	return cast.ToString(value)
}

// EnvBool get bool value from env with optional default.
func (app *Application) EnvBool(envName string, defaultValue ...bool) bool {
	value := app.Env(envName)
	// If no value and a default provided, return default
	if cast.ToString(value) == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return cast.ToBool(value)
}

// Add config to application.
func (app *Application) Add(name string, configuration any) {
	app.vip.Set(name, configuration)
}

// Get config from application.
func (app *Application) Get(path string, defaultValue ...any) any {
	if !app.vip.IsSet(path) {
		return convert.Default(defaultValue...)
	}
	return app.vip.Get(path)
}

// GetString get string type config from application.
func (app *Application) GetString(path string, defaultValue ...string) string {
	if !app.vip.IsSet(path) {
		return convert.Default(defaultValue...)
	}
	return app.vip.GetString(path)
}

// GetInt get int type config from application.
func (app *Application) GetInt(path string, defaultValue ...int) int {
	if !app.vip.IsSet(path) {
		return convert.Default(defaultValue...)
	}
	return app.vip.GetInt(path)
}

// GetBool get bool type config from application.
func (app *Application) GetBool(path string, defaultValue ...bool) bool {
	if !app.vip.IsSet(path) {
		return convert.Default(defaultValue...)
	}
	return app.vip.GetBool(path)
}

// GetDuration get time.Duration type config from application
func (app *Application) GetDuration(path string, defaultValue ...time.Duration) time.Duration {
	if !app.vip.IsSet(path) {
		return convert.Default(defaultValue...)
	}
	return app.vip.GetDuration(path)
}

// GetStringSlice get []string type config from application.
func (app *Application) GetStringSlice(path string, defaultValue ...[]string) []string {
	if !app.vip.IsSet(path) {
		return sliceOrDefault(defaultValue)
	}
	value := app.vip.GetStringSlice(path)
	if value == nil {
		return sliceOrDefault(defaultValue)
	}
	return value
}

// GetIntSlice get []int type config from application.
func (app *Application) GetIntSlice(path string, defaultValue ...[]int) []int {
	if !app.vip.IsSet(path) {
		return sliceOrDefault(defaultValue)
	}
	value := app.vip.GetIntSlice(path)
	if value == nil {
		return sliceOrDefault(defaultValue)
	}
	return value
}

// GetStringMap get map[string]any type config from application.
func (app *Application) GetStringMap(path string, defaultValue ...map[string]any) map[string]any {
	if !app.vip.IsSet(path) {
		return mapOrDefault(defaultValue)
	}
	value := app.vip.GetStringMap(path)
	if value == nil {
		return mapOrDefault(defaultValue)
	}
	return value
}

// GetStringMapString get map[string]string type config from application.
func (app *Application) GetStringMapString(path string, defaultValue ...map[string]string) map[string]string {
	if !app.vip.IsSet(path) {
		return mapOrDefault(defaultValue)
	}
	value := app.vip.GetStringMapString(path)
	if value == nil {
		return mapOrDefault(defaultValue)
	}
	return value
}

// Has reports whether the given path exists in the configuration.
func (app *Application) Has(path string) bool {
	return app.vip.IsSet(path)
}

// Forget removes the given path from the configuration.
func (app *Application) Forget(path string) {
	app.vip.Set(path, nil)
}

// All returns a copy of the entire configuration as a map[string]any.
func (app *Application) All() map[string]any {
	return app.vip.AllSettings()
}

// GetInt64 get int64 type config from application.
func (app *Application) GetInt64(path string, defaultValue ...int64) int64 {
	if !app.vip.IsSet(path) {
		return convert.Default(defaultValue...)
	}
	return app.vip.GetInt64(path)
}

// GetFloat64 get float64 type config from application.
func (app *Application) GetFloat64(path string, defaultValue ...float64) float64 {
	if !app.vip.IsSet(path) {
		return convert.Default(defaultValue...)
	}
	return app.vip.GetFloat64(path)
}

// GetTime get time.Time type config from application.
func (app *Application) GetTime(path string, defaultValue ...time.Time) time.Time {
	if !app.vip.IsSet(path) {
		return convert.Default(defaultValue...)
	}
	value := app.vip.Get(path)
	if t, ok := value.(time.Time); ok {
		return t
	}
	if str, ok := value.(string); ok {
		if layouts := []string{time.RFC3339, time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02"}; true {
			for _, layout := range layouts {
				if t, err := time.Parse(layout, str); err == nil {
					return t
				}
			}
		}
	}
	return convert.Default(defaultValue...)
}

func sliceOrDefault[T any](values []T) T {
	for _, v := range values {
		if !isZeroValue(v) {
			return v
		}
	}
	var zero T
	return zero
}

func mapOrDefault[T any](values []T) T {
	for _, v := range values {
		if !isZeroValue(v) {
			return v
		}
	}
	var zero T
	return zero
}

func isZeroValue(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array, reflect.Chan, reflect.Pointer, reflect.Interface:
		return rv.IsNil()
	}
	return false
}

// UnmarshalKey unmarshal a specific key from config into a struct.
func (app *Application) UnmarshalKey(key string, rawVal any) error {
	return app.vip.UnmarshalKey(key, rawVal, func(c *mapstructure.DecoderConfig) {
		c.TagName = "json"
	})
}
