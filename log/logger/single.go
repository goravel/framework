package logger

import (
	"github.com/goravel/framework/log/formatters"
	"github.com/goravel/framework/support/facades"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"os"
	"path"
)

type Single struct {
}

func (single Single) Handle(configPath string) logrus.Hook {
	logPath := facades.Config.GetString(configPath + ".path")

	err := os.MkdirAll(path.Dir(logPath), os.ModePerm)

	if err != nil {
		log.Fatalln("Create dir fail:", err.Error())
	}

	writer, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		log.Fatalln("Failed to log to file:", err.Error())
	}

	return lfshook.NewHook(
		setLevel(facades.Config.GetString(configPath+".level"), writer),
		&formatters.General{},
	)
}

func setLevel(level string, writer io.Writer) lfshook.WriterMap {
	if level == "error" {
		return lfshook.WriterMap{
			logrus.ErrorLevel: writer,
			logrus.FatalLevel: writer,
			logrus.PanicLevel: writer,
		}
	}

	if level == "warn" {
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
		logrus.TraceLevel: writer,
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}
}
