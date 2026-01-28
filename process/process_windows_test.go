//go:build windows

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
	"github.com/goravel/framework/errors"
)

func TestProcess_Run_Windows(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		setup func(p *Process)
		check func(t *testing.T, res *Result)
	}{
		{
			name: "echo via cmd",
			args: []string{"cmd", "/C", "echo hello"},
			setup: func(p *Process) {
				p.Quietly()
			},
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "hello\r\n", res.Output())
				assert.True(t, res.Successful())
			},
		},
		{
			name: "stderr and non-zero",
			args: []string{"powershell", "-NoLogo", "-NoProfile", "-Command", "Write-Error 'bad'; exit 2"},
			setup: func(p *Process) {
				// powershell: write-error writes to stderr and returns non-zero
				p.Quietly()
			},
			check: func(t *testing.T, res *Result) {
				assert.Contains(t, res.ErrorOutput(), "bad")
				assert.False(t, res.Successful())
				assert.NotEqual(t, 0, res.ExitCode())
			},
		},
		{
			name: "working directory changes",
			args: []string{"cmd", "/C", "script.bat"},
			setup: func(p *Process) {
				dir := t.TempDir()
				path := filepath.Join(dir, "script.bat")
				_ = os.WriteFile(path, []byte("@echo off\r\necho ok\r\n"), 0644)
				p.Path(dir).Quietly()
			},
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "ok\r\n", res.Output())
			},
		},
		{
			name: "stdin is piped",
			args: []string{"cmd", "/C", "more"},
			setup: func(p *Process) {
				p.Input(bytes.NewBufferString("ping\r\n")).Quietly()
			},
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "ping", strings.TrimSpace(res.Output()))
			},
		},
		{
			name: "timeout cancels long-running process",
			args: []string{"powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Seconds 2"},
			setup: func(p *Process) {
				p.Timeout(200 * time.Millisecond).Quietly()
			},
			check: func(t *testing.T, res *Result) {
				assert.False(t, res.Successful())
				assert.NotEqual(t, 0, res.ExitCode())
			},
		},
		{
			name: "disable buffering",
			args: []string{"cmd", "/C", "echo to_stdout & echo to_stderr >&2"},
			setup: func(p *Process) {
				p.DisableBuffering().Quietly()
			},
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
			res := p.Run(tt.args[0], tt.args[1:]...)
			r, ok := res.(*Result)
			assert.True(t, ok)
			tt.check(t, r)
		})
	}
}

func TestProcess_Pool_Windows(t *testing.T) {
	t.Run("creates pool builder and executes commands", func(t *testing.T) {
		p := New()
		results, err := p.Pool(func(pool contractsprocess.Pool) {
			pool.Command("cmd", "/C", "echo hello").As("hello")
			pool.Command("cmd", "/C", "echo world").As("world")
		}).Run()

		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Contains(t, results["hello"].Output(), "hello")
		assert.Contains(t, results["world"].Output(), "world")
	})

	t.Run("returns error with nil configurer", func(t *testing.T) {
		p := New()
		_, err := p.Pool(nil).Run()
		assert.ErrorIs(t, err, errors.ProcessPoolNilConfigurer)
	})
}

func TestProcess_Pipe_Windows(t *testing.T) {
	t.Run("creates pipeline and executes commands", func(t *testing.T) {
		p := New()
		result := p.Pipe(func(pipe contractsprocess.Pipe) {
			pipe.Command("cmd", "/C", "echo hello")
			pipe.Command("findstr", "hello")
		}).Run()

		assert.Contains(t, result.Output(), "hello")
	})

	t.Run("returns error with nil configurer", func(t *testing.T) {
		p := New()
		res := p.Pipe(nil).Run()
		assert.Equal(t, errors.ProcessPipeNilConfigurer, res.Error())
	})
}

func TestFormatCommand_Windows(t *testing.T) {
	tests := []struct {
		name         string
		inputName    string
		inputArgs    []string
		expectedName string
		expectedArgs []string
	}{
		{
			name:         "command with args - not wrapped",
			inputName:    "echo",
			inputArgs:    []string{"hello", "world"},
			expectedName: "echo",
			expectedArgs: []string{"hello", "world"},
		},
		{
			name:         "simple command - not wrapped",
			inputName:    "dir",
			inputArgs:    []string{},
			expectedName: "dir",
			expectedArgs: []string{},
		},
		{
			name:         "command with space only - wrapped",
			inputName:    "echo hello",
			inputArgs:    []string{},
			expectedName: "cmd",
			expectedArgs: []string{"/c", "echo hello"},
		},
		{
			name:         "command with space and ampersand but has args - not wrapped",
			inputName:    "timeout 5 &",
			inputArgs:    []string{"/nobreak"},
			expectedName: "timeout 5 &",
			expectedArgs: []string{"/nobreak"},
		},
		{
			name:         "background command - wrapped",
			inputName:    "timeout 5 &",
			inputArgs:    []string{},
			expectedName: "cmd",
			expectedArgs: []string{"/c", "timeout 5 &"},
		},
		{
			name:         "piped command - wrapped",
			inputName:    "type file.txt | findstr test",
			inputArgs:    []string{},
			expectedName: "cmd",
			expectedArgs: []string{"/c", "type file.txt | findstr test"},
		},
		{
			name:         "piped background command - wrapped",
			inputName:    "type file.txt | findstr test &",
			inputArgs:    []string{},
			expectedName: "cmd",
			expectedArgs: []string{"/c", "type file.txt | findstr test &"},
		},
		{
			name:         "logical AND operators - wrapped",
			inputName:    "echo hello && echo world",
			inputArgs:    []string{},
			expectedName: "cmd",
			expectedArgs: []string{"/c", "echo hello && echo world"},
		},
		{
			name:         "multiple pipes - wrapped",
			inputName:    "type file | sort | findstr test",
			inputArgs:    []string{},
			expectedName: "cmd",
			expectedArgs: []string{"/c", "type file | sort | findstr test"},
		},
		{
			name:         "command with ampersand only - wrapped",
			inputName:    "test&",
			inputArgs:    []string{},
			expectedName: "cmd",
			expectedArgs: []string{"/c", "test&"},
		},
		{
			name:         "command with pipe only - wrapped",
			inputName:    "test|",
			inputArgs:    []string{},
			expectedName: "cmd",
			expectedArgs: []string{"/c", "test|"},
		},
		{
			name:         "single ampersand - wrapped",
			inputName:    "&",
			inputArgs:    []string{},
			expectedName: "cmd",
			expectedArgs: []string{"/c", "&"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotArgs := formatCommand(tt.inputName, tt.inputArgs)
			assert.Equal(t, tt.expectedName, gotName)
			assert.Equal(t, tt.expectedArgs, gotArgs)
		})
	}
}
