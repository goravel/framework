package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/goravel/file-rotatelogs/v2"

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

func (daily *Daily) Handle(channel string) (slog.Handler, error) {
	logPath := daily.config.GetString(channel + ".path")
	if logPath == "" {
		return nil, errors.LogEmptyLogFilePath
	}

	ext := filepath.Ext(logPath)
	logPath = strings.ReplaceAll(logPath, ext, "")
	logPath = filepath.Join(support.RelativePath, logPath)

	// Ensure parent directory exists
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	writer, err := rotatelogs.New(
		logPath+"-%Y-%m-%d"+ext,
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
		rotatelogs.WithRotationCount(uint(daily.config.GetInt(channel+".days"))),
		rotatelogs.WithClock(rotatelogs.NewClock(carbon.Now().StdTime())),
	)
	if err != nil {
		return nil, err
	}

	level := getLevelFromString(daily.config.GetString(channel + ".level"))

	var writers []io.Writer
	writers = append(writers, writer)

	if daily.config.GetBool(channel + ".print") {
		writers = append(writers, os.Stdout)
	}

	// Create the slog handler
	handler := formatter.NewGeneralHandler(daily.config, daily.json, io.MultiWriter(writers...), level)
	return handler, nil
}
