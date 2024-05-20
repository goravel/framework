package color

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func captureOutput(f func()) string {
	return ""
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

			//assert.Equal(t, expected, captureOutput(func() { print(color.Print(testString)) }))
			//assert.Equal(t, expectedLn, captureOutput(func() { println(color.Println(testString)) }))
			//assert.Equal(t, expected, captureOutput(func() { print(color.Printf(format, testString)) }))
			//assert.Equal(t, expectedLn, captureOutput(func() { println(color.Printfln(format, testString)) }))
		})
	}
}
