package config

import (
	"os"

	"github.com/gookit/color"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/file"
)

var _ config.Config = &Application{}

type Application struct {
	vip *viper.Viper
}

func NewApplication(envPath string) *Application {
	if !file.Exists(envPath) {
		color.Redln("Please create " + envPath + " and initialize it first.")
		color.Warnln("Example command: \ncp .env.example .env && go run . artisan key:generate")
		os.Exit(0)
	}

	app := &Application{}
	app.vip = viper.New()
	app.vip.SetConfigType("env")
	app.vip.SetConfigFile(envPath)

	if err := app.vip.ReadInConfig(); err != nil {
		color.Redln("Invalid Config error: " + err.Error())
		os.Exit(0)
	}

	app.vip.SetEnvPrefix("goravel")
	app.vip.AutomaticEnv()

	appKey := app.Env("APP_KEY")
	if support.Env != support.EnvArtisan {
		if appKey == nil {
			color.Redln("Please initialize APP_KEY first.")
			color.Warnln("Example command: \ngo run . artisan key:generate")
			os.Exit(0)
		}

		if len(appKey.(string)) != 32 {
			color.Redln("Invalid APP_KEY, please reset it.")
			color.Warnln("Example command: \ngo run . artisan key:generate")
			os.Exit(0)
		}
	}

	return app
}

// Env Get config from env.
func (app *Application) Env(envName string, defaultValue ...any) any {
	value := app.Get(envName, defaultValue...)
	if cast.ToString(value) == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}

		return nil
	}

	return value
}

// Add config to application.
func (app *Application) Add(name string, configuration any) {
	app.vip.Set(name, configuration)
}

// Get config from application.
func (app *Application) Get(path string, defaultValue ...any) any {
	if !app.vip.IsSet(path) {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}

	return app.vip.Get(path)
}

// GetString Get string type config from application.
func (app *Application) GetString(path string, defaultValue ...any) string {
	value := cast.ToString(app.Get(path, defaultValue...))
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0].(string)
		}

		return ""
	}

	return value
}

// GetInt Get int type config from application.
func (app *Application) GetInt(path string, defaultValue ...any) int {
	value := app.Get(path, defaultValue...)
	if cast.ToString(value) == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0].(int)
		}

		return 0
	}

	return cast.ToInt(value)
}

// GetBool Get bool type config from application.
func (app *Application) GetBool(path string, defaultValue ...any) bool {
	value := app.Get(path, defaultValue...)
	if cast.ToString(value) == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0].(bool)
		}

		return false
	}

	return cast.ToBool(value)
}
