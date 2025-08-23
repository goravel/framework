package process

import (
	"context"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunning_Wait_And_Output(t *testing.T) {
	ctx := context.Background()
	cmd, args := echoBothCommand()

	r, err := Start(ctx, cmd, args...)
	assert.NoError(t, err)
	if r.PID() == 0 {
		t.Fatalf("expected PID > 0")
	}
	assert.True(t, r.Running())
	res := r.Wait()
	assert.Contains(t, res.Output(), "out")
	assert.Contains(t, res.ErrorOutput(), "err")
}

func TestRunning_Signal_Kill(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("signals behave differently on Windows")
	}

	ctx := context.Background()
	cmd, args := killableSleepCmd(5)

	r, err := Start(ctx, cmd, args...)
	assert.NoError(t, err)
	assert.True(t, r.Running())
	// Send SIGKILL and ensure process ends quickly
	err = r.Kill()
	assert.NoError(t, err)
	res := r.Wait()
	assert.True(t, res.Failed())
}

func killableSleepCmd(seconds int) (string, []string) {
	if runtime.GOOS == "windows" {
		return "powershell", []string{"-Command", "Start-Sleep -Seconds " + strconv.Itoa(seconds)}
	}
	return "sleep", []string{strconv.Itoa(seconds)}
}
