//go:build !windows

package process

import (
	"bytes"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

func TestRunningPipe_PIDs_Running_Done_Wait_Unix(t *testing.T) {
	rp, err := NewPipe().Quietly().Start(func(b contractsprocess.Pipe) {
		b.Command("sh", "-c", "printf 'start'; sleep 0.2; printf 'end'").As("first")
		b.Command("cat").As("second")
	})
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
	rp, err := NewPipe().Quietly().Start(func(b contractsprocess.Pipe) {
		b.Command("bash", "-c", script).As("first")
		b.Command("cat").As("second")
	})
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, rp.Signal(unix.SIGTERM))

	res := rp.Wait()
	assert.True(t, res.Failed())
}

func TestRunningPipe_Stop_GracefulThenKill_Unix(t *testing.T) {
	scriptGrace := `trap 'echo bye >&2; exit 0' TERM; echo run; while true; do sleep 0.1; done`
	rp1, err := NewPipe().Quietly().Start(func(b contractsprocess.Pipe) {
		b.Command("bash", "-c", scriptGrace).As("first")
	})
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, rp1.Stop(500*time.Millisecond))
	res1 := rp1.Wait()
	assert.True(t, res1.Successful())

	// Now, force kill after timeout (trap sleeps too long)
	scriptBlock := `trap 'sleep 2' TERM; echo run; while true; do sleep 0.1; done`
	rp2, err := NewPipe().Quietly().Start(func(b contractsprocess.Pipe) {
		b.Command("bash", "-c", scriptBlock).As("first")
	})
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, rp2.Stop(50*time.Millisecond))
	res2 := rp2.Wait()
	assert.False(t, res2.Successful())
}

func TestRunningPipe_Panic_AppendsToStderr_Unix(t *testing.T) {
	stderr := &bytes.Buffer{}
	// Create a RunningPipe with a nil command to force panic in Wait
	rp := NewRunningPipe([]*exec.Cmd{nil}, []*Step{{key: "0"}}, nil, nil, nil, []*bytes.Buffer{nil}, []*bytes.Buffer{stderr})
	<-rp.Done()
	assert.Equal(t, "panic: runtime error: invalid memory address or nil pointer dereference\n", stderr.String())
}
