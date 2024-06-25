package schedule

import (
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/support/color"
)

type Logger struct {
	log     log.Log
	logInfo bool
}

func NewLogger(log log.Log, logInfo bool) *Logger {
	return &Logger{
		logInfo: logInfo,
		log:     log,
	}
}

func (log *Logger) Info(msg string, keysAndValues ...any) {
	if !log.logInfo {
		return
	}
	color.Green().Printf("%s %v\n", msg, keysAndValues)
}

func (log *Logger) Error(err error, msg string, keysAndValues ...any) {
	log.log.Error(msg, keysAndValues)
}
