//go:build !windows

package process

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

func TestRunning_Basics_Unix(t *testing.T) {
	t.Run("PID, Command, Running, and Wait", func(t *testing.T) {
		r, err := New().Quietly().Start("sh", "-c", "sleep 0.2; exit 5")
		assert.NoError(t, err)

		run, ok := r.(*Running)
		assert.True(t, ok, "unexpected running type")

		assert.NotEqual(t, 0, run.PID(), "PID should not be zero")
		assert.Contains(t, run.Command(), "sleep 0.2", "Command string should be correct")
		assert.True(t, run.Running(), "Process should be running initially")

		res := run.Wait()

		assert.NotNil(t, res, "Result should not be nil")
		assert.False(t, run.Running(), "Process should not be running after Wait")
		assert.Equal(t, 5, res.ExitCode(), "Exit code should be captured correctly")
	})
}

func TestRunning_DoneChannel_Unix(t *testing.T) {
	t.Run("Done channel closes on normal exit", func(t *testing.T) {
		r, err := New().Quietly().Start("sleep", "0.2")
		assert.NoError(t, err)
		run, _ := r.(*Running)

		select {
		case <-run.Done():
			res := run.Wait()
			assert.True(t, res.Successful(), "Process should have exited successfully")
		case <-time.After(1 * time.Second):
			t.Fatal("Done channel was not closed within the expected time")
		}
	})

	t.Run("Done channel works with select timeout", func(t *testing.T) {
		r, err := New().Quietly().Start("sleep", "5")
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

func TestRunning_Stop_Unix(t *testing.T) {
	t.Run("Stop sends SIGTERM for graceful exit", func(t *testing.T) {
		cmdStr := `
			trap 'echo "graceful shutdown" >&2; exit 0' SIGTERM
			echo "Process started..."
			for i in {1..20}; do # Loop for 10 seconds total
				sleep 0.5
			done
		`
		r, err := New().Quietly().Start("bash", "-c", cmdStr)
		assert.NoError(t, err)
		run, _ := r.(*Running)

		time.Sleep(100 * time.Millisecond)
		assert.True(t, run.Running())

		err = run.Stop(1 * time.Second)
		assert.NoError(t, err, "Stop should not return an error on graceful exit")

		res := run.Wait()
		assert.True(t, res.Successful(), "Process should have exited successfully (exit 0)")
		assert.Contains(t, res.ErrorOutput(), "graceful shutdown")
	})

	t.Run("Stop escalates to SIGKILL if timeout is exceeded", func(t *testing.T) {
		// This script traps SIGTERM, but then sleeps for 2 seconds. This is
		// longer than our Stop() timeout, which will force an escalation to SIGKILL.
		cmdStr := `
			trap 'echo "ignoring SIGTERM"; sleep 2' SIGTERM
			echo "Process started..."
			# Loop for 10 seconds total. The SIGKILL will interrupt this loop.
			for i in {1..20}; do
				sleep 0.5
			done
		`
		r, err := New().Quietly().Start("bash", "-c", cmdStr)
		assert.NoError(t, err)
		run, _ := r.(*Running)
		time.Sleep(100 * time.Millisecond)

		// Stop gracefully, but with a very short timeout (50ms). The process's
		// trap will take too long (2s), so it should be forcibly killed.
		err = run.Stop(50 * time.Millisecond)
		assert.NoError(t, err)

		res := run.Wait()
		assert.False(t, res.Successful(), "Process should have been killed, resulting in failure")
		assert.Equal(t, -1, res.ExitCode())
	})
}

func TestRunning_Signal_Unix(t *testing.T) {
	t.Run("can send a trappable signal to a process", func(t *testing.T) {
		tempDir := t.TempDir()
		signalFile := filepath.Join(tempDir, "signal_received.txt")

		cmdStr := fmt.Sprintf(`
			trap 'echo "USR1 received" > %s' USR1
			echo "Process started, waiting for signal..."
			# Loop until the signal file is created by the trap
			while [ ! -f "%s" ]; do
				sleep 0.1
			done
		`, signalFile, signalFile)

		r, err := New().Quietly().Start("bash", "-c", cmdStr)
		assert.NoError(t, err)
		run, _ := r.(*Running)

		time.Sleep(200 * time.Millisecond)
		assert.True(t, run.Running())

		err = run.Signal(unix.SIGUSR1)
		assert.NoError(t, err)

		res := run.Wait()
		assert.False(t, res.Successful(), "Process should have exited cleanly after the signal")

		content, err := os.ReadFile(signalFile)
		assert.NoError(t, err)
		assert.Equal(t, "USR1 received\n", string(content))
	})
}

func TestRunning_Kill_Unix(t *testing.T) {
	r, err := New().Quietly().Start("sleep", "5")
	assert.NoError(t, err)
	run, _ := r.(*Running)
	time.Sleep(50 * time.Millisecond)
	assert.NoError(t, run.Kill())
	res := run.Wait()
	assert.False(t, res.Successful())
}

func TestRunning_OutputAndError_Unix(t *testing.T) {
	r, err := New().Quietly().Start("sh", "-c", "printf out; printf err 1>&2")
	assert.NoError(t, err)
	run, _ := r.(*Running)
	res := run.Wait()
	assert.True(t, res.Successful())
	assert.Equal(t, "out", run.Output())
	assert.Equal(t, "err", run.ErrorOutput())
}

func TestRunning_Stop_AlreadyExited_Unix(t *testing.T) {
	r, err := New().Quietly().Start("sh", "-c", "true")
	assert.NoError(t, err)
	run, _ := r.(*Running)
	res := run.Wait()
	assert.True(t, res.Successful())
	// Process already exited; Stop should be a no-op
	assert.NoError(t, run.Stop(50*time.Millisecond))
}

func TestRunning_LatestOutputAndError_Unix(t *testing.T) {
	// Produce large outputs to exercise latest tail behavior (4096)
	script := `
	for i in $(seq 1 5000); do echo -n a; done
	for i in $(seq 1 5000); do echo -n b 1>&2; done
	`
	r, err := New().Quietly().Start("bash", "-c", script)
	assert.NoError(t, err)
	run, _ := r.(*Running)
	_ = run.Wait()
	assert.Equal(t, 4096, len(run.LatestOutput()))
	assert.Equal(t, 4096, len(run.LatestErrorOutput()))
}

func TestRunning_DisableBuffering_OutputEmpty_Unix(t *testing.T) {
	// When buffering is disabled, Output/ErrorOutput should be empty
	r, err := New().DisableBuffering().Quietly().Start("sh", "-c", "printf out; printf err 1>&2")
	assert.NoError(t, err)
	run, _ := r.(*Running)
	_ = run.Wait()
	assert.Equal(t, "", run.Output())
	assert.Equal(t, "", run.ErrorOutput())
}

func TestRunning_Panic_AppendsToStderr_Unix(t *testing.T) {
	stderr := &bytes.Buffer{}
	// Intentionally pass nil cmd to trigger panic in goroutine
	r := NewRunning(nil, nil, nil, stderr)
	<-r.Done()
	assert.Equal(t, "panic: runtime error: invalid memory address or nil pointer dereference\n", stderr.String())
}