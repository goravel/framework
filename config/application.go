package config

import (
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"log"
)

type Application struct {
	vip *viper.Viper
}

func (app *Application) Init() *Application {
	app.vip = viper.New()
	app.vip.SetConfigName(".env")
	app.vip.SetConfigType("env")
	app.vip.AddConfigPath(".")
	err := app.vip.ReadInConfig()
	if err != nil {
		log.Fatalln("Please init .env file.", err.Error())
		return app
	}
	app.vip.SetEnvPrefix("goravel")
	app.vip.AutomaticEnv()

	return app
}

//Env Get config from env.
func (app *Application) Env(envName string, defaultValue ...interface{}) interface{} {
	if len(defaultValue) > 0 {
		return app.Get(envName, defaultValue[0])
	}
	return app.Get(envName)
}

//Add Add config to application.
func (app *Application) Add(name string, configuration map[string]interface{}) {
	app.vip.Set(name, configuration)
}

//Get Get config from application.
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
	return cast.ToString(app.Get(path, defaultValue...))
}

//GetInt Get int type config from application.
func (app *Application) GetInt(path string, defaultValue ...interface{}) int {
	return cast.ToInt(app.Get(path, defaultValue...))
}

//GetBool Get bool type config from application.
func (app *Application) GetBool(path string, defaultValue ...interface{}) bool {
	return cast.ToBool(app.Get(path, defaultValue...))
}
