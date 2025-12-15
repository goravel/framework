package logger

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support"
)

type Single struct {
	config config.Config
	json   foundation.Json
}

func NewSingle(config config.Config, json foundation.Json) *Single {
	return &Single{
		config: config,
		json:   json,
	}
}

func (single *Single) Handle(channel string) (slog.Handler, error) {
	logPath := single.config.GetString(channel + ".path")
	if logPath == "" {
		return nil, errors.LogEmptyLogFilePath
	}

	logPath = filepath.Join(support.RelativePath, logPath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	level := getLevelFromString(single.config.GetString(channel + ".level"))

	return NewFileHandler(file, single.config, single.json, level), nil
}

func getLevelFromString(level string) log.Level {
	l, err := log.ParseLevel(level)
	if err != nil {
		return log.DebugLevel
	}
	return l
}
