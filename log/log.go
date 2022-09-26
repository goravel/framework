package log

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/goravel/framework/contracts/log"
)

type Log struct {
	Instance *logrus.Logger
	Test     bool
}

func (r *Log) Testing() log.Log {
	r.Test = true

	return r
}

func (r *Log) Debug(args ...interface{}) {
	if r.Test {
		fmt.Print("Debug: ")
		fmt.Println(args...)
		return
	}

	r.Instance.Debug(args...)
}

func (r *Log) Debugf(format string, args ...interface{}) {
	if r.Test {
		fmt.Print("Debugf: ")
		fmt.Printf(format+"\n", args...)
		return
	}

	r.Instance.Debugf(format, args...)
}

func (r *Log) Info(args ...interface{}) {
	if r.Test {
		fmt.Print("Info: ")
		fmt.Println(args...)
		return
	}

	r.Instance.Info(args...)
}

func (r *Log) Infof(format string, args ...interface{}) {
	if r.Test {
		fmt.Print("Infof: ")
		fmt.Printf(format+"\n", args...)
		return
	}

	r.Instance.Infof(format, args...)
}

func (r *Log) Warning(args ...interface{}) {
	if r.Test {
		fmt.Print("Warningf: ")
		fmt.Println(args...)
		return
	}

	r.Instance.Warning(args...)
}

func (r *Log) Warningf(format string, args ...interface{}) {
	if r.Test {
		fmt.Print("Warningf: ")
		fmt.Printf(format+"\n", args...)
		return
	}

	r.Instance.Warningf(format, args...)
}

func (r *Log) Error(args ...interface{}) {
	if r.Test {
		fmt.Print("Error: ")
		fmt.Println(args...)
		return
	}

	r.Instance.Error(args...)
}

func (r *Log) Errorf(format string, args ...interface{}) {
	if r.Test {
		fmt.Print("Errorf: ")
		fmt.Printf(format+"\n", args...)
		return
	}

	r.Instance.Errorf(format, args...)
}

func (r *Log) Panic(args ...interface{}) {
	if r.Test {
		fmt.Print("Panic: ")
		fmt.Println(args...)
		return
	}

	r.Instance.Panic(args...)
}

func (r *Log) Panicf(format string, args ...interface{}) {
	if r.Test {
		fmt.Print("Panicf: ")
		fmt.Printf(format+"\n", args...)
		return
	}

	r.Instance.Panicf(format, args...)
}
