package log

import (
	"fmt"

	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/support/time"
)

type TestWriter struct {
}

func NewTestWriter() log.Writer {
	return &TestWriter{}
}

func (r *TestWriter) Debug(args ...any) {
	fmt.Print(prefix("debug"))
	fmt.Println(args...)
}

func (r *TestWriter) Debugf(format string, args ...any) {
	fmt.Print(prefix("debug"))
	fmt.Printf(format+"\n", args...)
}

func (r *TestWriter) Info(args ...any) {
	fmt.Print(prefix("info"))
	fmt.Println(args...)
}

func (r *TestWriter) Infof(format string, args ...any) {
	fmt.Print(prefix("info"))
	fmt.Printf(format+"\n", args...)
}

func (r *TestWriter) Warning(args ...any) {
	fmt.Print(prefix("warning"))
	fmt.Println(args...)
}

func (r *TestWriter) Warningf(format string, args ...any) {
	fmt.Print(prefix("warning"))
	fmt.Printf(format+"\n", args...)
}

func (r *TestWriter) Error(args ...any) {
	fmt.Print(prefix("error"))
	fmt.Println(args...)
}

func (r *TestWriter) Errorf(format string, args ...any) {
	fmt.Print(prefix("error"))
	fmt.Printf(format+"\n", args...)
}

func (r *TestWriter) Fatal(args ...any) {
	fmt.Print(prefix("fatal"))
	fmt.Println(args...)
}

func (r *TestWriter) Fatalf(format string, args ...any) {
	fmt.Print(prefix("fatal"))
	fmt.Printf(format+"\n", args...)
}

func (r *TestWriter) Panic(args ...any) {
	fmt.Print(prefix("panic"))
	fmt.Println(args...)
}

func (r *TestWriter) Panicf(format string, args ...any) {
	fmt.Print(prefix("panic"))
	fmt.Printf(format+"\n", args...)
}

func prefix(model string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	return fmt.Sprintf("[%s] %s.%s: ", timestamp, "test", model)
}
