//go:build windows

package process

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunning_Basics_Windows(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		setup func(p *Process)
		check func(t *testing.T, run *Running)
	}{
		{
			name:  "PID Command Running and Wait",
			args:  []string{"powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Milliseconds 200"},
			setup: func(p *Process) {},
			check: func(t *testing.T, run *Running) {
				assert.NotEqual(t, 0, run.PID())
				assert.Contains(t, run.Command(), "powershell")
				assert.True(t, run.Running())
				res := run.Wait()
				assert.NotNil(t, res)
				assert.Equal(t, 0, res.ExitCode())
			},
		},
		{
			name: "LatestOutput sizes",
			args: []string{
				"powershell", "-NoLogo", "-NoProfile", "-Command",
				"$s='x'*5000; [Console]::Out.Write($s); $e='y'*5000; [Console]::Error.Write($e)",
			},
			setup: func(p *Process) {
				// Generate large stdout/stderr using PowerShell
			},
			check: func(t *testing.T, run *Running) {
				_ = run.Wait()
				// Windows console writes may be buffered differently; assert non-empty
				assert.Greater(t, len(run.Output()), 0)
				assert.Greater(t, len(run.ErrorOutput()), 0)
				assert.LessOrEqual(t, len(run.LatestOutput()), 4096)
				assert.LessOrEqual(t, len(run.LatestErrorOutput()), 4096)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			tt.setup(p)
			r, err := p.Start(context.Background())
			assert.NoError(t, err)
			run, ok := r.(*Running)
			assert.True(t, ok, "unexpected running type")
			tt.check(t, run)
		})
	}
}

func TestRunning_Stop_Windows(t *testing.T) {
	r, err := New().Start("powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Seconds 10")
	assert.NoError(t, err)
	run := r.(*Running)
	// On Windows, Stop is alias for Kill
	assert.NoError(t, run.Stop(100*time.Millisecond))
	res := run.Wait()
	assert.False(t, res.Successful())
}
