package process

import (
	"context"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun_TableDriven(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name      string
		command   string
		args      []string
		wantOK    bool
		wantInOut string
	}{
		{"echo ok", echoCommand(), echoArgs("hello"), true, "hello"},
		{"non-zero exit", exitCommand(), exitArgs(1), false, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := NewCommand(tc.command, tc.args...).Run(ctx)
			// Non-zero exit should not return error from Run; only Start errors bubble up
			assert.NoError(t, err)
			if tc.wantOK {
				assert.True(t, res.Successful())
				assert.Contains(t, res.Output(), tc.wantInOut)
			} else {
				assert.True(t, res.Failed())
			}
		})
	}
}

func TestStart_Wait_OutputAndError(t *testing.T) {
	ctx := context.Background()
	cmd, args := echoBothCommand()

	r, err := NewCommand(cmd, args...).Start(ctx)
	assert.NoError(t, err)
	res := r.Wait()
	assert.True(t, res.Successful())
	assert.Contains(t, res.Output(), "out")
	assert.Contains(t, res.ErrorOutput(), "err")
}

func TestStart_Timeout(t *testing.T) {
	ctx := context.Background()
	cmd, args := sleepCommand(2)
	res, err := NewCommand(cmd, args...).Timeout(300 * time.Millisecond).Run(ctx)
	assert.NoError(t, err)
	assert.True(t, res.Failed())
}

func TestSingleBuilder_Env_And_Path(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	cmd := cwdAndEnvCommand()

	res, err := NewCommand(cmd.name, cmd.args...).Path(tmpDir).Env(map[string]string{"FOO": "BAR"}).Run(ctx)
	assert.NoError(t, err)
	out := normalizeNewlines(res.Output())
	// First line should be current directory
	assert.Contains(t, out, filepath.Clean(tmpDir))
	// Second line contains env var value
	assert.Contains(t, out, "BAR")
}

func TestOnOutput_Handler_Is_Called(t *testing.T) {
	ctx := context.Background()
	var captured []string
	h := func(typ, line string) {
		captured = append(captured, typ+":"+line)
	}

	res, err := NewCommand(echoCommand(), echoArgs("handler works")...).OnOutput(h).Run(ctx)
	assert.NoError(t, err)
	_ = res // ensure process completes
	assert.True(t, len(captured) >= 1)
}

// --- Helpers ---

func echoCommand() string {
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	return "sh"
}

func echoArgs(s string) []string {
	if runtime.GOOS == "windows" {
		return []string{"/C", "echo", s}
	}
	return []string{"-c", "echo " + shellEscape(s)}
}

func exitCommand() string {
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	return "sh"
}

func exitArgs(code int) []string {
	if runtime.GOOS == "windows" {
		return []string{"/C", "exit", strconv.Itoa(code)}
	}
	return []string{"-c", "exit " + strconv.Itoa(code)}
}

func echoBothCommand() (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C", "echo out && echo err 1>&2"}
	}
	return "sh", []string{"-c", "echo out; echo err 1>&2"}
}

func sleepCommand(seconds int) (string, []string) {
	if runtime.GOOS == "windows" {
		return "powershell", []string{"-BaseCommand", "Start-Sleep -Seconds " + strconv.Itoa(seconds)}
	}
	return "sleep", []string{strconv.Itoa(seconds)}
}

type commandSpec struct {
	name string
	args []string
}

func cwdAndEnvCommand() commandSpec {
	if runtime.GOOS == "windows" {
		return commandSpec{"cmd", []string{"/C", "cd && echo %FOO%"}}
	}
	return commandSpec{"sh", []string{"-c", "pwd; echo ${FOO}"}}
}

func shellEscape(s string) string {
	// minimal for test; inputs are static
	return strings.ReplaceAll(s, "'", "'\\''")
}

func normalizeNewlines(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
