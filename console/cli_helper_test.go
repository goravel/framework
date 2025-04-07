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
			name: "print_app_help",
			containsOutput: []string{
				color.Yellow().Sprint("Usage:"),
				color.Yellow().Sprint("Global options:"),
				color.Yellow().Sprint("Available commands:"),
				color.Yellow().Sprint("test"),
				color.Green().Sprint("test:foo"),
				color.Green().Sprint("test:bar"),
			},
		},
		{
			name: "print_command_help",
			call: "help test:foo",
			containsOutput: []string{
				color.Yellow().Sprint("Description:"),
				color.Yellow().Sprint("Usage:"),
				color.Yellow().Sprint("Global options:"),
				color.Green().Sprint("-h, --help"),
				color.Green().Sprint("    --no-ansi"),
				color.Green().Sprint("-v, --version"),
				color.Yellow().Sprint("Options:"),
				color.Green().Sprint("-b, --bool"),
				color.Green().Sprint("-i, --int"),
				color.Blue().Sprint("int"),
				color.Green().Sprint("-h, --help"),
			},
		},
		{
			name: "print_command_help(check_flag_sorted)",
			call: "help --no-ansi test:foo",
			containsOutput: []string{
				`Description:
   Test command

Usage:
   test [global options] test:foo [options]

Global options:
   -h, --help       Show help
       --no-ansi    Force disable ANSI output
   -v, --version    Print the version

Options:
   -b, --bool    Bool flag [default: false]
   -i, --int     int flag [default: 0]
   -h, --help    Show help`,
			},
		},
		{
			name: "print_version",
			call: "--version",
			containsOutput: []string{
				"test " + color.Green().Sprint("test"),
			},
		},
		{
			name: "command_not_found",
			call: "not-found",
			containsOutput: []string{
				color.New(color.FgLightRed).Sprint("Command 'not-found' is not defined."),
			},
		},
		{
			name: "command_not_found(suggest)",
			call: "test",
			containsOutput: []string{
				color.New(color.FgLightRed).Sprint("Command 'test' is not defined. Did you mean one of these?"),
				color.Gray().Sprint("  test:bar"),
				color.Gray().Sprint("  test:foo"),
			},
		},
		{
			name: "command_not_found(suggest)",
			call: "fo",
			containsOutput: []string{
				color.New(color.FgLightRed).Sprint("Command 'fo' is not defined. Did you mean this?"),
				color.Gray().Sprint("  test:foo"),
			},
		},
		{
			name: "option_not_found",
			call: "test:foo --not-found",
			containsOutput: []string{
				color.Red().Sprint("The 'not-found' option does not exist."),
			},
		},
		{
			name: "option_needs_a_value",
			call: "test:foo --int",
			containsOutput: []string{
				color.Red().Sprint("The '--int' option requires a value."),
			},
		},
		{
			name: "option_value_is_not_valid",
			call: "test:foo --int not-a-number",
			containsOutput: []string{
				color.Red().Sprint("Invalid value 'not-a-number' for option 'int'."),
			},
		},
		{
			name: "no_ansi_color",
			call: "--no-ansi",
			containsOutput: []string{
				"test test",
				`Usage:
   test

Global options:
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

func TestLexicographicLess(t *testing.T) {
	tests := []struct {
		i        string
		j        string
		expected bool
	}{
		{"", "a", true},
		{"a", "", false},
		{"a", "a", false},
		{"a", "A", false},
		{"A", "a", true},
		{"aa", "a", false},
		{"a", "aa", true},
		{"a", "b", true},
		{"a", "B", true},
		{"A", "b", true},
		{"A", "B", true},
	}

	for _, tt := range tests {
		actual := lexicographicLess(tt.i, tt.j)
		assert.Equal(t, tt.expected, actual)
	}
}
