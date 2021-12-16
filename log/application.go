package log

import (
	"github.com/goravel/framework/log/logger"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/facades"
	"github.com/sirupsen/logrus"
	"log"
)

type Application struct {
	log *logrus.Logger
}

func (app *Application) Init() *logrus.Logger {
	app.log = logrus.New()
	app.log.SetLevel(logrus.TraceLevel)
	app.registerHook(facades.Config.GetString("logging.default"))

	return app.log
}

//registerHook Register hook
func (app *Application) registerHook(channel string) {
	driver := facades.Config.GetString("logging.channels." + channel + ".driver")
	configPath := "logging.channels." + channel

	switch driver {
	case "stack":
		for _, stackChannel := range facades.Config.Get("logging.channels." + channel + ".channels").([]string) {
			if stackChannel == channel {
				log.Fatalln("Stack drive can't include self channel.")
			}

			app.registerHook(stackChannel)
		}
	case "single":
		app.log.AddHook(logger.Single{}.Handle(configPath))
	case "daily":
		app.log.AddHook(logger.Daily{}.Handle(configPath))
	case "custom":
		customLogger := facades.Config.Get("logging.channels." + channel + ".via").(support.Logger)
		app.log.AddHook(customLogger.Handle(configPath))
	default:
		log.Fatalln("Error logging channel: " + channel)
	}
}
