//go:build !windows

package process

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	contractsprocess "github.com/goravel/framework/contracts/process"
	"github.com/stretchr/testify/assert"
)

func TestProcess_Run_Echo_Unix(t *testing.T) {
	res, err := New().Command("sh", "-c", "printf 'hello'").Quietly().Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "hello", res.Output())
	assert.True(t, res.Successful())
}

func TestProcess_Run_StderrAndNonZero_Unix(t *testing.T) {
	res, err := New().Command("sh", "-c", "printf 'bad' 1>&2; exit 2").Quietly().Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "bad", res.ErrorOutput())
	assert.False(t, res.Successful())
	assert.NotEqual(t, 0, res.ExitCode())
}

func TestProcess_Run_EnvPassed_Unix(t *testing.T) {
	res, err := New().Command("sh", "-c", "printf \"$FOO\"").Env(map[string]string{"FOO": "BAR"}).Quietly().Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "BAR", res.Output())
}

func TestProcess_Run_WorkingDirectory_Unix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "script.sh")
	_ = os.WriteFile(path, []byte("#!/bin/sh\nprintf 'ok'\n"), 0o755)
	_ = os.Chmod(path, 0o755)
	res, err := New().Path(dir).Quietly().Command("./script.sh").Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "ok", res.Output())
}

func TestProcess_Run_InputPiped_Unix(t *testing.T) {
	res, err := New().Command("sh", "-c", "cat").Input(bytes.NewBufferString("ping")).Quietly().Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "ping", res.Output())
}

func TestProcess_Run_Timeout_Unix(t *testing.T) {
	res, err := New().Command("sh", "-c", "sleep 1").Timeout(100 * time.Millisecond).Quietly().Run(context.Background())
	// On timeout, err may or may not be set depending on platform; rely on Result
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.False(t, res.Successful())
}

func TestProcess_OnOutput_Callbacks_Unix(t *testing.T) {
	var outLines, errLines [][]byte
	p := New().Command("sh", "-c", "printf 'a\n'; printf 'b\n' 1>&2")
	p.OnOutput(func(typ contractsprocess.OutputType, line []byte) {
		if typ == contractsprocess.OutputTypeStdout {
			outLines = append(outLines, append([]byte(nil), line...))
		} else {
			errLines = append(errLines, append([]byte(nil), line...))
		}
	}).Quietly()
	res, err := p.Run(context.Background())
	assert.NoError(t, err)
	assert.True(t, res.Successful())
	if assert.NotEmpty(t, outLines) {
		assert.Equal(t, "a", strings.TrimSpace(string(outLines[0])))
	}
	if assert.NotEmpty(t, errLines) {
		assert.Equal(t, "b", strings.TrimSpace(string(errLines[0])))
	}
}

func TestProcess_Start_ErrorOnMissingCommand_Unix(t *testing.T) {
	_, err := New().Command("").Start(context.Background())
	assert.Error(t, err)
}
