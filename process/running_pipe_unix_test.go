//go:build !windows

package process

import (
	"bytes"
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

func TestRunningPipe_PIDs_Running_Done_Wait_Unix(t *testing.T) {
	rp, err := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("sh", "-c", "printf 'start'; sleep 0.2; printf 'end'").As("first")
		b.Command("cat").As("second")
	}).Start()
	assert.NoError(t, err)

	pids := rp.PIDs()
	assert.Len(t, pids, 2)
	assert.NotEqual(t, 0, pids["first"])
	assert.NotEqual(t, 0, pids["second"])

	assert.True(t, rp.Running())

	res := rp.Wait()
	assert.True(t, res.Successful())
	assert.False(t, rp.Running())
	assert.Equal(t, "startend", res.Output())
}

func TestRunningPipe_Signal_Unix(t *testing.T) {
	// Trap SIGTERM in the first stage but allow it to exit cleanly; ensure pipeline completes
	script := `trap 'echo term >&2; exit 0' TERM; echo begin; sleep 1`
	rp, err := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("bash", "-c", script).As("first")
		b.Command("cat").As("second")
	}).Start()
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, rp.Signal(unix.SIGTERM))

	res := rp.Wait()
	assert.True(t, res.Failed())
}

func TestRunningPipe_Stop_GracefulThenKill_Unix(t *testing.T) {
	scriptGrace := `trap 'echo bye >&2; exit 0' TERM; echo run; while true; do sleep 0.1; done`
	rp1, err := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("bash", "-c", scriptGrace).As("first")
	}).Start()
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, rp1.Stop(500*time.Millisecond))
	res1 := rp1.Wait()
	assert.True(t, res1.Successful())

	// Now, force kill after timeout (trap sleeps too long)
	scriptBlock := `trap 'sleep 2' TERM; echo run; while true; do sleep 0.1; done`
	rp2, err := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("bash", "-c", scriptBlock).As("first")
	}).Start()
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, rp2.Stop(50*time.Millisecond))
	res2 := rp2.Wait()
	assert.False(t, res2.Successful())
}

func TestRunningPipe_Panic_AppendsToStderr_Unix(t *testing.T) {
	stderr := &bytes.Buffer{}
	// Create a RunningPipe with a nil command to force panic in Wait
	rp := NewRunningPipe(context.Background(), []*exec.Cmd{nil}, []*PipeCommand{{key: "0"}}, nil, nil, nil, []*bytes.Buffer{nil}, []*bytes.Buffer{stderr}, false, "")
	<-rp.Done()
	assert.Equal(t, "panic: runtime error: invalid memory address or nil pointer dereference\n", stderr.String())
}

func TestRunningPipe_Interruption_MiddleCommandFails(t *testing.T) {
	rp, err := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("echo", "line1").As("first")
		b.Command("sh", "-c", "cat; exit 1").As("second")
		b.Command("echo", "line2")
		b.Command("cat")
	}).Start()
	assert.NoError(t, err)

	res := rp.Wait()
	assert.False(t, res.Successful())
	assert.Equal(t, 1, res.ExitCode())
	assert.Contains(t, res.Output(), "line1")
	assert.NotContains(t, res.Output(), "line2")
}

func TestRunningPipe_spinnerForCommand(t *testing.T) {
	t.Run("NoLoading", func(t *testing.T) {
		// Test when loading is disabled (both global and command-specific)
		ctx := context.Background()
		pc := NewPipeCommand("test", "echo", []string{"hello"})

		rp := &RunningPipe{
			ctx:            ctx,
			pipeCommands:   []*PipeCommand{pc},
			loading:        false,
			loadingMessage: "",
		}

		executed := false
		err := rp.spinnerForCommand(0, func() error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("WithCommandSpecificMessage", func(t *testing.T) {
		// Test when command has its own loading message
		ctx := context.Background()
		pc := NewPipeCommand("test", "echo", []string{"hello"})
		pc.WithSpinner("Custom loading message")

		rp := &RunningPipe{
			ctx:            ctx,
			pipeCommands:   []*PipeCommand{pc},
			loading:        false,
			loadingMessage: "",
		}

		executed := false
		err := rp.spinnerForCommand(0, func() error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("WithGeneratedMessage", func(t *testing.T) {
		// Test when no custom message is set - should generate from command
		ctx := context.Background()
		pc := NewPipeCommand("test", "echo", []string{"hello", "world"})
		pc.loading = true

		rp := &RunningPipe{
			ctx:            ctx,
			pipeCommands:   []*PipeCommand{pc},
			loading:        false,
			loadingMessage: "",
		}

		executed := false
		err := rp.spinnerForCommand(0, func() error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("WithError", func(t *testing.T) {
		// Test when the function returns an error
		ctx := context.Background()
		pc := NewPipeCommand("test", "false", []string{})
		pc.loading = true

		rp := &RunningPipe{
			ctx:            ctx,
			pipeCommands:   []*PipeCommand{pc},
			loading:        false,
			loadingMessage: "",
		}

		testErr := assert.AnError
		err := rp.spinnerForCommand(0, func() error {
			return testErr
		})

		assert.Error(t, err)
		assert.Equal(t, testErr, err)
	})

	t.Run("GlobalLoadingEnabled", func(t *testing.T) {
		// Test when global loading is enabled for all commands
		ctx := context.Background()
		pc := NewPipeCommand("test", "echo", []string{"hello"})

		rp := &RunningPipe{
			ctx:            ctx,
			pipeCommands:   []*PipeCommand{pc},
			loading:        true, // Global loading enabled
			loadingMessage: "",
		}

		executed := false
		err := rp.spinnerForCommand(0, func() error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
	})
}
