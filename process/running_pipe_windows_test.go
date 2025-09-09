//go:build windows

package process

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

func TestRunningPipe_PIDs_Running_Done_Wait_Windows(t *testing.T) {
	rp, err := NewPipe().Quietly().Start(func(b contractsprocess.Pipe) {
		b.Command("cmd", "/C", "(echo start & powershell -NoLogo -NoProfile -Command Start-Sleep -Milliseconds 200 & echo end)").As("first")
		b.Command("cmd", "/C", "more").As("second")
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
	assert.Equal(t, "start \r\nend\r\n\r\n", res.Output())
}

func TestRunningPipe_Stop_Windows(t *testing.T) {
	rp, err := NewPipe().Quietly().Start(func(b contractsprocess.Pipe) {
		b.Command("powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Seconds 10").As("sleep")
	})
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, rp.Stop(1*time.Second))
	res := rp.Wait()
	assert.False(t, res.Successful())
}

func TestRunningPipe_Signal_Windows_NoOp(t *testing.T) {
	rp, err := NewPipe().Quietly().Start(func(b contractsprocess.Pipe) {
		b.Command("powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Seconds 1").As("sleep")
	})
	assert.NoError(t, err)
	assert.NoError(t, rp.Signal(os.Interrupt))
	_ = rp.Wait()
}

func TestRunningPipe_Panic_AppendsToStderr_Windows(t *testing.T) {
	stderr := &bytes.Buffer{}
	rp := NewRunningPipe([]*exec.Cmd{nil}, []*Step{{key: "0"}}, nil, nil, nil, []*bytes.Buffer{nil}, []*bytes.Buffer{stderr})
	<-rp.Done()
	assert.Equal(t, "panic: runtime error: invalid memory address or nil pointer dereference\n", stderr.String())
}
