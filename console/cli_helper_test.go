package console

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/console/console"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
)

func TestShowCommandHelp_HelpPrinterCustom(t *testing.T) {
	output := &bytes.Buffer{}
	cliApp := &Application{
		name:       "artisan",
		usage:      "Goravel Framework",
		usageText:  "artisan command [options] [arguments...]",
		useArtisan: true,
		version:    "v1.16.0",
		writer:     output,
	}
	cliApp.Register([]contractsconsole.Command{
		&TestFooCommand{},
		&TestBarCommand{},
		console.NewHelpCommand(),
	})

	tests := []struct {
		name           string
		call           string
		containsOutput []string
	}{
		{
			name: "print app help",
			call: "help",
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
			name: "print command help",
			call: "test:foo --help",
			containsOutput: []string{
				color.Yellow().Sprint("Description:"),
				color.Yellow().Sprint("Usage:"),
				color.Yellow().Sprint("Global options:"),
				color.Yellow().Sprint("Options:"),
				color.Green().Sprint("-b, --bool"),
				color.Green().Sprint("-i, --int"),
				color.Blue().Sprint("int"),
				color.Green().Sprint("    --no-ansi"),
				color.Green().Sprint("-h, --help"),
			},
		},
		{
			name: "print command help(check flag sorted)",
			call: "test:foo --help --no-ansi",
			containsOutput: []string{
				`Description:
   Test command

Usage:
   artisan test:foo [options] <string_arg> <two_string_args...> [uint16_arg] [any_count_string_args...]

Options:
   -b, --bool       Bool flag [default: false]
   -h, --help       Show help
   -i, --int        int flag [default: 0]
       --no-ansi    Force disable ANSI output`,
			},
		},
		{
			name: "print version",
			call: "--version",
			containsOutput: []string{
				"Goravel Framework " + color.Green().Sprint("v1.16.0"),
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
				color.Red().Sprint("The '--int' option requires a value."),
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
			name: "argument need a value",
			call: "test:foo --int 0",
			containsOutput: []string{
				color.Red().Sprint("The 'string_arg' argument requires a value."),
			},
		},
		{
			name: "argument need a few values",
			call: "test:foo --int 0 string_arg",
			containsOutput: []string{
				color.Red().Sprint("The 'two_string_args' argument requires at least 2 values."),
			},
		},
		{
			name: "argument value is not valid",
			call: "test:foo --int 0 string_arg string_args1 string_args2 not-a-number",
			containsOutput: []string{
				color.Red().Sprint("Invalid value 'not-a-number' for argument 'uint16_arg'. Error: strconv.ParseUint: parsing \"not-a-number\": invalid syntax"),
			},
		},
		{
			name: "no ansi color",
			call: "--no-ansi",
			containsOutput: []string{
				`Goravel Framework v1.16.0

Usage:
   artisan command [options] [arguments...]

Global options:
       --no-ansi    Force disable ANSI output
   -v, --version    Print the version

Available commands:
  help      Shows a list of commands
 test:
  test:bar  Test command
  test:foo  Test command`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		Arguments: []command.Argument{
			&command.ArgumentString{
				Name:     "string_arg",
				Usage:    "string argument",
				Required: true,
			},
			&command.ArgumentStringSlice{
				Name:  "two_string_args",
				Usage: "string arguments",
				Min:   2,
				Max:   2,
			},
			&command.ArgumentUint16{
				Name:  "uint16_arg",
				Usage: "uint16 argument",
			},
			&command.ArgumentStringSlice{
				Name:  "any_count_string_args",
				Usage: "string arguments",
				Min:   0,
				Max:   -1,
			},
		},
	}
}

func (receiver *TestFooCommand) Handle(_ contractsconsole.Context) error {
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
