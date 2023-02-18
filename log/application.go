package log

import (
	"context"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/facades"
)

type Logrus struct {
	instance *logrus.Logger
	log.Writer
}

func NewApplication(writer log.Writer) *Logrus {
	return &Logrus{
		Writer: writer,
	}
}

func NewLogrusApplication() *Logrus {
	instance := newLogrus()

	return &Logrus{
		instance: instance,
		Writer:   NewWriter(instance.WithContext(context.Background())),
	}
}

func (r *Logrus) WithContext(ctx context.Context) log.Writer {
	switch r.Writer.(type) {
	case *Writer:
		return NewWriter(r.instance.WithContext(ctx))
	default:
		return r.Writer
	}
}

func newLogrus() *logrus.Logger {
	instance := logrus.New()
	instance.SetLevel(logrus.DebugLevel)

	if facades.Config != nil {
		if logging := facades.Config.GetString("logging.default"); logging != "" {
			if err := registerHook(instance, logging); err != nil {
				color.Redln("Init facades.Log error: " + err.Error())

				return nil
			}
		}
	}

	return instance
}
