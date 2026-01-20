//go:build windows

package process

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

func TestRunningPipe_PIDs_Running_Done_Wait_Windows(t *testing.T) {
	rp, err := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("cmd", "/C", "(echo start & powershell -NoLogo -NoProfile -Command Start-Sleep -Milliseconds 200 & echo end)").As("first")
		b.Command("cmd", "/C", "more").As("second")
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
	assert.Equal(t, "start \r\nend\r\n\r\n", res.Output())
}

func TestRunningPipe_Stop_Windows(t *testing.T) {
	rp, err := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Seconds 10").As("sleep")
	}).Start()
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, rp.Stop(1*time.Second))
	res := rp.Wait()
	assert.False(t, res.Successful())
}

func TestRunningPipe_Signal_Windows_NoOp(t *testing.T) {
	rp, err := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Seconds 1").As("sleep")
	}).Start()
	assert.NoError(t, err)
	assert.NoError(t, rp.Signal(os.Interrupt))
	_ = rp.Wait()
}

func TestRunningPipe_Panic_AppendsToStderr_Windows(t *testing.T) {
	stderr := &bytes.Buffer{}
	rp := NewRunningPipe(context.Background(), []*exec.Cmd{nil}, []*PipeCommand{{key: "0"}}, nil, nil, nil, []*bytes.Buffer{nil}, []*bytes.Buffer{stderr}, false, "")
	<-rp.Done()
	assert.Equal(t, "panic: runtime error: invalid memory address or nil pointer dereference\n", stderr.String())
}

func TestRunningPipe_Interruption_MiddleCommandFails(t *testing.T) {
	rp, err := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("cmd", "/C", "echo line1").As("first")
		b.Command("cmd", "/C", "more & exit 1").As("second")
		b.Command("cmd", "/C", "echo line2")
		b.Command("cmd", "/C", "more")
	}).Start()
	assert.NoError(t, err)

	res := rp.Wait()
	assert.False(t, res.Successful())
	assert.Equal(t, 1, res.ExitCode())
	assert.Contains(t, res.Output(), "line1")
	assert.NotContains(t, res.Output(), "line2")
}
