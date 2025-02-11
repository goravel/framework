package console

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
)

func TestShowCommandHelp_HelpPrinterCustom(t *testing.T) {
	cliApp := NewApplication("test", "test", "test", "test", true)
	cliApp.Register([]console.Command{
		&TestFooCommand{},
		&TestBarCommand{},
	})
	tests := []struct {
		name           string
		call           string
		containsOutput []string
	}{
		{
			name: "print app help",
			containsOutput: []string{
				color.Yellow().Sprint("Usage:"),
				color.Yellow().Sprint("Options:"),
				color.Yellow().Sprint("Available commands:"),
				color.Yellow().Sprint("test"),
				color.Green().Sprint("test:foo"),
				color.Green().Sprint("test:bar"),
			},
		},
		{
			name: "print command help",
			call: "help test:foo",
			containsOutput: []string{
				color.Yellow().Sprint("Description:"),
				color.Yellow().Sprint("Usage:"),
				color.Yellow().Sprint("Options:"),
				color.Green().Sprint("-b, --bool"),
				color.Green().Sprint("-i, --int"),
				color.Blue().Sprint("int"),
			},
		},
		{
			name: "print version",
			call: "--version",
			containsOutput: []string{
				"test " + color.Green().Sprint("test"),
			},
		},
		{
			name: "command not found",
			call: "not-found",
			containsOutput: []string{
				color.New(color.FgLightRed).Sprint("Command 'not-found' is not defined."),
			},
		},
		{
			name: "command not found(suggest)",
			call: "test",
			containsOutput: []string{
				color.New(color.FgLightRed).Sprint("Command 'test' is not defined. Did you mean one of these?"),
				color.Gray().Sprint("  test:bar"),
				color.Gray().Sprint("  test:foo"),
			},
		},
		{
			name: "command not found(suggest)",
			call: "fo",
			containsOutput: []string{
				color.New(color.FgLightRed).Sprint("Command 'fo' is not defined. Did you mean this?"),
				color.Gray().Sprint("  test:foo"),
			},
		},
		{
			name: "option not found",
			call: "test:foo --not-found",
			containsOutput: []string{
				color.Red().Sprint("The 'not-found' option does not exist."),
			},
		},
		{
			name: "option needs a value",
			call: "test:foo --int",
			containsOutput: []string{
				color.Red().Sprint("The 'int' option requires a value."),
			},
		},
		{
			name: "option value is not valid",
			call: "test:foo --int not-a-number",
			containsOutput: []string{
				color.Red().Sprint("Invalid value 'not-a-number' for option 'int'."),
			},
		},
		{
			name: "no ansi color",
			call: "--no-ansi",
			containsOutput: []string{
				"test test",
				`Usage:
   test

Options:
   -h, --help       Show help
       --no-ansi    Force disable ANSI output
   -v, --version    Print the version`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			cliApp.(*Application).instance.Writer = output
			got := color.CaptureOutput(func(io.Writer) {
				assert.NoError(t, cliApp.Call(tt.call))
			})
			if len(got) == 0 {
				got = output.String()
			}
			for _, contain := range tt.containsOutput {
				assert.Contains(t, got, contain)
			}
		})
	}
}

type TestFooCommand struct {
}

type TestBarCommand struct {
	TestFooCommand
}

func (receiver *TestFooCommand) Signature() string {
	return "test:foo"
}

func (receiver *TestFooCommand) Description() string {
	return "Test command"
}

func (receiver *TestFooCommand) Extend() command.Extend {
	return command.Extend{
		Category: "test",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "bool",
				Aliases: []string{"b"},
				Usage:   "bool flag",
			},
			&command.IntFlag{
				Name:    "int",
				Aliases: []string{"i"},
				Usage:   "<fg=blue>int</> flag",
			},
		},
	}
}

func (receiver *TestFooCommand) Handle(_ console.Context) error {

	return nil
}

func (receiver *TestBarCommand) Signature() string {
	return "test:bar"
}
