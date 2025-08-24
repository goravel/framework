//go:build !windows

package process

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	contractsprocess "github.com/goravel/framework/contracts/process"
	"github.com/stretchr/testify/assert"
)

func TestProcess_Run_Unix(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(p *Process)
		expectOK bool
		check    func(t *testing.T, res *Result)
	}{
		{
			name: "echo to stdout",
			setup: func(p *Process) {
				p.Command("sh", "-c", "printf 'hello'").Quietly()
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
			setup: func(p *Process) {
				p.Command("sh", "-c", "printf 'bad' 1>&2; exit 2").Quietly()
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
			setup: func(p *Process) {
				p.Command("sh", "-c", "printf \"$FOO\"").Env(map[string]string{"FOO": "BAR"}).Quietly()
			},
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "BAR", res.Output())
			},
		},
		{
			name: "working directory changes",
			setup: func(p *Process) {
				dir := t.TempDir()
				path := filepath.Join(dir, "script.sh")
				_ = os.WriteFile(path, []byte("#!/bin/sh\nprintf 'ok'\n"), 0o755)
				_ = os.Chmod(path, 0o755)
				p.Path(dir).Quietly().Command("./script.sh")
			},
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "ok", res.Output())
			},
		},
		{
			name: "stdin is piped",
			setup: func(p *Process) {
				p.Command("sh", "-c", "cat").Input(bytes.NewBufferString("ping")).Quietly()
			},
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "ping", res.Output())
			},
		},
		{
			name: "timeout cancels long-running process",
			setup: func(p *Process) {
				p.Command("sh", "-c", "sleep 1").Timeout(100 * time.Millisecond).Quietly()
			},
			expectOK: true, // Run doesn't error on timeout
			check: func(t *testing.T, res *Result) {
				assert.False(t, res.Successful())
				assert.NotEqual(t, 0, res.ExitCode())
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			tt.setup(p)
			res, err := p.Run(context.Background())
			if tt.expectOK {
				assert.NoError(t, err)
			}
			assert.NotNil(t, res)
			r, ok := res.(*Result)
			assert.True(t, ok, "unexpected result type")
			if ok && tt.check != nil {
				tt.check(t, r)
			}
		})
	}
}

func TestProcess_OnOutput_Callbacks_Unix(t *testing.T) {
	var outLines, errLines [][]byte
	p := New().Command("sh", "-c", "printf 'a\n'; printf 'b\n' 1>&2")
	p.OnOutput(func(typ contractsprocess.OutputType, line []byte) {
		if typ == contractsprocess.OutputTypeStdout {
			outLines = append(outLines, append([]byte(nil), line...))
		} else {
			errLines = append(errLines, append([]byte(nil), line...))
		}
	}).Quietly()
	res, err := p.Run(context.Background())
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
	_, err := New().Command("").Start(context.Background())
	assert.Error(t, err)
}
