package logger

import (
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/goravel/file-rotatelogs/v2"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/log/formatter"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/carbon"
)

type Daily struct {
	config config.Config
	json   foundation.Json
}

func NewDaily(config config.Config, json foundation.Json) *Daily {
	return &Daily{
		config: config,
		json:   json,
	}
}

func (daily *Daily) Handle(channel string) (logrus.Hook, error) {
	var hook logrus.Hook
	logPath := daily.config.GetString(channel + ".path")
	if logPath == "" {
		return hook, errors.LogEmptyLogFilePath
	}

	ext := filepath.Ext(logPath)
	logPath = strings.ReplaceAll(logPath, ext, "")
	logPath = filepath.Join(support.RelativePath, logPath)

	writer, err := rotatelogs.New(
		logPath+"-%Y-%m-%d"+ext,
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
		rotatelogs.WithRotationCount(uint(daily.config.GetInt(channel+".days"))),
		// When using carbon.SetTestNow(), carbon.Now().StdTime() should always be used to get the current time.
		// Hence, WithLocation cannot be used here.
		rotatelogs.WithClock(NewRotatelogsClock()),
	)
	if err != nil {
		return hook, err
	}

	levels := getLevels(daily.config.GetString(channel + ".level"))
	writerMap := lfshook.WriterMap{}
	for _, level := range levels {
		writerMap[level] = writer
	}

	return lfshook.NewHook(
		writerMap,
		formatter.NewGeneral(daily.config, daily.json),
	), nil
}

type rotatelogsClock struct{}

func (clock *rotatelogsClock) Now() time.Time {
	return carbon.Now().StdTime()
}

func NewRotatelogsClock() rotatelogs.Clock {
	return &rotatelogsClock{}
}
