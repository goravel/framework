package color

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorHTMLLikeTag(t *testing.T) {
	var tests = []struct {
		name     string
		actual   string
		expected string
	}{
		{
			name:     "print with style tag:red",
			actual:   CaptureOutput(func(io.Writer) { Print("<red>MSG</>") }),
			expected: "\x1b[0;31mMSG\x1b[0m",
		},
		{
			name:     "print with style tag:bold",
			actual:   CaptureOutput(func(io.Writer) { Println("<bold>MSG</>") }),
			expected: "\x1b[1mMSG\x1b[0m\n",
		},
		{
			name:     "print with style tag:info",
			actual:   CaptureOutput(func(io.Writer) { Printf("<info>%s</>", "MSG") }),
			expected: "\x1b[0;32mMSG\x1b[0m",
		},

		{
			name:     "print with color attributes tag:fg=red",
			actual:   CaptureOutput(func(io.Writer) { Printfln("<fg=red>%s</>", "MSG") }),
			expected: "\x1b[31mMSG\x1b[0m\n",
		},
		{
			name:     "print with color attributes tag:bg=blue",
			actual:   CaptureOutput(func(io.Writer) { Print("<bg=blue>MSG</>") }),
			expected: "\x1b[44mMSG\x1b[0m",
		},
		{
			name:     "print with color attributes tag:fg=red;bg=blue",
			actual:   Sprint("<fg=red;bg=blue>MSG</>"),
			expected: "\x1b[31;44mMSG\x1b[0m",
		},
		{
			name:     "print with color attributes tag:op=bold",
			actual:   Sprintln("<op=bold>MSG</>"),
			expected: "\x1b[1mMSG\x1b[0m\n",
		},
		{
			name:     "print with color attributes tag:fg=red;bg=blue;op=bold",
			actual:   Sprintf("<fg=red;bg=blue;op=bold>%s</>", "MSG"),
			expected: "\x1b[31;44;1mMSG\x1b[0m",
		},
		{
			name:     "print with color attributes tag:fg=11aa23;bg=120,35,156;op=underscore",
			actual:   Sprintfln("<fg=11aa23;bg=120,35,156;op=underscore>%s</>", "MSG"),
			expected: "\x1b[38;2;17;170;35;48;2;120;35;156;4mMSG\x1b[0m\n",
		},
	}
	for _, tt := range tests {
		_ = tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.actual)
		})
	}
}
