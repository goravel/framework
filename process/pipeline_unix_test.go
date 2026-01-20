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

	"github.com/stretchr/testify/assert"

	contractsprocess "github.com/goravel/framework/contracts/process"
)

func TestPipe_ErrorOnNoSteps_Unix(t *testing.T) {
	res := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {}).Run()
	assert.Contains(t, res.Error(), "pipeline must have at least one command")
}

func TestPipe_Run_SimplePipeline_Unix(t *testing.T) {
	res := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("sh", "-c", "printf 'hello'").As("first")
		b.Command("tr", "a-z", "A-Z").As("second")
	}).Run()
	assert.True(t, res.Successful())
	assert.Equal(t, "HELLO", res.Output())
	assert.Equal(t, "", res.ErrorOutput())
}

func TestPipe_Run_Input_Path_Env_Unix(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "upper.sh")
	_ = os.WriteFile(script, []byte("#!/bin/sh\ntr 'a-z' 'A-Z'\n"), 0o755)
	_ = os.Chmod(script, 0o755)

	res := NewPipe().
		Input(bytes.NewBufferString("abc")).
		Path(dir).
		Env(map[string]string{"FOO": "BAR"}).
		Quietly().
		Pipe(func(b contractsprocess.Pipe) {
			b.Command("sh", "-c", "cat; printf \"$FOO\"").As("combine")
			b.Command("./upper.sh").As("upper")
		}).Run()

	// Expected: input "abc" + env "BAR" uppercased
	assert.True(t, res.Successful())
	assert.Equal(t, "ABCBAR", strings.TrimSpace(res.Output()))
}

func TestPipe_OnOutput_ReceivesFromEachStep_Unix(t *testing.T) {
	var byKey = map[string][]string{}
	res := NewPipe().Quietly().OnOutput(func(typ contractsprocess.OutputType, line []byte, key string) {
		byKey[key] = append(byKey[key], string(line))
	}).Pipe(func(b contractsprocess.Pipe) {
		b.Command("sh", "-c", "printf 'a\\nb\\n'").As("first")
		b.Command("cat").As("second")
	}).Run()
	assert.True(t, res.Successful())
	// We should receive lines from both the producer and the final consumer
	if assert.Contains(t, byKey, "first") {
		assert.Equal(t, []string{"a", "b"}, byKey["first"])
	}
	if assert.Contains(t, byKey, "second") {
		assert.Equal(t, []string{"a", "b"}, byKey["second"])
	}
}

func TestPipe_DisableBuffering_Unix(t *testing.T) {
	var stdoutLines, stderrLines int
	res := NewPipe().DisableBuffering().Quietly().OnOutput(func(typ contractsprocess.OutputType, line []byte, key string) {
		if typ == contractsprocess.OutputTypeStdout {
			stdoutLines++
		} else {
			stderrLines++
		}
	}).Pipe(func(b contractsprocess.Pipe) {
		b.Command("sh", "-c", "printf 'x\\n'; printf 'y\\n' 1>&2").As("only")
	}).Run()
	assert.True(t, res.Successful())
	// Buffers are disabled, so the aggregated output should be empty
	assert.Equal(t, "", res.Output())
	assert.Equal(t, "", res.ErrorOutput())
	assert.Equal(t, 1, stdoutLines)
	assert.Equal(t, 1, stderrLines)
}

func TestPipe_Timeout_Unix(t *testing.T) {
	res := NewPipe().Timeout(100 * time.Millisecond).Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("sleep", "1").As("long")
	}).Run()
	assert.True(t, res.Failed())
	assert.NotEqual(t, 0, res.ExitCode())
}

func TestPipe_Start_ErrorOnStartFailure_Unix(t *testing.T) {
	_, err := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("sleep", "1").As("ok")
		b.Command("definitely-not-a-real-binary-xyz").As("bad")
	}).Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start pipeline:")
}

func TestPipe_WithContext_Unix(t *testing.T) {
	res := NewPipe().WithContext(context.TODO()).Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("sh", "-c", "printf 'ok'")
	}).Run()
	assert.True(t, res.Successful())
}

func TestPipe_DefaultStepKeys_Unix(t *testing.T) {
	rp, err := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("sh", "-c", "printf 'a\\n'")
		b.Command("cat")
	}).Start()
	assert.NoError(t, err)
	pids := rp.PIDs()
	assert.Greater(t, pids["0"], 0)
	assert.Greater(t, pids["1"], 0)
	_ = rp.Stop(1 * time.Second)
	_ = rp.Wait()
}
