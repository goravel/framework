package log

import (
	"context"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/log"
)

type Application struct {
	log.Writer
	instance *logrus.Logger
}

func NewApplication(config config.Config) *Application {
	instance := logrus.New()
	instance.SetLevel(logrus.DebugLevel)

	if config != nil {
		if logging := config.GetString("logging.default"); logging != "" {
			if err := registerHook(config, instance, logging); err != nil {
				color.Redln("Init facades.Log error: " + err.Error())

				return nil
			}
		}
	}

	return &Application{
		instance: instance,
		Writer:   NewWriter(instance.WithContext(context.Background())),
	}
}

func (r *Application) WithContext(ctx context.Context) log.Writer {
	switch r.Writer.(type) {
	case *Writer:
		return NewWriter(r.instance.WithContext(ctx))
	default:
		return r.Writer
	}
}
