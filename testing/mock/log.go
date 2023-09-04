package mock

import (
	"fmt"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/support/carbon"
)

type TestLog struct {
}

func NewTestLog() log.Writer {
	return &TestLog{}
}

func (r *TestLog) Debug(args ...any) {
	fmt.Print(prefix("debug"))
	fmt.Println(args...)
}

func (r *TestLog) Debugf(format string, args ...any) {
	fmt.Print(prefix("debug"))
	fmt.Printf(format+"\n", args...)
}

func (r *TestLog) Info(args ...any) {
	fmt.Print(prefix("info"))
	fmt.Println(args...)
}

func (r *TestLog) Infof(format string, args ...any) {
	fmt.Print(prefix("info"))
	fmt.Printf(format+"\n", args...)
}

func (r *TestLog) Warning(args ...any) {
	fmt.Print(prefix("warning"))
	fmt.Println(args...)
}

func (r *TestLog) Warningf(format string, args ...any) {
	fmt.Print(prefix("warning"))
	fmt.Printf(format+"\n", args...)
}

func (r *TestLog) Error(args ...any) {
	fmt.Print(prefix("error"))
	fmt.Println(args...)
}

func (r *TestLog) Errorf(format string, args ...any) {
	fmt.Print(prefix("error"))
	fmt.Printf(format+"\n", args...)
}

func (r *TestLog) Fatal(args ...any) {
	fmt.Print(prefix("fatal"))
	fmt.Println(args...)
}

func (r *TestLog) Fatalf(format string, args ...any) {
	fmt.Print(prefix("fatal"))
	fmt.Printf(format+"\n", args...)
}

func (r *TestLog) Panic(args ...any) {
	fmt.Print(prefix("panic"))
	fmt.Println(args...)
}

func (r *TestLog) Panicf(format string, args ...any) {
	fmt.Print(prefix("panic"))
	fmt.Printf(format+"\n", args...)
}

func (r *TestLog) User(user any) log.Writer {
	return r
}

func (r *TestLog) Owner(owner any) log.Writer {
	return r
}

func (r *TestLog) Hint(hint string) log.Writer {
	return r
}

func (r *TestLog) Code(code string) log.Writer {
	return r
}

func (r *TestLog) With(data map[string]any) log.Writer {
	return r
}

func (r *TestLog) Tags(tags ...string) log.Writer {
	return r
}

func (r *TestLog) Request(req http.ContextRequest) log.Writer {
	return r
}

func (r *TestLog) Response(res http.ContextResponse) log.Writer {
	return r
}

func (r *TestLog) In(domain string) log.Writer {
	return r
}

func prefix(model string) string {
	timestamp := carbon.Now().ToDateTimeString()

	return fmt.Sprintf("[%s] %s.%s: ", timestamp, "test", model)
}
