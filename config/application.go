package config

import (
	"os"
	"reflect"
	"strings"
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

// timeLayouts are the layouts GetTime tries, in order, for values stored as strings.
// time.RFC3339 already accepts fractional seconds, so time.RFC3339Nano is not needed.
var timeLayouts = []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"}

type Application struct {
	vip       *viper.Viper
	forgotten map[string]struct{}
}

func NewApplication(envFilePath string) *Application {
	app := &Application{
		forgotten: make(map[string]struct{}),
	}
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
	key := strings.ToLower(name)
	for forgotten := range app.forgotten {
		if forgotten == key || strings.HasPrefix(forgotten, key+".") {
			delete(app.forgotten, forgotten)
		}
	}

	app.vip.Set(name, configuration)
}

// Get config from application.
func (app *Application) Get(path string, defaultValue ...any) any {
	if !app.has(path) {
		return convert.Default(defaultValue...)
	}
	return app.vip.Get(path)
}

// GetString get string type config from application.
func (app *Application) GetString(path string, defaultValue ...string) string {
	if !app.has(path) {
		return convert.Default(defaultValue...)
	}
	return app.vip.GetString(path)
}

// GetInt get int type config from application.
func (app *Application) GetInt(path string, defaultValue ...int) int {
	if !app.has(path) {
		return convert.Default(defaultValue...)
	}
	return app.vip.GetInt(path)
}

// GetInt64 get int64 type config from application.
func (app *Application) GetInt64(path string, defaultValue ...int64) int64 {
	if !app.has(path) {
		return convert.Default(defaultValue...)
	}
	return app.vip.GetInt64(path)
}

// GetFloat64 get float64 type config from application.
func (app *Application) GetFloat64(path string, defaultValue ...float64) float64 {
	if !app.has(path) {
		return convert.Default(defaultValue...)
	}
	return app.vip.GetFloat64(path)
}

// GetBool get bool type config from application.
func (app *Application) GetBool(path string, defaultValue ...bool) bool {
	if !app.has(path) {
		return convert.Default(defaultValue...)
	}
	return app.vip.GetBool(path)
}

// GetDuration get time.Duration type config from application.
func (app *Application) GetDuration(path string, defaultValue ...time.Duration) time.Duration {
	if !app.has(path) {
		return convert.Default(defaultValue...)
	}
	return app.vip.GetDuration(path)
}

// GetTime get time.Time type config from application.
func (app *Application) GetTime(path string, defaultValue ...time.Time) time.Time {
	if !app.has(path) {
		return convert.Default(defaultValue...)
	}

	value := app.vip.Get(path)
	if parsed, ok := value.(time.Time); ok {
		return parsed
	}

	str, ok := value.(string)
	if !ok {
		return time.Time{}
	}

	for _, layout := range timeLayouts {
		if parsed, err := time.Parse(layout, str); err == nil {
			return parsed
		}
	}

	return time.Time{}
}

// GetStringSlice get []string type config from application.
func (app *Application) GetStringSlice(path string, defaultValue ...[]string) []string {
	if !app.has(path) {
		return firstNonNil(defaultValue)
	}

	value := app.vip.GetStringSlice(path)
	if value == nil {
		return firstNonNil(defaultValue)
	}

	return value
}

// GetIntSlice get []int type config from application.
func (app *Application) GetIntSlice(path string, defaultValue ...[]int) []int {
	if !app.has(path) {
		return firstNonNil(defaultValue)
	}

	value := app.vip.GetIntSlice(path)
	if value == nil {
		return firstNonNil(defaultValue)
	}

	return value
}

// GetStringMap get map[string]any type config from application.
func (app *Application) GetStringMap(path string, defaultValue ...map[string]any) map[string]any {
	if !app.has(path) {
		return firstNonNil(defaultValue)
	}

	value := app.vip.GetStringMap(path)
	if value == nil {
		return firstNonNil(defaultValue)
	}

	return value
}

// GetStringMapString get map[string]string type config from application.
func (app *Application) GetStringMapString(path string, defaultValue ...map[string]string) map[string]string {
	if !app.has(path) {
		return firstNonNil(defaultValue)
	}

	value := app.vip.GetStringMapString(path)
	if value == nil {
		return firstNonNil(defaultValue)
	}

	return value
}

// Has reports whether the given path has a value in the configuration.
func (app *Application) Has(path string) bool {
	return app.has(path)
}

// Forget treats the given path, and everything nested under it, as not set.
func (app *Application) Forget(path string) {
	if app.forgotten == nil {
		app.forgotten = make(map[string]struct{})
	}

	app.forgotten[strings.ToLower(path)] = struct{}{}
}

// All returns a copy of the entire configuration. Nested maps and slices are copied, other
// values such as structs and pointers are shared.
func (app *Application) All() map[string]any {
	all := deepCopyMap(app.vip.AllSettings())
	for path := range app.forgotten {
		removePath(all, strings.Split(path, "."))
	}

	return all
}

// UnmarshalKey unmarshal a specific key from config into a struct.
func (app *Application) UnmarshalKey(key string, rawVal any) error {
	if !app.has(key) {
		return nil
	}

	return app.vip.UnmarshalKey(key, rawVal, func(c *mapstructure.DecoderConfig) {
		c.TagName = "json"
	})
}

// has reports whether path holds a value and has not been forgotten.
func (app *Application) has(path string) bool {
	return !app.isForgotten(path) && app.vip.IsSet(path)
}

// isForgotten reports whether path, or any of its parents, has been forgotten.
func (app *Application) isForgotten(path string) bool {
	if len(app.forgotten) == 0 {
		return false
	}

	key := strings.ToLower(path)
	for {
		if _, ok := app.forgotten[key]; ok {
			return true
		}

		index := strings.LastIndex(key, ".")
		if index < 0 {
			return false
		}
		key = key[:index]
	}
}

// removePath deletes the value stored at path from settings, walking maps of any type so a
// path inside a map[string]string is removed as well. A parent left empty is kept, Has still
// reports it.
func removePath(settings map[string]any, path []string) {
	removeMapPath(reflect.ValueOf(settings), path)
}

// removeMapPath deletes path from a map held in value. Keys are compared without case because
// viper lowercases every lookup path.
func removeMapPath(value reflect.Value, path []string) {
	// MapRange yields the interface type for map[string]any elements, so unwrap before checking.
	for value.Kind() == reflect.Interface && !value.IsNil() {
		value = value.Elem()
	}

	if len(path) == 0 || value.Kind() != reflect.Map || value.IsNil() || value.Type().Key().Kind() != reflect.String {
		return
	}

	key := strings.ToLower(path[0])
	iterator := value.MapRange()
	for iterator.Next() {
		if strings.ToLower(iterator.Key().String()) != key {
			continue
		}

		if len(path) == 1 {
			value.SetMapIndex(iterator.Key(), reflect.Value{})
			return
		}

		removeMapPath(iterator.Value(), path[1:])
		return
	}
}

// firstNonNil returns the first non-nil value, or the zero value when every value is nil.
func firstNonNil[T any](values []T) T {
	for _, value := range values {
		if !isNil(value) {
			return value
		}
	}

	var zero T
	return zero
}

// isNil reports whether value is nil, including typed nil slices, maps and pointers.
// reflect.Array is intentionally not handled: IsNil panics on arrays.
func isNil(value any) bool {
	if value == nil {
		return true
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	}

	return false
}
