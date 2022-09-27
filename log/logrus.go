package log

import (
	"errors"
	"fmt"
	"github.com/gookit/color"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/log/logger"

	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/contracts/log"
)

type Logrus struct {
	Instance *logrus.Logger
	Test     bool
}

func NewLogrus() log.Log {
	instance := logrus.New()
	instance.SetLevel(logrus.DebugLevel)
	if err := registerHook(instance, facades.Config.GetString("logging.default")); err != nil {
		color.Redln("Init facades.Log error: " + err.Error())

		return nil
	}

	return &Logrus{instance, false}
}

func (r *Logrus) Testing() log.Log {
	r.Test = true

	return r
}

func (r *Logrus) Debug(args ...interface{}) {
	if r.Test {
		fmt.Print("Debug: ")
		fmt.Println(args...)
		return
	}

	r.Instance.Debug(args...)
}

func (r *Logrus) Debugf(format string, args ...interface{}) {
	if r.Test {
		fmt.Print("Debugf: ")
		fmt.Printf(format+"\n", args...)
		return
	}

	r.Instance.Debugf(format, args...)
}

func (r *Logrus) Info(args ...interface{}) {
	if r.Test {
		fmt.Print("Info: ")
		fmt.Println(args...)
		return
	}

	r.Instance.Info(args...)
}

func (r *Logrus) Infof(format string, args ...interface{}) {
	if r.Test {
		fmt.Print("Infof: ")
		fmt.Printf(format+"\n", args...)
		return
	}

	r.Instance.Infof(format, args...)
}

func (r *Logrus) Warning(args ...interface{}) {
	if r.Test {
		fmt.Print("Warningf: ")
		fmt.Println(args...)
		return
	}

	r.Instance.Warning(args...)
}

func (r *Logrus) Warningf(format string, args ...interface{}) {
	if r.Test {
		fmt.Print("Warningf: ")
		fmt.Printf(format+"\n", args...)
		return
	}

	r.Instance.Warningf(format, args...)
}

func (r *Logrus) Error(args ...interface{}) {
	if r.Test {
		fmt.Print("Error: ")
		fmt.Println(args...)
		return
	}

	r.Instance.Error(args...)
}

func (r *Logrus) Errorf(format string, args ...interface{}) {
	if r.Test {
		fmt.Print("Errorf: ")
		fmt.Printf(format+"\n", args...)
		return
	}

	r.Instance.Errorf(format, args...)
}

func (r *Logrus) Fatal(args ...interface{}) {
	if r.Test {
		fmt.Print("Error: ")
		fmt.Println(args...)
		return
	}

	r.Instance.Fatal(args...)
}

func (r *Logrus) Fatalf(format string, args ...interface{}) {
	if r.Test {
		fmt.Print("Errorf: ")
		fmt.Printf(format+"\n", args...)
		return
	}

	r.Instance.Fatalf(format, args...)
}

func (r *Logrus) Panic(args ...interface{}) {
	if r.Test {
		fmt.Print("Panic: ")
		fmt.Println(args...)
		return
	}

	r.Instance.Panic(args...)
}

func (r *Logrus) Panicf(format string, args ...interface{}) {
	if r.Test {
		fmt.Print("Panicf: ")
		fmt.Printf(format+"\n", args...)
		return
	}

	r.Instance.Panicf(format, args...)
}

func registerHook(instance *logrus.Logger, channel string) error {
	var hook log.Hook
	driver := facades.Config.GetString("logging.channels." + channel + ".driver")
	configPath := "logging.channels." + channel

	switch driver {
	case "stack":
		for _, stackChannel := range facades.Config.Get("logging.channels." + channel + ".channels").([]string) {
			if stackChannel == channel {
				return errors.New("stack drive can't include self channel")
			}

			if err := registerHook(instance, stackChannel); err != nil {
				return err
			}
		}

		return nil
	case "single":
		hook = logger.Single{}
	case "daily":
		hook = logger.Daily{}
	case "custom":
		hook = facades.Config.Get("logging.channels." + channel + ".via").(log.Hook)
	default:
		return errors.New("Error logging channel: " + channel)
	}

	logHook, err := hook.Handle(configPath)
	if err != nil {
		return err
	}

	instance.AddHook(logHook)

	return nil
}
