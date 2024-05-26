package log

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/support/color"
)

type Application struct {
	log.Writer
	instance *logrus.Logger
	config   config.Config
}

func NewApplication(config config.Config) *Application {
	instance := logrus.New()
	instance.SetLevel(logrus.DebugLevel)

	if config != nil {
		if channel := config.GetString("logging.default"); channel != "" {
			if err := registerHook(config, instance, channel); err != nil {
				color.Red().Println("Init facades.Log error: " + err.Error())
				return nil
			}
		}
	}

	return &Application{
		instance: instance,
		Writer:   NewWriter(instance.WithContext(context.Background())),
		config:   config,
	}
}

func (r *Application) WithContext(ctx context.Context) log.Writer {
	return NewWriter(r.instance.WithContext(ctx))
}

func (r *Application) Channel(channel string) log.Writer {
	if channel == "" || r.config == nil {
		return r.Writer
	}

	instance := logrus.New()
	instance.SetLevel(logrus.DebugLevel)

	if err := registerHook(r.config, instance, channel); err != nil {
		color.Red().Println("Init facades.Log error: " + err.Error())
		return nil
	}

	return NewWriter(instance.WithContext(context.Background()))
}

func (r *Application) Stack(channels []string) log.Writer {
	if r.config == nil || len(channels) == 0 {
		return r.Writer
	}

	instance := logrus.New()
	instance.SetLevel(logrus.DebugLevel)

	for _, channel := range channels {
		if channel == "" {
			continue
		}

		if err := registerHook(r.config, instance, channel); err != nil {
			color.Red().Println("Init facades.Log error: " + err.Error())
			return nil
		}
	}

	return NewWriter(instance.WithContext(context.Background()))
}
