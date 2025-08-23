package process

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommandBuilderMethods(t *testing.T) {
	cmd := &BaseCommand{
		name: "echo",
		args: []string{"hello"},
	}

	t.Run("Path sets working directory", func(t *testing.T) {
		c := *cmd
		c.Path("/tmp")
		assert.Equal(t, "/tmp", c.dir)
	})

	t.Run("Env appends key=value pairs", func(t *testing.T) {
		c := *cmd
		c.Env(map[string]string{"FOO": "BAR", "HELLO": "WORLD"})
		joined := strings.Join(c.env, ",")
		assert.Contains(t, joined, "FOO=BAR")
		assert.Contains(t, joined, "HELLO=WORLD")
	})

	t.Run("Input sets stdin reader", func(t *testing.T) {
		c := *cmd
		reader := strings.NewReader("input")
		c.Input(reader)
		assert.NotNil(t, c.stdin)
		_, ok := c.stdin.(io.Reader)
		assert.True(t, ok)
	})

	t.Run("Timeout and IdleTimeout set durations", func(t *testing.T) {
		c := *cmd
		c.Timeout(1500 * time.Millisecond)
		c.IdleTimeout(2 * time.Second)
		assert.Equal(t, 1500*time.Millisecond, c.timeout)
		assert.Equal(t, 2*time.Second, c.idleTimeout)
	})

	t.Run("Quietly and Tty flags", func(t *testing.T) {
		c := *cmd
		c.Quietly()
		c.Tty()
		assert.True(t, c.quietly)
		assert.True(t, c.tty)
	})

	t.Run("OnOutput assigns handler", func(t *testing.T) {
		c := *cmd
		var captured string
		c.OnOutput(func(typ, line string) { captured = typ + ":" + line })
		assert.NotNil(t, c.outputHandler)
		c.outputHandler("stdout", "foo")
		assert.Equal(t, "stdout:foo", captured)
	})
}
