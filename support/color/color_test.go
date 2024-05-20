package color

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

// captureStdout simulates capturing of os.stdout with a buffer and returns what was written to the screen
func captureOutput(f func(w io.Writer)) string {
	var outBuf bytes.Buffer
	pterm.SetDefaultOutput(&outBuf)
	f(&outBuf)

	content := outBuf.String()
	outBuf.Reset()
	return content
}

func TestColors(t *testing.T) {
	colors := map[string]Printer{
		"FgBlack":        Black(),
		"FgRed":          Red(),
		"FgGreen":        Green(),
		"FgYellow":       Yellow(),
		"FgBlue":         Blue(),
		"FgMagenta":      Magenta(),
		"FgCyan":         Cyan(),
		"FgWhite":        White(),
		"FgDefault":      Normal(),
		"FgDarkGray":     FgDarkGray,
		"FgLightRed":     FgLightRed,
		"FgLightGreen":   FgLightGreen,
		"FgLightYellow":  FgLightYellow,
		"FgLightBlue":    FgLightBlue,
		"FgLightMagenta": FgLightMagenta,
		"FgLightCyan":    FgLightCyan,
		"FgLightWhite":   FgLightWhite,
		"FgGray":         Gray(),
	}

	for name, color := range colors {
		t.Run(name, func(t *testing.T) {
			testString := "test"
			format := "%s"
			expected := fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, testString)
			expectedLn := fmt.Sprintf("\x1b[%dm%s\x1b[0m\n\x1b[%dm\x1b[0m", color, testString, color)

			assert.Equal(t, expected, color.Sprint(testString))
			assert.Equal(t, expectedLn, color.Sprintln(testString))
			assert.Equal(t, expected, color.Sprintf(format, testString))
			assert.Equal(t, expectedLn, color.Sprintfln(format, testString))

			assert.Equal(t, expected, captureOutput(func(w io.Writer) { color.Print(testString) }))
			assert.Equal(t, expectedLn, captureOutput(func(w io.Writer) { color.Println(testString) }))
			assert.Equal(t, expected, captureOutput(func(w io.Writer) { color.Printf(format, testString) }))
			assert.Equal(t, expectedLn, captureOutput(func(w io.Writer) { color.Printfln(format, testString) }))
		})
	}
}
