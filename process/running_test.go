package process

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

func TestRunning_Basics_TableDriven(t *testing.T) {
	tests := []struct {
		name  string
		cmd   []string
		check func(t *testing.T, run *Running)
	}{
		{
			name: "PID and Command non-empty while running",
			cmd:  []string{"sh", "-c", "sleep 0.2"},
			check: func(t *testing.T, run *Running) {
				assert.NotEqual(t, 0, run.PID())
				assert.Contains(t, run.Command(), "sleep 0.2")
				assert.True(t, run.Running())
				res := run.Wait()
				assert.NotNil(t, res)
				assert.Equal(t, 0, res.ExitCode())
			},
		},
		{
			name: "LatestOutput and LatestErrorOutput limited",
			cmd:  []string{"sh", "-c", "yes x | head -c 5000 1>/dev/stdout; yes y | head -c 5000 1>/dev/stderr"},
			check: func(t *testing.T, run *Running) {
				res := run.Wait()
				assert.True(t, res.Successful())
				assert.GreaterOrEqual(t, len(run.Output()), 4096)
				assert.GreaterOrEqual(t, len(run.ErrorOutput()), 4096)
				assert.Equal(t, 4096, len(run.LatestOutput()))
				assert.Equal(t, 4096, len(run.LatestErrorOutput()))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := New().Command(test.cmd[0], test.cmd[1:]...).Start(context.Background())
			assert.NoError(t, err)
			run, ok := r.(*Running)
			assert.True(t, ok, "unexpected running type")
			if ok {
				test.check(t, run)
			}
		})
	}
}

func TestRunning_SignalAndStop(t *testing.T) {
	r, err := New().Command("sh", "-c", "sleep 10").Start(context.Background())
	assert.NoError(t, err)
	run := r.(*Running)
	assert.True(t, run.Running())
	// send SIGTERM and wait for graceful stop
	assert.NoError(t, run.Signal(unix.SIGTERM))
	assert.NoError(t, run.Stop(500*time.Millisecond))
	res := run.Wait()
	assert.False(t, res.Successful())
}
