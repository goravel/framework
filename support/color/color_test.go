package color

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/support"
)

func TestColors(t *testing.T) {
	colors := map[string]support.Printer{
		"FgBlack":        Black(),
		"FgRed":          Red(),
		"FgGreen":        Green(),
		"FgYellow":       Yellow(),
		"FgBlue":         Blue(),
		"FgMagenta":      Magenta(),
		"FgCyan":         Cyan(),
		"FgWhite":        White(),
		"FgDefault":      Default(),
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

			assert.Equal(t, expected, CaptureOutput(func(w io.Writer) { color.Print(testString) }))
			assert.Equal(t, expectedLn, CaptureOutput(func(w io.Writer) { color.Println(testString) }))
			assert.Equal(t, expected, CaptureOutput(func(w io.Writer) { color.Printf(format, testString) }))
			assert.Equal(t, expectedLn, CaptureOutput(func(w io.Writer) { color.Printfln(format, testString) }))
		})
	}
}

func TestDisableEnable(t *testing.T) {
	Disable()
	t.Cleanup(Enable)
	assert.Equal(t, "test", Sprint("<red>test</>"))
	Enable()
	assert.Equal(t, "\x1b[0;31mtest\x1b[0m", Sprint("<red>test</>"))
}

func TestPrefixPrintersRenderTags(t *testing.T) {
	out := CaptureOutput(func(w io.Writer) { Infof("<red>%s</>", "x") })
	assert.Contains(t, out, "\x1b[0;31mx\x1b[0m")
	out = CaptureOutput(func(w io.Writer) { Errorln("<fg=blue>boom</>") })
	assert.Contains(t, out, "\x1b[34mboom\x1b[0m")
	out = CaptureOutput(func(w io.Writer) { Warningln("<fg=yellow>warn</>") })
	assert.Contains(t, out, "\x1b[33mwarn\x1b[0m")
	out = CaptureOutput(func(w io.Writer) { Successln("<green>ok</>") })
	assert.Contains(t, out, "\x1b[0;32mok\x1b[0m")
}

func TestLnHelpersKeepSpacing(t *testing.T) {
	out := CaptureOutput(func(w io.Writer) { Infoln("Rolled back:", "2024_01_01_000000_create_users_table") })
	assert.Contains(t, out, "Rolled back: 2024_01_01_000000_create_users_table")
}
