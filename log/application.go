package log

import (
	"errors"

	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/log/logger"
	"github.com/goravel/framework/support/facades"
	"github.com/sirupsen/logrus"
)

type Application struct {
	log *logrus.Logger
}

func (app *Application) Init() *logrus.Logger {
	app.log = logrus.New()
	app.log.SetLevel(logrus.TraceLevel)
	if err := app.registerHook(facades.Config.GetString("logging.default")); err != nil {
		panic("Log Init error: " + err.Error())
	}

	return app.log
}

//registerHook Register hook
func (app *Application) registerHook(channel string) error {
	var hook log.Logger
	driver := facades.Config.GetString("logging.channels." + channel + ".driver")
	configPath := "logging.channels." + channel

	switch driver {
	case "stack":
		for _, stackChannel := range facades.Config.Get("logging.channels." + channel + ".channels").([]string) {
			if stackChannel == channel {
				return errors.New("stack drive can't include self channel")
			}

			if err := app.registerHook(stackChannel); err != nil {
				return err
			}
		}

		return nil
	case "single":
		hook = logger.Single{}
	case "daily":
		hook = logger.Daily{}
	case "custom":
		hook = facades.Config.Get("logging.channels." + channel + ".via").(log.Logger)
	default:
		return errors.New("Error logging channel: " + channel)
	}

	logHook, err := hook.Handle(configPath)
	if err != nil {
		return err
	}

	app.log.AddHook(logHook)

	return nil
}
