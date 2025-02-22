package logger

import (
	"context"
	"net"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	gormlogger "gorm.io/gorm/logger"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/logger"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/support/carbon"
)

func NewLogger(config config.Config, log log.Log) logger.Logger {
	level := logger.Warn
	if config.GetBool("app.debug") {
		level = logger.Info
	}

	slowThreshold := config.GetInt("database.slow_threshold", 200)
	if slowThreshold <= 0 {
		slowThreshold = 200
	}

	return &Logger{
		log:           log,
		level:         level,
		slowThreshold: time.Duration(slowThreshold) * time.Millisecond,
	}
}

type Logger struct {
	log           log.Log
	level         logger.Level
	slowThreshold time.Duration
}

func (r *Logger) Level(level logger.Level) logger.Logger {
	r.level = level

	return r
}

func (r *Logger) Info(ctx context.Context, msg string, data ...any) {
	if r.level >= logger.Info {
		r.log.Infof(msg, data...)
	}
}

func (r *Logger) Warn(ctx context.Context, msg string, data ...any) {
	if r.level >= logger.Warn {
		r.log.Warningf(msg, data...)
	}
}

func (r *Logger) Error(ctx context.Context, msg string, data ...any) {
	// Let upper layer function deals with connection refused error
	var cancel bool
	for _, item := range data {
		if tempItem, ok := item.(*net.OpError); ok {
			if strings.Contains(tempItem.Error(), "connection refused") {
				return
			}

		}
		if tempItem, ok := item.(error); ok {
			// Avoid duplicate output
			if strings.Contains(tempItem.Error(), "Access denied") {
				cancel = true
			}
		}
	}

	if cancel {
		return
	}

	if r.level >= logger.Error {
		r.log.Errorf(msg, data...)
	}
}

func (r *Logger) Trace(ctx context.Context, begin carbon.Carbon, sql string, rowsAffected int64, err error) {
	if r.level <= logger.Silent {
		return
	}

	var (
		traceStr     = "[%.3fms] [rows:%v] %s"
		traceWarnStr = "[%.3fms] [rows:%v] [SLOW] %s"
		traceErrStr  = "[%.3fms] [rows:%v] %s\t%s"
	)

	elapsed := begin.DiffInDuration()

	switch {
	case err != nil && r.level >= logger.Error:
		if rowsAffected == -1 {
			r.log.Errorf(traceErrStr, float64(elapsed.Nanoseconds())/1e6, "-", sql, err)
		} else {
			r.log.Errorf(traceErrStr, float64(elapsed.Nanoseconds())/1e6, rowsAffected, sql, err)
		}
	case elapsed > r.slowThreshold && r.slowThreshold != 0 && r.level >= logger.Warn:
		if rowsAffected == -1 {
			r.log.Warningf(traceWarnStr, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			r.log.Warningf(traceWarnStr, float64(elapsed.Nanoseconds())/1e6, rowsAffected, sql)
		}
	case r.level == logger.Info:
		if rowsAffected == -1 {
			r.log.Infof(traceStr, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			r.log.Infof(traceStr, float64(elapsed.Nanoseconds())/1e6, rowsAffected, sql)
		}
	}
}

func (r *Logger) ToGorm() gormlogger.Interface {
	return NewGorm(r)
}

type Gorm struct {
	logger logger.Logger
}

func NewGorm(logger logger.Logger) *Gorm {
	return &Gorm{
		logger: logger,
	}
}

func (r *Gorm) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	_ = r.logger.Level(GormLevelToLevel(level))

	return r
}

func (r *Gorm) Info(ctx context.Context, msg string, data ...any) {
	r.logger.Info(ctx, msg, data...)
}

func (r *Gorm) Warn(ctx context.Context, msg string, data ...any) {
	r.logger.Warn(ctx, msg, data...)
}

func (r *Gorm) Error(ctx context.Context, msg string, data ...any) {
	r.logger.Error(ctx, msg, data...)
}

func (r *Gorm) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rowsAffected := fc()
	r.logger.Trace(ctx, carbon.FromStdTime(begin), sql, rowsAffected, err)
}

// FileWithLineNum return the file name and line number of the current file
func FileWithLineNum() string {
	_, file, _, _ := runtime.Caller(0)
	gormSourceDir := regexp.MustCompile(`utils.utils\.go`).ReplaceAllString(file, "")
	goravelSourceDir := "database/gorm.go"

	// the second caller usually from gorm internal, so set i start from 5
	for i := 5; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && ((!strings.HasPrefix(file, gormSourceDir) && !strings.Contains(file, goravelSourceDir)) || strings.HasSuffix(file, "_test.go")) {
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	return ""
}

func GormLevelToLevel(level gormlogger.LogLevel) logger.Level {
	switch level {
	case gormlogger.Silent:
		return logger.Silent
	case gormlogger.Error:
		return logger.Error
	case gormlogger.Warn:
		return logger.Warn
	default:
		return logger.Info
	}
}
