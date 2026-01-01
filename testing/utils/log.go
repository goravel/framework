package utils

import (
	"context"
	"fmt"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/support/carbon"
)

var _ log.Log = &TestLog{}

type TestLog struct {
	*TestLogWriter
}

func NewTestLog() log.Log {
	return &TestLog{
		TestLogWriter: NewTestLogWriter(),
	}
}

func (r *TestLog) WithContext(ctx context.Context) log.Log {
	return r
}

func (r *TestLog) Channel(channel string) log.Log {
	return r
}

func (r *TestLog) Stack(channels []string) log.Log {
	return r
}

type TestLogWriter struct {
	data map[string]any
}

func NewTestLogWriter() *TestLogWriter {
	return &TestLogWriter{
		data: make(map[string]any),
	}
}

func (r *TestLogWriter) Debug(args ...any) {
	fmt.Print(prefix("debug"))
	fmt.Println(args...)
	r.printData()
}

func (r *TestLogWriter) Debugf(format string, args ...any) {
	fmt.Print(prefix("debug"))
	fmt.Printf(format+"\n", args...)
	r.printData()
}

func (r *TestLogWriter) Info(args ...any) {
	fmt.Print(prefix("info"))
	fmt.Println(args...)
	r.printData()
}

func (r *TestLogWriter) Infof(format string, args ...any) {
	fmt.Print(prefix("info"))
	fmt.Printf(format+"\n", args...)
	r.printData()
}

func (r *TestLogWriter) Warning(args ...any) {
	fmt.Print(prefix("warning"))
	fmt.Println(args...)
	r.printData()
}

func (r *TestLogWriter) Warningf(format string, args ...any) {
	fmt.Print(prefix("warning"))
	fmt.Printf(format+"\n", args...)
	r.printData()
}

func (r *TestLogWriter) Error(args ...any) {
	fmt.Print(prefix("error"))
	fmt.Println(args...)
	r.printData()
}

func (r *TestLogWriter) Errorf(format string, args ...any) {
	fmt.Print(prefix("error"))
	fmt.Printf(format+"\n", args...)
	r.printData()
}

func (r *TestLogWriter) Fatal(args ...any) {
	fmt.Print(prefix("fatal"))
	fmt.Println(args...)
	r.printData()
}

func (r *TestLogWriter) Fatalf(format string, args ...any) {
	fmt.Print(prefix("fatal"))
	fmt.Printf(format+"\n", args...)
	r.printData()
}

func (r *TestLogWriter) Panic(args ...any) {
	fmt.Print(prefix("panic"))
	fmt.Println(args...)
	r.printData()
}

func (r *TestLogWriter) Panicf(format string, args ...any) {
	fmt.Print(prefix("panic"))
	fmt.Printf(format+"\n", args...)
	r.printData()
}

func (r *TestLogWriter) User(user any) log.Writer {
	r.data["user"] = user

	return r
}

func (r *TestLogWriter) Owner(owner any) log.Writer {
	r.data["owner"] = owner

	return r
}

func (r *TestLogWriter) Hint(hint string) log.Writer {
	r.data["hint"] = hint

	return r
}

func (r *TestLogWriter) Code(code string) log.Writer {
	r.data["code"] = code

	return r
}

func (r *TestLogWriter) With(data map[string]any) log.Writer {
	r.data["with"] = data

	return r
}

func (r *TestLogWriter) Tags(tags ...string) log.Writer {
	r.data["tags"] = tags

	return r
}

func (r *TestLogWriter) WithTrace() log.Writer {
	return r
}

func (r *TestLogWriter) Request(req http.ContextRequest) log.Writer {
	r.data["request"] = req

	return r
}

func (r *TestLogWriter) Response(res http.ContextResponse) log.Writer {
	r.data["response"] = res

	return r
}

func (r *TestLogWriter) In(domain string) log.Writer {
	r.data["in"] = domain

	return r
}

func (r *TestLogWriter) printData() {
	if len(r.data) > 0 {
		fmt.Println(r.data)
	}
}

func prefix(model string) string {
	timestamp := carbon.Now().ToDateTimeString()

	return fmt.Sprintf("[%s] %s.%s: ", timestamp, "test", model)
}
