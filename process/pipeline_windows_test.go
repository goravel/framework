//go:build windows

package process

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	contractsprocess "github.com/goravel/framework/contracts/process"
	"github.com/goravel/framework/errors"
)

func TestPipe_ErrorOnNoSteps_Windows(t *testing.T) {
	res := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {}).Run()
	assert.Equal(t, errors.ProcessPipelineEmpty, res.Error())
}

func TestPipe_Run_SimplePipeline_Windows(t *testing.T) {
	// Step1: emit "hello" without newline; Step2: uppercase via PowerShell reading from STDIN
	res := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("cmd", "/C", "set", "/p=hello<nul").As("first")
		b.Command("powershell", "-NoLogo", "-NoProfile", "-Command",
			"$t=[Console]::In.ReadToEnd(); [Console]::Out.Write($t.ToUpper())").As("upper")
	}).Run()
	assert.True(t, res.Successful())
	assert.Equal(t, "HELLO", res.Output())
}

func TestPipe_Run_Input_Windows(t *testing.T) {
	// Provide input -> pass through cmd more -> uppercase
	res := NewPipe().Input(bytes.NewBufferString("abc")).Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("cmd", "/C", "more").As("pass")
		b.Command("powershell", "-NoLogo", "-NoProfile", "-Command",
			"$t=[Console]::In.ReadToEnd(); [Console]::Out.Write($t.ToUpper())").As("upper")
	}).Run()
	assert.True(t, res.Successful())
	assert.Equal(t, "ABC\r\n", res.Output())
}

func TestPipe_Env_Windows(t *testing.T) {
	res := NewPipe().Env(map[string]string{"FOO": "BAR"}).Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("cmd", "/C", "echo %FOO%").As("env")
	}).Run()
	// Trim CRLF
	assert.Equal(t, "BAR\r\n", res.Output())
}

func TestPipe_OnOutput_ReceivesFromEachStep_Windows(t *testing.T) {
	var byKey = map[string][]string{}
	res := NewPipe().Quietly().OnOutput(func(typ contractsprocess.OutputType, line []byte, key string) {
		byKey[key] = append(byKey[key], strings.TrimRight(string(line), "\r"))
	}).Pipe(func(b contractsprocess.Pipe) {
		b.Command("cmd", "/C", "(echo a & echo b)").As("first")
		b.Command("cmd", "/C", "more").As("second")
	}).Run()
	assert.True(t, res.Successful())
	if assert.Contains(t, byKey, "first") {
		assert.Equal(t, []string{"a ", "b"}, byKey["first"])
	}
	if assert.Contains(t, byKey, "second") {
		assert.Equal(t, []string{"a ", "b", ""}, byKey["second"])
	}
}

func TestPipe_DisableBuffering_Windows(t *testing.T) {
	var stdoutLines, stderrLines int
	res := NewPipe().DisableBuffering().Quietly().OnOutput(func(typ contractsprocess.OutputType, line []byte, key string) {
		if typ == contractsprocess.OutputTypeStdout {
			stdoutLines++
		} else {
			stderrLines++
		}
	}).Pipe(func(b contractsprocess.Pipe) {
		b.Command("cmd", "/C", "(echo x & echo y 1>&2)").As("only")
	}).Run()
	assert.True(t, res.Successful())
	assert.Equal(t, "", res.Output())
	assert.Equal(t, "", res.ErrorOutput())
	assert.Equal(t, 1, stdoutLines)
	assert.Equal(t, 1, stderrLines)
}

func TestPipe_Timeout_Windows(t *testing.T) {
	res := NewPipe().Timeout(200 * time.Millisecond).Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("powershell", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Seconds 2").As("sleep")
	}).Run()
	assert.True(t, res.Failed())
}

func TestPipe_Start_ErrorOnStartFailure_Windows(t *testing.T) {
	_, err := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("cmd", "/C", "echo ok").As("ok")
		b.Command("definitely-not-a-real-binary-xyz").As("bad")
	}).Start()
	assert.Error(t, err)
}

func TestPipe_WithContext_Windows(t *testing.T) {
	res := NewPipe().WithContext(context.Background()).Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("cmd", "/C", "echo hi").As("echo")
	}).Run()
	assert.Equal(t, "hi\r\n", res.Output())
}

func TestPipe_DefaultStepKeys_Windows(t *testing.T) {
	rp, err := NewPipe().Quietly().Pipe(func(b contractsprocess.Pipe) {
		b.Command("cmd", "/C", "(echo a & echo b)")
		b.Command("cmd", "/C", "more")
	}).Start()
	assert.NoError(t, err)
	pids := rp.PIDs()
	assert.Greater(t, pids["0"], 0)
	assert.Greater(t, pids["1"], 0)
	_ = rp.Stop(1 * time.Second)
	_ = rp.Wait()
}
