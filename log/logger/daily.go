package logger

import (
	"errors"
	"path"
	"strings"
	"time"

	rotatelogs "github.com/goravel/file-rotatelogs/v2"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/log/formatter"
)

type Daily struct {
}

func (daily *Daily) Handle(channel string) (logrus.Hook, error) {
	var hook logrus.Hook
	logPath := facades.Config.GetString(channel + ".path")

	if logPath == "" {
		return hook, errors.New("error log path")
	}

	ext := path.Ext(logPath)
	logPath = strings.ReplaceAll(logPath, ext, "")

	writer, err := rotatelogs.New(
		logPath+"-%Y-%m-%d"+ext,
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
		rotatelogs.WithRotationCount(uint(facades.Config.GetInt(channel+".days"))),
	)

	if err != nil {
		return hook, errors.New("Config local file system for logger error: " + err.Error())
	}

	return lfshook.NewHook(
		setLevel(facades.Config.GetString(channel+".level"), writer),
		&formatter.General{},
	), nil
}
