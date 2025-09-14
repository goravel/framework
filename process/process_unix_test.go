//go:build !windows

package process

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

type fakeSig struct{}

func (fakeSig) String() string { return "fake" }
func (fakeSig) Signal()        {}

func TestProcess_Run_Unix(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		setup    func(p *Process)
		expectOK bool
		check    func(t *testing.T, res *Result)
	}{
		{
			name: "echo to stdout",
			args: []string{"sh", "-c", "printf 'hello'"},
			setup: func(p *Process) {
				p.Quietly()
			},
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "hello", res.Output())
				assert.Equal(t, "", res.ErrorOutput())
				assert.True(t, res.Successful())
			},
		},
		{
			name: "stderr and non-zero exit",
			args: []string{"sh", "-c", "printf 'bad' 1>&2; exit 2"},
			setup: func(p *Process) {
				p.Quietly()
			},
			expectOK: true, // Run doesn't error on non-zero exit
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "bad", res.ErrorOutput())
				assert.Equal(t, "", res.Output())
				assert.False(t, res.Successful())
				assert.Equal(t, 2, res.ExitCode())
			},
		},
		{
			name: "env vars passed to process",
			args: []string{"sh", "-c", "printf \"$FOO\""},
			setup: func(p *Process) {
				p.Env(map[string]string{"FOO": "BAR"}).Quietly()
			},
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "BAR", res.Output())
			},
		},
		{
			name: "working directory changes",
			args: []string{"./script.sh"},
			setup: func(p *Process) {
				dir := t.TempDir()
				path := filepath.Join(dir, "script.sh")
				_ = os.WriteFile(path, []byte("#!/bin/sh\nprintf 'ok'\n"), 0o755)
				_ = os.Chmod(path, 0o755)
				p.Path(dir).Quietly()
			},
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "ok", res.Output())
			},
		},
		{
			name: "stdin is piped",
			args: []string{"sh", "-c", "cat"},
			setup: func(p *Process) {
				p.Input(bytes.NewBufferString("ping")).Quietly()
			},
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "ping", res.Output())
			},
		},
		{
			name: "timeout cancels long-running process",
			args: []string{"sh", "-c", "sleep 1"},
			setup: func(p *Process) {
				p.Timeout(100 * time.Millisecond).Quietly()
			},
			expectOK: true, // Run doesn't error on timeout
			check: func(t *testing.T, res *Result) {
				assert.False(t, res.Successful())
				assert.NotEqual(t, 0, res.ExitCode())
			},
		},
		{
			name: "disable buffering with OnOutput",
			args: []string{"sh", "-c", "printf 'to_stdout'; printf 'to_stderr' 1>&2"},
			setup: func(p *Process) {
				p.DisableBuffering().Quietly().OnOutput(func(typ contractsprocess.OutputType, line []byte) {
					// The handler still works, but the Result buffer is what we're testing.
				})
			},
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "", res.Output())
				assert.Equal(t, "", res.ErrorOutput())
				assert.True(t, res.Successful())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			tt.setup(p)
			res, err := p.Run(tt.args[0], tt.args[1:]...)
			assert.Equal(t, tt.expectOK, err == nil)
			assert.NotNil(t, res)
			r, ok := res.(*Result)
			assert.True(t, ok, "unexpected result type")
			tt.check(t, r)
		})
	}
}

func TestProcess_OnOutput_Callbacks_Unix(t *testing.T) {
	var outLines, errLines [][]byte
	p := New()
	p.OnOutput(func(typ contractsprocess.OutputType, line []byte) {
		if typ == contractsprocess.OutputTypeStdout {
			outLines = append(outLines, append([]byte(nil), line...))
		} else {
			errLines = append(errLines, append([]byte(nil), line...))
		}
	}).Quietly()
	res, err := p.Run("sh", "-c", "printf 'a\n'; printf 'b\n' 1>&2")
	assert.NoError(t, err)
	assert.True(t, res.Successful())
	if assert.NotEmpty(t, outLines) {
		assert.Equal(t, "a", strings.TrimSpace(string(outLines[0])))
	}
	if assert.NotEmpty(t, errLines) {
		assert.Equal(t, "b", strings.TrimSpace(string(errLines[0])))
	}
}

func TestProcess_ErrorOnMissingCommand_Unix(t *testing.T) {
	_, err := New().Quietly().Start("")
	assert.Error(t, err)

	_, err = New().Quietly().Run("")
	assert.Error(t, err)
}

func TestProcess_WithContext(t *testing.T) {
	t.Run("Successful when context is not cancelled", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := New().WithContext(ctx).Quietly().Run("echo", "hello world")

		assert.NoError(t, err, "Run should not return an error on success")
		assert.NotNil(t, res, "Result object should not be nil")
		assert.True(t, res.Successful(), "Process should be successful")
		assert.Equal(t, 0, res.ExitCode(), "Exit code should be 0 for a successful command")
		assert.Contains(t, res.Output(), "hello world", "Output should contain the echoed string")
	})

	t.Run("Terminates process when context times out", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		res, err := New().WithContext(ctx).Quietly().Run("sleep", "2")
		assert.NoError(t, err, "Run should not return an error, but the result should indicate failure")
		assert.NotNil(t, res, "Result object should not be nil even on failure")
		assert.False(t, res.Successful(), "Process should have failed because it was killed")
		assert.NotEqual(t, 0, res.ExitCode(), "Exit code should be non-zero")
	})
}

func TestGetExitCode_Unix(t *testing.T) {
	assert.Equal(t, 0, getExitCode(nil, nil))

	cmd := exec.Command("sh", "-c", "exit 7")
	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, 7, getExitCode(nil, err))

	cmd2 := exec.Command("sh", "-c", "exit 0")
	_ = cmd2.Run()
	assert.Equal(t, 0, getExitCode(cmd2, nil))
}

func TestUnixHelpers_DirectCalls_Unix(t *testing.T) {
	assert.False(t, running(nil))
	assert.Error(t, kill(nil))
	assert.Error(t, signal(nil, fakeSig{}))
	assert.Error(t, stop(nil, make(chan struct{}), 0))
}
