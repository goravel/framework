package logger

import (
	"errors"
	"path"
	"strings"
	"time"

	rotatelogs "github.com/goravel/file-rotatelogs/v2"
	"github.com/goravel/framework/log/formatters"
	"github.com/goravel/framework/support/facades"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

type Daily struct {
}

func (daily Daily) Handle(configPath string) (logrus.Hook, error) {
	var hook logrus.Hook
	logPath := facades.Config.GetString(configPath + ".path")

	if logPath == "" {
		return hook, errors.New("error log path")
	}

	ext := path.Ext(logPath)
	logPath = strings.ReplaceAll(logPath, ext, "")

	writer, err := rotatelogs.New(
		logPath+"-%Y-%m-%d"+ext,
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
		rotatelogs.WithRotationCount(uint(facades.Config.GetInt(configPath+".days"))),
	)

	if err != nil {
		return hook, errors.New("Config local file system for logger error: " + err.Error())
	}

	return lfshook.NewHook(
		setLevel(facades.Config.GetString(configPath+".level"), writer),
		&formatters.General{},
	), nil
}
