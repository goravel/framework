package log

import "github.com/sirupsen/logrus"

//go:generate mockery --name=Log
type Log interface {
	Testing() Log
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warning(args ...interface{})
	Warningf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
}

type Hook interface {
	Handle(configPath string) (logrus.Hook, error)
}
