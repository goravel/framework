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
)

func TestProcess_Run_Windows(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		setup    func(p *Process)
		expectOK bool
		check    func(t *testing.T, res *Result)
	}{
		{
			name: "echo via cmd",
			args: []string{"cmd", "/C", "echo hello"},
			setup: func(p *Process) {
				p.Quietly()
			},
			expectOK: true,
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
			expectOK: true, // Run doesn't error on non-zero exit
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
			expectOK: true,
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
			expectOK: true,
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
			expectOK: true, // Run doesn't error on timeout
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
			r, ok := res.(*Result)
			assert.True(t, ok, "unexpected result type")
			tt.check(t, r)
		})
	}
}
