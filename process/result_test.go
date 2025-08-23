package process

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResultMethods(t *testing.T) {
	r := &Result{
		exitCode: 0,
		command:  "echo hello",
		duration: 123 * time.Millisecond,
		stdout:   "hello world",
		stderr:   "",
	}

	assert.True(t, r.Successful())
	assert.False(t, r.Failed())
	assert.Equal(t, 0, r.ExitCode())
	assert.Equal(t, "hello world", r.Output())
	assert.Equal(t, "", r.ErrorOutput())
	assert.Equal(t, "echo hello", r.Command())
	assert.Equal(t, 123*time.Millisecond, r.Duration())
	assert.True(t, r.SeeInOutput("world"))
	assert.False(t, r.SeeInOutput("missing"))
}
