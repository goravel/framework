package queue

import (
	"github.com/goravel/framework/contracts/log"
)

type Debug struct {
	debug bool
	log   log.Log
}

func NewDebug(debug bool, log log.Log) *Debug {
	return &Debug{
		debug: debug,
		log:   log,
	}
}

func (r *Debug) Print(args ...any) {
	if r.debug {
		r.log.Debug(args...)
	}
}

func (r *Debug) Printf(format string, args ...any) {
	if r.debug {
		r.log.Debugf(format, args...)
	}
}

func (r *Debug) Println(args ...any) {
	if r.debug {
		r.log.Debug(args...)
	}
}

func (r *Debug) Fatal(args ...any) {
	r.log.Error(args...)
}

func (r *Debug) Fatalf(format string, args ...any) {
	r.log.Errorf(format, args...)
}

func (r *Debug) Fatalln(args ...any) {
	r.log.Error(args...)
}

func (r *Debug) Panic(args ...any) {
	r.log.Panic(args...)
}

func (r *Debug) Panicf(format string, args ...any) {
	r.log.Panicf(format, args...)
}

func (r *Debug) Panicln(args ...any) {
	r.log.Panic(args...)
}

type Info struct {
	debug bool
	log   log.Log
}

func NewInfo(debug bool, log log.Log) *Info {
	return &Info{
		debug: debug,
		log:   log,
	}
}

func (r *Info) Print(args ...any) {
	if r.debug {
		r.log.Info(args...)
	}
}

func (r *Info) Printf(format string, args ...any) {
	if r.debug {
		r.log.Infof(format, args...)
	}
}

func (r *Info) Println(args ...any) {
	if r.debug {
		r.log.Info(args...)
	}
}

func (r *Info) Fatal(args ...any) {
	r.log.Error(args...)
}

func (r *Info) Fatalf(format string, args ...any) {
	r.log.Errorf(format, args...)
}

func (r *Info) Fatalln(args ...any) {
	r.log.Error(args...)
}

func (r *Info) Panic(args ...any) {
	r.log.Panic(args...)
}

func (r *Info) Panicf(format string, args ...any) {
	r.log.Panicf(format, args...)
}

func (r *Info) Panicln(args ...any) {
	r.log.Panic(args...)
}

type Warning struct {
	debug bool
	log   log.Log
}

func NewWarning(debug bool, log log.Log) *Warning {
	return &Warning{
		debug: debug,
		log:   log,
	}
}

func (r *Warning) Print(args ...any) {
	r.log.Warning(args...)
}

func (r *Warning) Printf(format string, args ...any) {
	r.log.Warningf(format, args...)
}

func (r *Warning) Println(args ...any) {
	r.log.Warning(args...)
}

func (r *Warning) Fatal(args ...any) {
	r.log.Error(args...)
}

func (r *Warning) Fatalf(format string, args ...any) {
	r.log.Errorf(format, args...)
}

func (r *Warning) Fatalln(args ...any) {
	r.log.Error(args...)
}

func (r *Warning) Panic(args ...any) {
	r.log.Panic(args...)
}

func (r *Warning) Panicf(format string, args ...any) {
	r.log.Panicf(format, args...)
}

func (r *Warning) Panicln(args ...any) {
	r.log.Panic(args...)
}

type Error struct {
	debug bool
	log   log.Log
}

func NewError(debug bool, log log.Log) *Error {
	return &Error{
		debug: debug,
		log:   log,
	}
}

func (r *Error) Print(args ...any) {
	r.log.Error(args...)
}

func (r *Error) Printf(format string, args ...any) {
	r.log.Errorf(format, args...)
}

func (r *Error) Println(args ...any) {
	r.log.Error(args...)
}

func (r *Error) Fatal(args ...any) {
	r.log.Error(args...)
}

func (r *Error) Fatalf(format string, args ...any) {
	r.log.Errorf(format, args...)
}

func (r *Error) Fatalln(args ...any) {
	r.log.Error(args...)
}

func (r *Error) Panic(args ...any) {
	r.log.Panic(args...)
}

func (r *Error) Panicf(format string, args ...any) {
	r.log.Panicf(format, args...)
}

func (r *Error) Panicln(args ...any) {
	r.log.Panic(args...)
}

type Fatal struct {
	debug bool
	log   log.Log
}

func NewFatal(debug bool, log log.Log) *Fatal {
	return &Fatal{
		debug: debug,
		log:   log,
	}
}

func (r *Fatal) Print(args ...any) {
	r.log.Fatal(args...)
}

func (r *Fatal) Printf(format string, args ...any) {
	r.log.Fatalf(format, args...)
}

func (r *Fatal) Println(args ...any) {
	r.log.Fatal(args...)
}

func (r *Fatal) Fatal(args ...any) {
	r.log.Fatal(args...)
}

func (r *Fatal) Fatalf(format string, args ...any) {
	r.log.Fatalf(format, args...)
}

func (r *Fatal) Fatalln(args ...any) {
	r.log.Fatal(args...)
}

func (r *Fatal) Panic(args ...any) {
	r.log.Panic(args...)
}

func (r *Fatal) Panicf(format string, args ...any) {
	r.log.Panicf(format, args...)
}

func (r *Fatal) Panicln(args ...any) {
	r.log.Panic(args...)
}
