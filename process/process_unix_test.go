//go:build !windows

package process

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

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

func TestProcess_Start_ErrorOnMissingCommand_Unix(t *testing.T) {
	_, err := New().Quietly().Start("")
	assert.Error(t, err)
}
