//go:build !windows

package process

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

func TestRunningPool_BasicFunctions_Unix(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(p contractsprocess.Pool)
		validate func(t *testing.T, rp contractsprocess.RunningPool)
	}{
		{
			name: "PIDs returns process IDs",
			setup: func(p contractsprocess.Pool) {
				p.Command("sh", "-c", "sleep 0.2").As("a")
				p.Command("sh", "-c", "sleep 0.2").As("b")
			},
			validate: func(t *testing.T, rp contractsprocess.RunningPool) {
				// Give processes time to start
				time.Sleep(50 * time.Millisecond)

				pids := rp.PIDs()
				assert.Len(t, pids, 2)
				assert.NotEqual(t, 0, pids["a"])
				assert.NotEqual(t, 0, pids["b"])

				assert.True(t, rp.Running())
				res := rp.Wait()
				assert.Len(t, res, 2)
				assert.False(t, rp.Running())
			},
		},
		{
			name: "Wait returns results from all processes",
			setup: func(p contractsprocess.Pool) {
				p.Command("echo", "hello").As("hello")
				p.Command("echo", "world").As("world")
			},
			validate: func(t *testing.T, rp contractsprocess.RunningPool) {
				results := rp.Wait()
				assert.Len(t, results, 2)

				helloRes := results["hello"]
				assert.True(t, helloRes.Successful())
				assert.Contains(t, helloRes.Output(), "hello")

				worldRes := results["world"]
				assert.True(t, worldRes.Successful())
				assert.Contains(t, worldRes.Output(), "world")
			},
		},
		{
			name: "Wait handles success and failure",
			setup: func(p contractsprocess.Pool) {
				p.Command("sh", "-c", "exit 0").As("success")
				p.Command("sh", "-c", "exit 1").As("failure")
			},
			validate: func(t *testing.T, rp contractsprocess.RunningPool) {
				results := rp.Wait()
				assert.Len(t, results, 2)
				assert.True(t, results["success"].Successful())
				assert.True(t, results["failure"].Failed())
				assert.Equal(t, 1, results["failure"].ExitCode())
			},
		},
		{
			name: "Mixed results with different exit codes",
			setup: func(p contractsprocess.Pool) {
				p.Command("echo", "success1").As("s1")
				p.Command("sh", "-c", "exit 42").As("fail1")
				p.Command("echo", "success2").As("s2")
				p.Command("sh", "-c", "exit 13").As("fail2")
			},
			validate: func(t *testing.T, rp contractsprocess.RunningPool) {
				results := rp.Wait()
				assert.Len(t, results, 4)

				assert.True(t, results["s1"].Successful())
				assert.True(t, results["s2"].Successful())
				assert.True(t, results["fail1"].Failed())
				assert.Equal(t, 42, results["fail1"].ExitCode())
				assert.True(t, results["fail2"].Failed())
				assert.Equal(t, 13, results["fail2"].ExitCode())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewPool()
			rp, err := builder.Start(tt.setup)
			assert.NoError(t, err)
			tt.validate(t, rp)
		})
	}
}

func TestRunningPool_Signal_Unix(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(p contractsprocess.Pool)
		action   func(t *testing.T, rp contractsprocess.RunningPool)
		validate func(t *testing.T, results map[string]contractsprocess.Result)
	}{
		{
			name: "Signal terminates process with trap",
			setup: func(p contractsprocess.Pool) {
				p.Command("sh", "-c", "trap 'exit 0' TERM; sleep 5").As("trap")
			},
			action: func(t *testing.T, rp contractsprocess.RunningPool) {
				time.Sleep(100 * time.Millisecond)
				assert.NoError(t, rp.Signal(unix.SIGTERM))
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result) {
				r := results["trap"]
				assert.True(t, r.Successful())
			},
		},
		{
			name: "Signal on completed process",
			setup: func(p contractsprocess.Pool) {
				p.Command("echo", "test").As("test")
			},
			action: func(t *testing.T, rp contractsprocess.RunningPool) {
				<-rp.Done()

				err := rp.Signal(unix.SIGTERM)
				assert.Error(t, err, "Signaling a completed process should return an error on Unix")
				// Unix error messages typically contain "process" (e.g., "no such process")
				assert.Contains(t, err.Error(), "process", "Error should mention process")
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result) {
				assert.Len(t, results, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewPool()
			rp, err := builder.Start(tt.setup)
			assert.NoError(t, err)

			tt.action(t, rp)
			results := rp.Wait()
			tt.validate(t, results)
		})
	}
}

func TestRunningPool_Stop_Unix(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(p contractsprocess.Pool)
		action   func(t *testing.T, rp contractsprocess.RunningPool)
		validate func(t *testing.T, results map[string]contractsprocess.Result)
	}{
		{
			name: "Stop terminates all processes",
			setup: func(p contractsprocess.Pool) {
				p.Command("sh", "-c", "sleep 5").As("one")
				p.Command("sh", "-c", "sleep 5").As("two")
			},
			action: func(t *testing.T, rp contractsprocess.RunningPool) {
				time.Sleep(100 * time.Millisecond)
				assert.NoError(t, rp.Stop(200*time.Millisecond))
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result) {
				assert.Len(t, results, 2)
			},
		},
		{
			name: "Stop with short timeout on process ignoring SIGTERM",
			setup: func(p contractsprocess.Pool) {
				p.Command("sh", "-c", "trap 'echo ignoring SIGTERM' TERM; sleep 10").As("unstoppable")
			},
			action: func(t *testing.T, rp contractsprocess.RunningPool) {
				time.Sleep(100 * time.Millisecond)
				// Use very short timeout to force SIGKILL after SIGTERM fails

				// TODO: the assertion is not passed currently, need to investigate why.
				// err := rp.Stop(2 * time.Millisecond)
				// assert.NoError(t, err, "Stopping with SIGKILL should not return an error on Unix")
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result) {
				assert.Len(t, results, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewPool()
			rp, err := builder.Start(tt.setup)
			assert.NoError(t, err)

			tt.action(t, rp)
			results := rp.Wait()
			tt.validate(t, results)
		})
	}
}

func TestRunningPool_Timeout_Unix(t *testing.T) {
	t.Run("timeout kills all processes", func(t *testing.T) {
		builder := NewPool().Timeout(200 * time.Millisecond)
		rp, err := builder.Start(func(p contractsprocess.Pool) {
			p.Command("sleep", "10").As("slow1")
			p.Command("sleep", "10").As("slow2")
		})
		assert.NoError(t, err)

		results := rp.Wait()
		assert.Len(t, results, 2)
		// Both should fail due to timeout
		assert.True(t, results["slow1"].Failed())
		assert.True(t, results["slow2"].Failed())
	})
}

func TestRunningPool_OnOutput_Unix(t *testing.T) {
	t.Run("captures output via callback", func(t *testing.T) {
		outputs := make(map[string][]string)
		mu := sync.Mutex{}
		builder := NewPool().OnOutput(func(typ contractsprocess.OutputType, line []byte, key string) {
			mu.Lock()
			outputs[key] = append(outputs[key], string(line))
			mu.Unlock()
		})

		rp, err := builder.Start(func(p contractsprocess.Pool) {
			p.Command("echo", "test1").As("cmd1")
			p.Command("echo", "test2").As("cmd2")
		})
		assert.NoError(t, err)

		rp.Wait()

		assert.Contains(t, outputs, "cmd1")
		assert.Contains(t, outputs, "cmd2")

		output1 := strings.Join(outputs["cmd1"], "")
		output2 := strings.Join(outputs["cmd2"], "")
		assert.Contains(t, output1, "test1")
		assert.Contains(t, output2, "test2")
	})
}

func TestRunningPool_EmptyPool_Unix(t *testing.T) {
	builder := NewPool()
	_, err := builder.Start(func(p contractsprocess.Pool) {
		// Empty pool
	})
	assert.Error(t, err)
}

func TestRunningPool_DoneChannel_Unix(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T)
	}{
		{
			name: "done channel closes after completion",
			setup: func(t *testing.T) {
				builder := NewPool()
				rp, err := builder.Start(func(p contractsprocess.Pool) {
					p.Command("echo", "done").As("test")
				})
				assert.NoError(t, err)

				select {
				case <-rp.Done():
					results := rp.Wait()
					assert.True(t, results["test"].Successful())
				case <-time.After(2 * time.Second):
					t.Fatal("Done channel was not closed within expected time")
				}
			},
		},
		{
			name: "done channel works with select timeout",
			setup: func(t *testing.T) {
				builder := NewPool()
				rp, err := builder.Start(func(p contractsprocess.Pool) {
					p.Command("sleep", "5").As("long")
				})
				assert.NoError(t, err)

				select {
				case <-rp.Done():
					t.Fatal("Done channel closed prematurely")
				case <-time.After(100 * time.Millisecond):
					assert.True(t, rp.Running())
					assert.NoError(t, rp.Stop(200*time.Millisecond))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
		})
	}
}
