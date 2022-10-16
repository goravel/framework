package config

import (
	"os"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/testing"

	"github.com/gookit/color"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type Application struct {
	vip *viper.Viper
}

func (app *Application) Init() config.Config {
	if !file.Exist(".env") {
		color.Redln("Please create .env and initialize it first\nRun command: \ncp .env.example .env && go run . artisan key:generate")
		os.Exit(0)

		return nil
	}

	app.vip = viper.New()
	app.vip.SetConfigName(".env")
	app.vip.SetConfigType("env")
	app.vip.AddConfigPath(".")
	err := app.vip.ReadInConfig()
	if err != nil {
		if !testing.RunInTest() {
			panic(err.Error())
		}
	}
	app.vip.SetEnvPrefix("goravel")
	app.vip.AutomaticEnv()

	return app
}

//Env Get config from env.
func (app *Application) Env(envName string, defaultValue ...interface{}) interface{} {
	value := app.Get(envName, defaultValue...)
	if cast.ToString(value) == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}

		return nil
	}

	return value
}

//Add config to application.
func (app *Application) Add(name string, configuration map[string]interface{}) {
	app.vip.Set(name, configuration)
}

//Get config from application.
func (app *Application) Get(path string, defaultValue ...interface{}) interface{} {
	if !app.vip.IsSet(path) {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}

	return app.vip.Get(path)
}

//GetString Get string type config from application.
func (app *Application) GetString(path string, defaultValue ...interface{}) string {
	value := cast.ToString(app.Get(path, defaultValue...))
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0].(string)
		}

		return ""
	}

	return value
}

//GetInt Get int type config from application.
func (app *Application) GetInt(path string, defaultValue ...interface{}) int {
	value := app.Get(path, defaultValue...)
	if cast.ToString(value) == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0].(int)
		}

		return 0
	}

	return cast.ToInt(value)
}

//GetBool Get bool type config from application.
func (app *Application) GetBool(path string, defaultValue ...interface{}) bool {
	value := app.Get(path, defaultValue...)
	if cast.ToString(value) == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0].(bool)
		}

		return false
	}

	return cast.ToBool(value)
}
