package log

import (
	"errors"
	"io/ioutil"

	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/log/logger"
)

type Writer struct {
	instance *logrus.Entry
}

func NewWriter(instance *logrus.Entry) log.Writer {
	return &Writer{instance: instance}
}

func (r *Writer) Debug(args ...any) {
	r.instance.Debug(args...)
}

func (r *Writer) Debugf(format string, args ...any) {
	r.instance.Debugf(format, args...)
}

func (r *Writer) Info(args ...any) {
	r.instance.Info(args...)
}

func (r *Writer) Infof(format string, args ...any) {
	r.instance.Infof(format, args...)
}

func (r *Writer) Warning(args ...any) {
	r.instance.Warning(args...)
}

func (r *Writer) Warningf(format string, args ...any) {
	r.instance.Warningf(format, args...)
}

func (r *Writer) Error(args ...any) {
	r.instance.Error(args...)
}

func (r *Writer) Errorf(format string, args ...any) {
	r.instance.Errorf(format, args...)
}

func (r *Writer) Fatal(args ...any) {
	r.instance.Fatal(args...)
}

func (r *Writer) Fatalf(format string, args ...any) {
	r.instance.Fatalf(format, args...)
}

func (r *Writer) Panic(args ...any) {
	r.instance.Panic(args...)
}

func (r *Writer) Panicf(format string, args ...any) {
	r.instance.Panicf(format, args...)
}

func registerHook(instance *logrus.Logger, channel string) error {
	channelPath := "logging.channels." + channel
	driver := facades.Config.GetString(channelPath + ".driver")

	var hook logrus.Hook
	var err error
	switch driver {
	case log.StackDriver:
		for _, stackChannel := range facades.Config.Get(channelPath + ".channels").([]string) {
			if stackChannel == channel {
				return errors.New("stack drive can't include self channel")
			}

			if err := registerHook(instance, stackChannel); err != nil {
				return err
			}
		}

		return nil
	case log.SingleDriver:
		if !facades.Config.GetBool(channelPath + ".print") {
			instance.SetOutput(ioutil.Discard)
		}

		logLogger := &logger.Single{}
		hook, err = logLogger.Handle(channelPath)
		if err != nil {
			return err
		}
	case log.DailyDriver:
		if !facades.Config.GetBool(channelPath + ".print") {
			instance.SetOutput(ioutil.Discard)
		}

		logLogger := &logger.Daily{}
		hook, err = logLogger.Handle(channelPath)
		if err != nil {
			return err
		}
	case log.CustomDriver:
		logLogger := facades.Config.Get(channelPath + ".via").(log.Logger)
		logHook, err := logLogger.Handle(channelPath)
		if err != nil {
			return err
		}

		hook = &Hook{logHook}
	default:
		return errors.New("Error logging channel: " + channel)
	}

	instance.AddHook(hook)

	return nil
}

type Hook struct {
	instance log.Hook
}

func (h *Hook) Levels() []logrus.Level {
	levels := h.instance.Levels()
	var logrusLevels []logrus.Level
	for _, item := range levels {
		logrusLevels = append(logrusLevels, logrus.Level(item))
	}

	return logrusLevels
}

func (h *Hook) Fire(entry *logrus.Entry) error {
	return h.instance.Fire(&Entry{
		ctx:     entry.Context,
		level:   log.Level(entry.Level),
		time:    entry.Time,
		message: entry.Message,
	})
}
