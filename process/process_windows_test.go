//go:build windows

package process

import (
	"bytes"
	"context"
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
		setup    func(p *Process)
		expectOK bool
		check    func(t *testing.T, res *Result)
	}{
		{
			name: "echo via cmd",
			setup: func(p *Process) {
				p.Command("cmd", "/C", "echo hello").Quietly()
			},
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "hello\r\n", res.Output())
				assert.True(t, res.Successful())
			},
		},
		{
			name: "stderr and non-zero",
			setup: func(p *Process) {
				// powershell: write-error writes to stderr and returns non-zero
				p.Command("powershell", "-NoLogo", "-NoProfile", "-Command", "Write-Error 'bad'; exit 2").Quietly()
			},
			expectOK: false,
			check: func(t *testing.T, res *Result) {
				assert.Contains(t, res.ErrorOutput(), "bad")
				assert.False(t, res.Successful())
				assert.NotEqual(t, 0, res.ExitCode())
			},
		},
		{
			name: "working directory changes",
			setup: func(p *Process) {
				dir := t.TempDir()
				path := filepath.Join(dir, "script.bat")
				_ = os.WriteFile(path, []byte("@echo off\r\necho ok\r\n"), 0644)
				p.Path(dir).Quietly().Command("cmd", "/C", "script.bat")
			},
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "ok\r\n", res.Output())
			},
		},
		{
			name: "stdin is piped",
			setup: func(p *Process) {
				p.Command("cmd", "/C", "more").Input(bytes.NewBufferString("ping\r\n")).Quietly()
			},
			expectOK: true,
			check: func(t *testing.T, res *Result) {
				assert.Equal(t, "ping", strings.TrimSpace(res.Output()))
			},
		},
		{
			name: "timeout cancels long-running process",
			setup: func(p *Process) {
				p.Command("powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Seconds 2").Timeout(200 * time.Millisecond).Quietly()
			},
			expectOK: false,
			check: func(t *testing.T, res *Result) {
				assert.False(t, res.Successful())
				assert.NotEqual(t, 0, res.ExitCode())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			if tt.setup != nil {
				tt.setup(p)
			}
			res, err := p.Run(context.Background())
			if tt.expectOK {
				assert.NoError(t, err)
			}
			assert.NotNil(t, res)
			if tt.check != nil {
				r, ok := res.(*Result)
				assert.True(t, ok, "unexpected result type")
				if ok {
					tt.check(t, r)
				}
			}
		})
	}
}
