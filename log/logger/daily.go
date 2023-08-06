package logger

import (
	"errors"
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/goravel/file-rotatelogs/v2"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/log/formatter"
	"github.com/goravel/framework/support"
)

type Daily struct {
	config config.Config
}

func NewDaily(config config.Config) *Daily {
	return &Daily{
		config: config,
	}
}

func (daily *Daily) Handle(channel string) (logrus.Hook, error) {
	var hook logrus.Hook
	logPath := daily.config.GetString(channel + ".path")
	if logPath == "" {
		return hook, errors.New("error log path")
	}

	ext := filepath.Ext(logPath)
	logPath = strings.ReplaceAll(logPath, ext, "")
	logPath = filepath.Join(support.RelativePath, logPath)

	writer, err := rotatelogs.New(
		logPath+"-%Y-%m-%d"+ext,
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
		rotatelogs.WithRotationCount(uint(daily.config.GetInt(channel+".days"))),
	)
	if err != nil {
		return hook, errors.New("Config local file system for logger error: " + err.Error())
	}

	levels := getLevels(daily.config.GetString(channel + ".level"))
	writerMap := lfshook.WriterMap{}
	for _, level := range levels {
		writerMap[level] = writer
	}

	return lfshook.NewHook(
		writerMap,
		formatter.NewGeneral(daily.config),
	), nil
}
