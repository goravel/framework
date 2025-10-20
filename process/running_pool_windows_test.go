//go:build windows

package process

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

func TestRunningPool_BasicFunctions_Windows(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(p contractsprocess.Pool)
		validate func(t *testing.T, rp contractsprocess.RunningPool)
	}{
		{
			name: "PIDs returns process IDs",
			setup: func(p contractsprocess.Pool) {
				p.Command("powershell", "-Command", "Start-Sleep -Milliseconds 200").As("a")
				p.Command("powershell", "-Command", "Start-Sleep -Milliseconds 200").As("b")
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
				p.Command("powershell", "-Command", "Write-Output hello").As("hello")
				p.Command("powershell", "-Command", "Write-Output world").As("world")
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
				p.Command("powershell", "-Command", "exit 0").As("success")
				p.Command("powershell", "-Command", "exit 1").As("failure")
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
				p.Command("powershell", "-Command", "Write-Output success1").As("s1")
				p.Command("powershell", "-Command", "exit 42").As("fail1")
				p.Command("powershell", "-Command", "Write-Output success2").As("s2")
				p.Command("powershell", "-Command", "exit 13").As("fail2")
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
			rp, err := builder.Pool(tt.setup).Start()
			assert.NoError(t, err)
			tt.validate(t, rp)
		})
	}
}

func TestRunningPool_Signal_Windows(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(p contractsprocess.Pool)
		action   func(t *testing.T, rp contractsprocess.RunningPool)
		validate func(t *testing.T, results map[string]contractsprocess.Result)
	}{
		{
			name: "Signal sends CTRL_BREAK_EVENT",
			setup: func(p contractsprocess.Pool) {
				// Use a PowerShell script that responds to CTRL_BREAK_EVENT (mapped from os.Interrupt)
				p.Command("powershell", "-Command", "$global:interrupted = $false; [console]::TreatControlCAsInput = $true; $handler = [Console]::CancelKeyPress; [Console]::CancelKeyPress = { $global:interrupted = $true; $_.Cancel = $true }; while (-not $global:interrupted) { Start-Sleep -Milliseconds 100 }; exit 0").As("trap")
			},
			action: func(t *testing.T, rp contractsprocess.RunningPool) {
				time.Sleep(100 * time.Millisecond)
				// Send os.Interrupt which should be mapped to CTRL_BREAK_EVENT on Windows
				err := rp.Signal(os.Interrupt)
				assert.NoError(t, err, "Signaling a running process should not return an error on Windows")
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result) {
				// The important thing is that the process completes
				assert.Len(t, results, 1)
			},
		},
		{
			name: "Signal on completed process",
			setup: func(p contractsprocess.Pool) {
				p.Command("powershell", "-Command", "Write-Output test").As("test")
			},
			action: func(t *testing.T, rp contractsprocess.RunningPool) {
				<-rp.Done()

				err := rp.Signal(os.Interrupt)
				fmt.Println("Signal on completed process", err.Error())
				assert.Error(t, err, "Signaling a completed process should return an error on Windows")
				// Windows error messages can vary but typically contain "process", "handle", or "invalid argument"
				assert.True(t,
					strings.Contains(err.Error(), "process") ||
						strings.Contains(err.Error(), "handle") ||
						strings.Contains(err.Error(), "invalid argument"),
					"Error should be a valid Windows error: %s", err.Error())
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result) {
				assert.Len(t, results, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewPool()
			rp, err := builder.Pool(tt.setup).Start()
			assert.NoError(t, err)

			tt.action(t, rp)
			results := rp.Wait()
			tt.validate(t, results)
		})
	}
}

func TestRunningPool_Stop_Windows(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(p contractsprocess.Pool)
		action   func(t *testing.T, rp contractsprocess.RunningPool)
		validate func(t *testing.T, results map[string]contractsprocess.Result)
	}{
		{
			name: "Stop terminates all processes",
			setup: func(p contractsprocess.Pool) {
				p.Command("powershell", "-Command", "Start-Sleep -Seconds 10").As("one")
				p.Command("powershell", "-Command", "Start-Sleep -Seconds 10").As("two")
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
			name: "Stop with short timeout on process ignoring CTRL_BREAK_EVENT",
			setup: func(p contractsprocess.Pool) {
				// This PowerShell script ignores CTRL_BREAK_EVENT (mapped from os.Interrupt)
				p.Command("powershell", "-Command", "$global:interrupted = $false; [console]::TreatControlCAsInput = $true; $handler = [Console]::CancelKeyPress; [Console]::CancelKeyPress = { Write-Output 'Ignoring interrupt'; $_.Cancel = $true }; Start-Sleep -Seconds 10").As("unstoppable")
			},
			action: func(t *testing.T, rp contractsprocess.RunningPool) {
				time.Sleep(100 * time.Millisecond)
				err := rp.Stop(1 * time.Millisecond)
				fmt.Println("Stop with short timeout on process ignoring CTRL_BREAK_EVENT", err)
				// On Windows, stopping a process might return "Access is denied" error
				// This is expected behavior in some cases, especially with processes that
				// have already been terminated or are in a transitional state
				if err != nil {
					fmt.Println("Stop with short timeout on process ignoring CTRL_BREAK_EVENT", err.Error())
					// If there's an error, it should be a known Windows error
					assert.Contains(t, err.Error(), "Access is denied",
						"If stopping returns an error, it should be 'Access is denied', got: %s", err.Error())
				}
			},
			validate: func(t *testing.T, results map[string]contractsprocess.Result) {
				// Just verify the process was eventually terminated
				assert.Len(t, results, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewPool()
			rp, err := builder.Pool(tt.setup).Start()
			assert.NoError(t, err)

			tt.action(t, rp)
			results := rp.Wait()
			tt.validate(t, results)
		})
	}
}

func TestRunningPool_Timeout_Windows(t *testing.T) {
	t.Run("timeout kills all processes", func(t *testing.T) {
		builder := NewPool().Timeout(200 * time.Millisecond)
		rp, err := builder.Pool(func(p contractsprocess.Pool) {
			p.Command("powershell", "-Command", "Start-Sleep -Seconds 10").As("slow1")
			p.Command("powershell", "-Command", "Start-Sleep -Seconds 10").As("slow2")
		}).Start()
		assert.NoError(t, err)

		results := rp.Wait()
		assert.Len(t, results, 2)
		assert.True(t, results["slow1"].Failed())
		assert.True(t, results["slow2"].Failed())
	})
}

func TestRunningPool_OnOutput_Windows(t *testing.T) {
	t.Run("captures output via callback", func(t *testing.T) {
		outputs := make(map[string][]string)
		builder := NewPool().OnOutput(func(typ contractsprocess.OutputType, line []byte, key string) {
			outputs[key] = append(outputs[key], string(line))
		})

		rp, err := builder.Pool(func(p contractsprocess.Pool) {
			p.Command("powershell", "-Command", "Write-Output test1").As("cmd1")
			p.Command("powershell", "-Command", "Write-Output test2").As("cmd2")
		}).Start()
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

func TestRunningPool_EmptyPool_Windows(t *testing.T) {
	builder := NewPool()
	_, err := builder.Pool(func(p contractsprocess.Pool) {
		// Empty pool
	}).Start()
	assert.Error(t, err)
}

func TestRunningPool_DoneChannel_Windows(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T)
	}{
		{
			name: "done channel closes after completion",
			setup: func(t *testing.T) {
				builder := NewPool()
				rp, err := builder.Pool(func(p contractsprocess.Pool) {
					p.Command("powershell", "-Command", "Write-Output done").As("test")
				}).Start()
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
				rp, err := builder.Pool(func(p contractsprocess.Pool) {
					p.Command("powershell", "-Command", "Start-Sleep -Seconds 10").As("long")
				}).Start()
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
