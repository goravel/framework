package process

import (
	"os"
	"strings"
	"time"
)

type Result struct {
	exitCode     int
	command      string
	duration     time.Duration
	processState *os.ProcessState
	stdout       string
	stderr       string
}

func (r *Result) Successful() bool {
	return r.exitCode == 0
}

func (r *Result) Failed() bool {
	return r.exitCode != 0
}

func (r *Result) ExitCode() int {
	return r.exitCode
}

func (r *Result) Output() string {
	return r.stdout
}

func (r *Result) ErrorOutput() string {
	return r.stderr
}

func (r *Result) Command() string {
	return r.command
}

func (r *Result) Duration() time.Duration {
	return r.duration
}

func (r *Result) ProcessState() *os.ProcessState {
	return r.processState
}

func (r *Result) SeeInOutput(needle string) bool {
	return strings.Contains(r.stdout, needle)
}
