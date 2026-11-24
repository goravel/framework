package config

import (
	"strings"
	"time"

	"github.com/goravel/framework/contracts/config"
)

var _ config.Repository = &Module{}

// Module is the config of one module: every path is read under "modules.<name>".
type Module struct {
	config config.Repository
	name   string
}

func NewModule(config config.Repository, name string) *Module {
	return &Module{config: config, name: name}
}

func (r *Module) Add(name string, configuration any) {
	r.config.Add(r.path(name), configuration)
}

func (r *Module) Get(path string, defaultValue ...any) any {
	return r.config.Get(r.path(path), defaultValue...)
}

func (r *Module) GetString(path string, defaultValue ...string) string {
	return r.config.GetString(r.path(path), defaultValue...)
}

func (r *Module) GetInt(path string, defaultValue ...int) int {
	return r.config.GetInt(r.path(path), defaultValue...)
}

func (r *Module) GetInt64(path string, defaultValue ...int64) int64 {
	return r.config.GetInt64(r.path(path), defaultValue...)
}

func (r *Module) GetFloat64(path string, defaultValue ...float64) float64 {
	return r.config.GetFloat64(r.path(path), defaultValue...)
}

func (r *Module) GetBool(path string, defaultValue ...bool) bool {
	return r.config.GetBool(r.path(path), defaultValue...)
}

func (r *Module) GetDuration(path string, defaultValue ...time.Duration) time.Duration {
	return r.config.GetDuration(r.path(path), defaultValue...)
}

func (r *Module) GetTime(path string, defaultValue ...time.Time) time.Time {
	return r.config.GetTime(r.path(path), defaultValue...)
}

func (r *Module) GetStringSlice(path string, defaultValue ...[]string) []string {
	return r.config.GetStringSlice(r.path(path), defaultValue...)
}

func (r *Module) GetIntSlice(path string, defaultValue ...[]int) []int {
	return r.config.GetIntSlice(r.path(path), defaultValue...)
}

func (r *Module) GetStringMap(path string, defaultValue ...map[string]any) map[string]any {
	return r.config.GetStringMap(r.path(path), defaultValue...)
}

func (r *Module) GetStringMapString(path string, defaultValue ...map[string]string) map[string]string {
	return r.config.GetStringMapString(r.path(path), defaultValue...)
}

func (r *Module) Has(path string) bool {
	return r.config.Has(r.path(path))
}

func (r *Module) Forget(path string) {
	r.config.Forget(r.path(path))
}

// All returns a copy of the module's configuration, empty when the module has none.
func (r *Module) All() map[string]any {
	modules, _ := r.config.All()["modules"].(map[string]any)
	// viper lowercases the keys it stores.
	all, _ := modules[strings.ToLower(r.name)].(map[string]any)
	if all == nil {
		return map[string]any{}
	}

	return all
}

func (r *Module) UnmarshalKey(key string, rawVal any) error {
	return r.config.UnmarshalKey(r.path(key), rawVal)
}

// path scopes path to the module, an empty path is the module's whole config.
func (r *Module) path(path string) string {
	if path == "" {
		return "modules." + r.name
	}

	return "modules." + r.name + "." + path
}
