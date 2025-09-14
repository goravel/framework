package process

import (
	"strings"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

var _ contractsprocess.Result = (*Result)(nil)

type Result struct {
	err      error
	exitCode int
	command  string
	stdout   string
	stderr   string
}

func NewResult(err error, exitCode int, command, stdout, stderr string) *Result {
	return &Result{
		err:      err,
		exitCode: exitCode,
		command:  command,
		stdout:   stdout,
		stderr:   stderr,
	}
}

func (r *Result) Successful() bool {
	if r == nil {
		return false
	}
	return r.exitCode == 0
}

func (r *Result) Failed() bool {
	if r == nil {
		return true
	}
	return r.exitCode != 0
}

func (r *Result) ExitCode() int {
	if r == nil {
		return -1
	}
	return r.exitCode
}

func (r *Result) Output() string {
	if r == nil {
		return ""
	}
	return r.stdout
}

func (r *Result) ErrorOutput() string {
	if r == nil {
		return ""
	}
	return r.stderr
}

func (r *Result) Error() error {
	if r == nil {
		return nil
	}

	return r.err
}

func (r *Result) Command() string {
	if r == nil {
		return ""
	}
	return r.command
}

func (r *Result) SeeInOutput(needle string) bool {
	if r == nil || needle == "" {
		return false
	}
	return strings.Contains(r.stdout, needle)
}

func (r *Result) SeeInErrorOutput(needle string) bool {
	if r == nil || needle == "" {
		return false
	}
	return strings.Contains(r.stderr, needle)
}
