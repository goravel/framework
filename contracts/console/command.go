package console

import (
	"context"

	"github.com/goravel/framework/contracts/console/command"
)

type Command interface {
	// Signature set the unique signature for the command.
	Signature() string
	// Description the console command description.
	Description() string
	// Extend the console command extend.
	Extend() command.Extend
	// Handle execute the console command.
	Handle(ctx Context) error
}

type Context interface {
	// Ask prompts the user for input.
	Ask(question string, option ...AskOption) (string, error)
	// Choice prompts the user to select from a list of options.
	Choice(question string, options []Choice, option ...ChoiceOption) (string, error)
	// Comment writes a comment message to the console.
	Comment(message string)
	// Confirm prompts the user for a confirmation.
	Confirm(question string, option ...ConfirmOption) (bool, error)
	// Argument get the value of a command argument.
	Argument(index int) string
	// Arguments get all the arguments passed to command.
	Arguments() []string
	// Info writes an information message to the console.
	Info(message string)
	// Error writes an error message to the console.
	Error(message string)
	// Line writes a string to the console.
	Line(message string)
	// MultiSelect prompts the user to select multiple options from a list of options.
	MultiSelect(question string, options []Choice, option ...MultiSelectOption) ([]string, error)
	// NewLine writes a newline character to the console.
	NewLine(times ...int)
	// Option gets the value of a command option.
	Option(key string) string
	// OptionSlice looks up the value of a local StringSliceFlag, returns nil if not found
	OptionSlice(key string) []string
	// OptionBool looks up the value of a local BoolFlag, returns false if not found
	OptionBool(key string) bool
	// OptionFloat64 looks up the value of a local Float64Flag, returns zero if not found
	OptionFloat64(key string) float64
	// OptionFloat64Slice looks up the value of a local Float64SliceFlag, returns nil if not found
	OptionFloat64Slice(key string) []float64
	// OptionInt looks up the value of a local IntFlag, returns zero if not found
	OptionInt(key string) int
	// OptionIntSlice looks up the value of a local IntSliceFlag, returns nil if not found
	OptionIntSlice(key string) []int
	// OptionInt64 looks up the value of a local Int64Flag, returns zero if not found
	OptionInt64(key string) int64
	// OptionInt64Slice looks up the value of a local Int64SliceFlag, returns nil if not found
	OptionInt64Slice(key string) []int64
	// Question writes a question to the console.
	Question(question string)
	// Secret prompts the user for a password.
	Secret(question string, option ...SecretOption) (string, error)
	// Spinner creates a new spinner instance.
	Spinner(message string, option ...SpinnerOption) error
	// Warn writes a warning message to the console.
	Warn(message string)
}

type Choice struct {
	Key      string
	Value    string
	Selected bool
}

type AskOption struct {
	Default     string
	Prompt      string
	Multiple    bool
	Validate    func(string) error
	Limit       int
	Description string
	Placeholder string
	Lines       int
}

type ChoiceOption struct {
	Default     string
	Validate    func(string) error
	Description string
}

type ConfirmOption struct {
	Default     bool
	Description string
	Affirmative string
	Negative    string
}

type SecretOption struct {
	Default     string
	Validate    func(string) error
	Limit       int
	Description string
	Placeholder string
}

type MultiSelectOption struct {
	Default     []string
	Validate    func([]string) error
	Description string
}

type SpinnerOption struct {
	Ctx    context.Context
	Action func()
}
