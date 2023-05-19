package mock

import (
	"fmt"

	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/support/time"
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

func prefix(model string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	return fmt.Sprintf("[%s] %s.%s: ", timestamp, "test", model)
}
