package process

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	contractsprocess "github.com/goravel/framework/contracts/process"
	"github.com/stretchr/testify/assert"
)

func TestProcess_Run_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		cmd      []string
		setup    func(p *Process)
		expectOK bool
		check    func(t *testing.T, res *Result)
	}{
		{
			name:     "successful echo to stdout",
			cmd:      []string{"sh", "-c", "printf 'hello'"},
			setup:    func(p *Process) { p.Quietly() },
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "hello", res.Output())
				assert.Equal(t, "", res.ErrorOutput())
				assert.True(t, res.Successful())
			},
		},
		{
			name:     "failure with stderr",
			cmd:      []string{"sh", "-c", "printf 'bad' 1>&2; exit 2"},
			setup:    func(p *Process) { p.Quietly() },
			expectOK: false,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "bad", res.ErrorOutput())
				assert.Equal(t, "", res.Output())
				assert.False(t, res.Successful())
				assert.NotEqual(t, 0, res.ExitCode())
			},
		},
		{
			name:     "env is passed to process",
			cmd:      []string{"sh", "-c", "printf \"$FOO\""},
			setup:    func(p *Process) { p.Env(map[string]string{"FOO": "BAR"}).Quietly() },
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "BAR", res.Output())
			},
		},
		{
			name: "path changes working directory",
			cmd:  []string{"./script.sh"},
			setup: func(p *Process) {
				dir := t.TempDir()
				path := filepath.Join(dir, "script.sh")
				if err := os.WriteFile(path, []byte("#!/bin/sh\nprintf 'ok'\n"), 0o755); err != nil {
					assert.Failf(t, "write script", "%v", err)
				}
				if err := os.Chmod(path, 0o755); err != nil {
					assert.Failf(t, "chmod", "%v", err)
				}
				p.Path(dir).Quietly()
			},
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "ok", res.Output())
			},
		},
		{
			name:     "stdin is piped when Input provided",
			cmd:      []string{"sh", "-c", "cat"},
			setup:    func(p *Process) { p.Input(bytes.NewBufferString("ping")).Quietly() },
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "ping", res.Output())
			},
		},
		{
			name:     "timeout cancels long-running process",
			cmd:      []string{"sh", "-c", "sleep 1"},
			setup:    func(p *Process) { p.Timeout(100 * time.Millisecond).Quietly() },
			expectOK: false,
			check: func(t *testing.T, res *Result) {
				assert.False(t, res.Successful())
				assert.NotEqual(t, 0, res.ExitCode())
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := New()
			if len(test.cmd) == 0 {
				t.Fatal("missing command")
			}
			p.Command(test.cmd[0], test.cmd[1:]...)
			if test.setup != nil {
				test.setup(p)
			}
			res, err := p.Run(context.Background())
			if test.expectOK {
				assert.NoError(t, err)
			}
			assert.NotNil(t, res)
			if test.check != nil {
				ttRes, ok := res.(*Result)
				assert.True(t, ok, "unexpected result type")
				if ok {
					test.check(t, ttRes)
				}
			}
		})
	}
}

func TestProcess_Start_ErrorOnMissingCommand(t *testing.T) {
	p := New().Command("")
	_, err := p.Start(context.Background())
	assert.Error(t, err)
}

func TestProcess_OnOutput_Callbacks(t *testing.T) {
	var outLines, errLines [][]byte
	p := New().Command("sh", "-c", "printf 'a\n'; printf 'b\n' 1>&2")
	p.OnOutput(func(typ contractsprocess.OutputType, line []byte) {
		if typ == contractsprocess.OutputTypeStdout {
			outLines = append(outLines, append([]byte(nil), line...))
		} else {
			errLines = append(errLines, append([]byte(nil), line...))
		}
	})
	res, err := p.Run(context.Background())
	assert.NoError(t, err)
	assert.True(t, res.Successful())
	if assert.NotEmpty(t, outLines) {
		assert.Equal(t, "a", string(outLines[0]))
	}
	if assert.NotEmpty(t, errLines) {
		assert.Equal(t, "b", string(errLines[0]))
	}
}
