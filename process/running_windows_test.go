//go:build windows

package process

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunning_Basics_Windows(t *testing.T) {
	t.Run("PID, Command, Running, and Wait", func(t *testing.T) {
		r, err := New().Quietly().Start("powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Milliseconds 200; exit 7")
		assert.NoError(t, err)

		run, ok := r.(*Running)
		assert.True(t, ok, "unexpected running type")

		assert.NotEqual(t, 0, run.PID(), "PID should not be zero")
		assert.Contains(t, run.Command(), "powershell", "Command string should be correct")
		assert.True(t, run.Running(), "Process should be running initially")

		// Block and wait for the final result
		res := run.Wait()

		assert.NotNil(t, res, "Result should not be nil")
		assert.False(t, run.Running(), "Process should not be running after Wait")
		assert.Equal(t, 7, res.ExitCode(), "Exit code should be captured correctly")
	})
}

func TestRunning_DoneChannel_Windows(t *testing.T) {
	t.Run("Done channel closes on normal exit", func(t *testing.T) {
		r, err := New().Quietly().Start("powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Milliseconds 200")
		assert.NoError(t, err)
		run, _ := r.(*Running)

		select {
		case <-run.Done():
			res := run.Wait()
			assert.True(t, res.Successful(), "Process should have exited successfully")
		case <-time.After(2 * time.Second):
			t.Fatal("Done channel was not closed within the expected time")
		}
	})

	t.Run("Done channel works with select timeout", func(t *testing.T) {
		r, err := New().Quietly().Start("powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Seconds 5")
		assert.NoError(t, err)
		run, _ := r.(*Running)

		select {
		case <-run.Done():
			t.Fatal("Done channel closed prematurely")
		case <-time.After(100 * time.Millisecond):
			assert.True(t, run.Running(), "Process should still be running")
			assert.NoError(t, run.Stop(500*time.Millisecond))
		}
	})
}

func TestRunning_Stop_Windows(t *testing.T) {
	t.Run("Stop terminates a running process", func(t *testing.T) {
		r, err := New().Quietly().Start("powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Seconds 10")
		assert.NoError(t, err)
		run, _ := r.(*Running)

		// Give it a moment to start.
		time.Sleep(100 * time.Millisecond)
		assert.True(t, run.Running())

		// On Windows, there is no graceful SIGTERM. Stop should forcefully
		// terminate the process. The timeout is less relevant but still part
		// of the method signature.
		err = run.Stop(1 * time.Second)
		assert.NoError(t, err)

		res := run.Wait()
		assert.False(t, res.Successful(), "Process should have been terminated, resulting in failure")

		// A process terminated via TerminateProcess on Windows typically has an exit code of 1.
		assert.Greater(t, res.ExitCode(), 0)
	})
}

func TestRunning_Panic_AppendsToStderr_Windows(t *testing.T) {
	stderr := &bytes.Buffer{}
	// Intentionally pass nil cmd to trigger panic in goroutine
	r := NewRunning(nil, nil, nil, stderr)
	<-r.Done()
	assert.Equal(t, "panic: runtime error: invalid memory address or nil pointer dereference\n", stderr.String())
}