//go:build windows

package process

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunning_Basics_Windows(t *testing.T) {
	r, err := New().Command("powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Milliseconds 200").Start(context.Background())
	assert.NoError(t, err)
	run := r.(*Running)
	assert.NotEqual(t, 0, run.PID())
	assert.Contains(t, run.Command(), "powershell")
	assert.True(t, run.Running())
	res := run.Wait()
	assert.NotNil(t, res)
}

func TestRunning_LatestOutput_Windows(t *testing.T) {
	// Generate large stdout/stderr using PowerShell
	script := "$s='x'*5000; [Console]::Out.Write($s); $e='y'*5000; [Console]::Error.Write($e)"
	r, err := New().Command("powershell", "-NoLogo", "-NoProfile", "-Command", script).Start(context.Background())
	assert.NoError(t, err)
	run := r.(*Running)
	_ = run.Wait()
	// Windows console writes may be buffered differently; assert non-empty
	assert.Greater(t, len(run.Output()), 0)
	assert.Greater(t, len(run.ErrorOutput()), 0)
}

func TestRunning_Stop_Windows(t *testing.T) {
	r, err := New().Command("powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Seconds 10").Start(context.Background())
	assert.NoError(t, err)
	run := r.(*Running)
	// On Windows, Stop is alias for Kill
	assert.NoError(t, run.Stop(100*time.Millisecond))
}
