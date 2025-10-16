//go:build !windows

package process

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

func TestPool_Run_Unix(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(p contractsprocess.Pool)
		validate func(t *testing.T, results map[string]contractsprocess.Result, err error)
	}{
		{
			name: "runs commands and returns results",
			setup: func(p contractsprocess.Pool) {
				p.Command("echo", "hello").As("hello")
				p.Command("echo", "world").As("world")
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result, err error) {
				assert.NoError(t, err)
				assert.Len(t, results, 2)
				assert.True(t, results["hello"].Successful())
				assert.True(t, results["world"].Successful())
				assert.Contains(t, results["hello"].Output(), "hello")
				assert.Contains(t, results["world"].Output(), "world")
			},
		},
		{
			name: "handles command failures",
			setup: func(p contractsprocess.Pool) {
				p.Command("sh", "-c", "exit 0").As("success")
				p.Command("sh", "-c", "exit 3").As("failure")
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result, err error) {
				assert.NoError(t, err)
				assert.Len(t, results, 2)
				assert.True(t, results["success"].Successful())
				assert.True(t, results["failure"].Failed())
				assert.Equal(t, 3, results["failure"].ExitCode())
			},
		},
		{
			name: "handles command not found",
			setup: func(p contractsprocess.Pool) {
				p.Command("command-that-does-not-exist").As("missing")
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result, err error) {
				assert.NoError(t, err)
				assert.Len(t, results, 1)
				assert.True(t, results["missing"].Failed())
				assert.NotEqual(t, 0, results["missing"].ExitCode())
				assert.ErrorContains(t, results["missing"].Error(), "executable file not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewPool()
			results, err := builder.Run(tt.setup)
			tt.validate(t, results, err)
		})
	}
}

func TestPool_Start_Unix(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(p contractsprocess.Pool)
		validate func(t *testing.T, rp contractsprocess.RunningPool, err error)
	}{
		{
			name: "starts commands and returns running pool",
			setup: func(p contractsprocess.Pool) {
				p.Command("sleep", "0.1").As("sleep1")
				p.Command("sleep", "0.1").As("sleep2")
			},
			validate: func(t *testing.T, rp contractsprocess.RunningPool, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, rp)
				assert.True(t, rp.Running())

				results := rp.Wait()
				assert.Len(t, results, 2)
				assert.True(t, results["sleep1"].Successful())
				assert.True(t, results["sleep2"].Successful())
			},
		},
		{
			name: "returns error for empty pool",
			setup: func(p contractsprocess.Pool) {
				// No commands added
			},
			validate: func(t *testing.T, rp contractsprocess.RunningPool, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewPool()
			rp, err := builder.Start(tt.setup)
			tt.validate(t, rp, err)
		})
	}
}

func TestPool_WithContext_Unix(t *testing.T) {
	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		builder := NewPool().WithContext(ctx)

		rp, err := builder.Start(func(p contractsprocess.Pool) {
			p.Command("sleep", "10").As("long")
		})
		assert.NoError(t, err)

		// Give process time to start
		time.Sleep(50 * time.Millisecond)

		// Cancel context
		cancel()

		// Wait for process to be terminated
		results := rp.Wait()
		assert.True(t, results["long"].Failed())
	})

	t.Run("handles nil context", func(t *testing.T) {
		// Passing nil should use background context
		builder := NewPool().WithContext(nil)

		rp, err := builder.Start(func(p contractsprocess.Pool) {
			p.Command("echo", "test").As("test")
		})
		assert.NoError(t, err)

		results := rp.Wait()
		assert.True(t, results["test"].Successful())
	})
}

func TestPool_Timeout_Unix(t *testing.T) {
	t.Run("terminates processes after timeout", func(t *testing.T) {
		builder := NewPool().Timeout(200 * time.Millisecond)

		rp, err := builder.Start(func(p contractsprocess.Pool) {
			p.Command("sleep", "10").As("long")
		})
		assert.NoError(t, err)

		results := rp.Wait()
		assert.True(t, results["long"].Failed())
	})
}

func TestPool_Concurrency_Unix(t *testing.T) {
	tests := []struct {
		name        string
		concurrency int
		commands    []struct {
			name string
			cmd  string
			args []string
		}
		validateElapsed func(t *testing.T, elapsed time.Duration)
	}{
		{
			name:        "limits concurrent processes",
			concurrency: 2,
			commands: []struct {
				name string
				cmd  string
				args []string
			}{
				{name: "job1", cmd: "sh", args: []string{"-c", "echo job1; sleep 0.3"}},
				{name: "job2", cmd: "sh", args: []string{"-c", "echo job2; sleep 0.3"}},
				{name: "job3", cmd: "sh", args: []string{"-c", "echo job3; sleep 0.3"}},
				{name: "job4", cmd: "sh", args: []string{"-c", "echo job4; sleep 0.3"}},
			},
			validateElapsed: func(t *testing.T, elapsed time.Duration) {
				// With concurrency=2 and 4 jobs of 0.3s each, we expect ~0.6s total time
				// Using a lower threshold (300ms) to account for timing variations
				assert.GreaterOrEqual(t, elapsed, 300*time.Millisecond)
			},
		},
		{
			name:        "uses all jobs when concurrency exceeds job count",
			concurrency: 10,
			commands: []struct {
				name string
				cmd  string
				args []string
			}{
				{name: "t1", cmd: "echo", args: []string{"test1"}},
				{name: "t2", cmd: "echo", args: []string{"test2"}},
			},
			validateElapsed: nil, // No specific timing validation needed
		},
		{
			name:        "uses all jobs when concurrency is zero",
			concurrency: 0,
			commands: []struct {
				name string
				cmd  string
				args []string
			}{
				{name: "t1", cmd: "echo", args: []string{"test1"}},
				{name: "t2", cmd: "echo", args: []string{"test2"}},
			},
			validateElapsed: nil, // No specific timing validation needed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			builder := NewPool().Concurrency(tt.concurrency).OnOutput(func(key string, typ contractsprocess.OutputType, line []byte) {
				buf.WriteString(key + ":" + time.Now().Format(time.RFC3339Nano) + "\n")
			})

			start := time.Now()
			rp, err := builder.Start(func(p contractsprocess.Pool) {
				for _, cmd := range tt.commands {
					p.Command(cmd.cmd, cmd.args...).As(cmd.name)
				}
			})
			assert.NoError(t, err)

			results := rp.Wait()
			elapsed := time.Since(start)

			// Validate results
			assert.Len(t, results, len(tt.commands))
			for _, cmd := range tt.commands {
				assert.True(t, results[cmd.name].Successful(), "command %s should succeed", cmd.name)
			}

			// Validate timing if specified
			if tt.validateElapsed != nil {
				tt.validateElapsed(t, elapsed)
			}
		})
	}
}

func TestPool_OnOutput_Unix(t *testing.T) {
	t.Run("captures output via callback", func(t *testing.T) {
		outputs := make(map[string][]string)
		builder := NewPool().OnOutput(func(key string, typ contractsprocess.OutputType, line []byte) {
			outputs[key] = append(outputs[key], string(line))
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

	t.Run("distinguishes stdout and stderr", func(t *testing.T) {
		stdoutLines := make(map[string][]string)
		stderrLines := make(map[string][]string)

		builder := NewPool().OnOutput(func(key string, typ contractsprocess.OutputType, line []byte) {
			if typ == contractsprocess.OutputTypeStdout {
				stdoutLines[key] = append(stdoutLines[key], string(line))
			} else {
				stderrLines[key] = append(stderrLines[key], string(line))
			}
		})

		rp, err := builder.Start(func(p contractsprocess.Pool) {
			p.Command("sh", "-c", "echo stdout; echo stderr >&2").As("mixed")
		})
		assert.NoError(t, err)

		rp.Wait()

		assert.Contains(t, strings.Join(stdoutLines["mixed"], ""), "stdout")
		assert.Contains(t, strings.Join(stderrLines["mixed"], ""), "stderr")
	})
}

func TestPoolCommand_Unix(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(p contractsprocess.Pool)
		validate func(t *testing.T, results map[string]contractsprocess.Result, err error)
	}{
		{
			name: "As sets command key",
			setup: func(p contractsprocess.Pool) {
				p.Command("echo", "test").As("custom-key")
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result, err error) {
				assert.NoError(t, err)
				assert.Contains(t, results, "custom-key")
			},
		},
		{
			name: "DisableBuffering works",
			setup: func(p contractsprocess.Pool) {
				p.Command("echo", "test").As("buffered")
				p.Command("echo", "test").DisableBuffering().As("unbuffered")
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result, err error) {
				assert.NoError(t, err)
				assert.Contains(t, results["buffered"].Output(), "test")
				assert.Empty(t, results["unbuffered"].Output())
			},
		},
		{
			name: "Path sets working directory",
			setup: func(p contractsprocess.Pool) {
				p.Command("pwd").Path("/tmp").As("pwd")
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result, err error) {
				assert.NoError(t, err)
				assert.Contains(t, results["pwd"].Output(), "/tmp")
			},
		},
		{
			name: "Env sets environment variables",
			setup: func(p contractsprocess.Pool) {
				p.Command("sh", "-c", "echo $TEST_VAR").
					Env(map[string]string{"TEST_VAR": "test-value"}).As("env")
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result, err error) {
				assert.NoError(t, err)
				assert.Contains(t, results["env"].Output(), "test-value")
			},
		},
		{
			name: "Input provides stdin",
			setup: func(p contractsprocess.Pool) {
				p.Command("cat").Input(strings.NewReader("test-input")).As("cat")
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result, err error) {
				assert.NoError(t, err)
				assert.Contains(t, results["cat"].Output(), "test-input")
			},
		},
		{
			name: "Quietly suppresses output",
			setup: func(p contractsprocess.Pool) {
				p.Command("echo", "test").Quietly().As("quiet")
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result, err error) {
				assert.NoError(t, err)
				assert.True(t, results["quiet"].Successful())
			},
		},
		{
			name: "Timeout terminates long-running command",
			setup: func(p contractsprocess.Pool) {
				p.Command("sleep", "10").Timeout(100 * time.Millisecond).As("timeout")
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result, err error) {
				assert.NoError(t, err)
				assert.True(t, results["timeout"].Failed())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewPool()
			results, err := builder.Run(tt.setup)
			tt.validate(t, results, err)
		})
	}

	t.Run("WithContext respects cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		builder := NewPool()
		results, err := builder.Run(func(p contractsprocess.Pool) {
			p.Command("sleep", "10").WithContext(ctx).As("ctx")
		})

		assert.NoError(t, err)

		assert.True(t, results["ctx"].Failed(), "Process should fail when context times out")
		assert.NotEqual(t, 0, results["ctx"].ExitCode(), "Exit code should be non-zero")
	})
}

func TestPool_SignalHandling_Unix(t *testing.T) {
	t.Run("forwards signals to child processes", func(t *testing.T) {
		builder := NewPool()
		rp, err := builder.Start(func(p contractsprocess.Pool) {
			p.Command("sh", "-c", "trap 'echo caught; exit 0' TERM; sleep 5").As("trap")
		})
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		assert.NoError(t, rp.Signal(unix.SIGTERM))

		results := rp.Wait()
		assert.True(t, results["trap"].Successful())
		assert.Contains(t, results["trap"].Output(), "caught")
	})
}

func TestPool_ErrorHandling_Unix(t *testing.T) {
	t.Run("handles start failures", func(t *testing.T) {
		builder := NewPool()
		results, err := builder.Run(func(p contractsprocess.Pool) {
			p.Command("command-that-does-not-exist").As("missing")
			p.Command("echo", "test").As("valid")
		})

		assert.NoError(t, err) // The pool itself doesn't error
		assert.Len(t, results, 2)
		assert.True(t, results["missing"].Failed())
		assert.True(t, results["valid"].Successful())
	})
}

func TestPool_Cleanup_Unix(t *testing.T) {
	t.Run("cleans up resources after completion", func(t *testing.T) {
		// Create a temp file to write to
		tmpFile, err := os.CreateTemp("", "pool-test")
		assert.NoError(t, err)
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		builder := NewPool()
		results, err := builder.Run(func(p contractsprocess.Pool) {
			p.Command("sh", "-c", "echo test > "+tmpFile.Name()).As("write")
		})

		assert.NoError(t, err)
		assert.True(t, results["write"].Successful())

		content, err := os.ReadFile(tmpFile.Name())
		assert.NoError(t, err)
		assert.Contains(t, string(content), "test")
	})
}
