package color

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColor(t *testing.T) {
	t.Run("TestColor", func(t *testing.T) {
		red := New(FgRed)

		assert.Equal(t, fmt.Sprintf("\x1b[%dm%s\x1b[0m", FgRed, "test"), red.Sprint("test"))
		assert.Equal(t, fmt.Sprintf("\x1b[%dm%s\x1b[0m\n\x1b[%dm\x1b[0m", FgRed, "test", FgRed), red.Sprintln("test"))
		assert.Equal(t, fmt.Sprintf("\x1b[%dm%s\x1b[0m", FgRed, "test"), red.Sprintf("%s", "test"))
		assert.Equal(t, fmt.Sprintf("\x1b[%dm%s\x1b[0m\n\u001B[%dm\u001B[0m", FgRed, "test", FgRed), red.Sprintfln("%s", "test"))
	})

}
