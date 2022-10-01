package logger

import (
	"errors"
	"io"
	"os"
	"path"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/log/formatter"
)

type Single struct {
}

func (single *Single) Handle(channel string) (logrus.Hook, error) {
	logPath := facades.Config.GetString(channel + ".path")
	err := os.MkdirAll(path.Dir(logPath), os.ModePerm)
	if err != nil {
		return nil, errors.New("Create dir fail:" + err.Error())
	}

	writer, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		return nil, errors.New("Failed to log to file:" + err.Error())
	}

	return lfshook.NewHook(
		setLevel(facades.Config.GetString(channel+".level"), writer),
		&formatter.General{},
	), nil
}

func setLevel(level string, writer io.Writer) lfshook.WriterMap {
	if level == "panic" {
		return lfshook.WriterMap{
			logrus.PanicLevel: writer,
		}
	}

	if level == "fatal" {
		return lfshook.WriterMap{
			logrus.FatalLevel: writer,
			logrus.PanicLevel: writer,
		}
	}

	if level == "error" {
		return lfshook.WriterMap{
			logrus.ErrorLevel: writer,
			logrus.FatalLevel: writer,
			logrus.PanicLevel: writer,
		}
	}

	if level == "warning" {
		return lfshook.WriterMap{
			logrus.WarnLevel:  writer,
			logrus.ErrorLevel: writer,
			logrus.FatalLevel: writer,
			logrus.PanicLevel: writer,
		}
	}

	if level == "info" {
		return lfshook.WriterMap{
			logrus.InfoLevel:  writer,
			logrus.WarnLevel:  writer,
			logrus.ErrorLevel: writer,
			logrus.FatalLevel: writer,
			logrus.PanicLevel: writer,
		}
	}

	return lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}
}
