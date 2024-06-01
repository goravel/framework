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
	// CreateProgressBar creates a new progress bar instance.
	CreateProgressBar(total int) Progress
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
	// Secret prompts the user for a password.
	Secret(question string, option ...SecretOption) (string, error)
	// Spinner creates a new spinner instance.
	Spinner(message string, option SpinnerOption) error
	// Warning writes a warning message to the console.
	Warning(message string)
	// WithProgressBar executes a callback with a progress bar.
	WithProgressBar(items []any, callback func(any) error) ([]any, error)
}

type Progress interface {
	// Advance advances the progress bar by a given step.
	Advance(step ...int)
	// Finish completes the progress bar.
	Finish() error
	// ShowElapsedTime sets if the elapsed time should be displayed in the progress bar.
	ShowElapsedTime(b ...bool) Progress
	// ShowTitle sets the title of the progress bar.
	ShowTitle(b ...bool) Progress
	// SetTitle sets the message of the progress bar.
	SetTitle(message string)
	// Start starts the progress bar.
	Start() error
}

type Choice struct {
	// Key the choice key.
	Key string
	// Selected determines if the choice is selected.
	Selected bool
	// Value the choice value.
	Value string
}

type AskOption struct {
	// Default the default value for the input.
	Default string
	// Description the input description.
	Description string
	// Lines the number of lines for the input.(use for multiple lines text)
	Lines int
	// Limit the character limit for the input.
	Limit int
	// Multiple determines if input is single line or multiple lines text
	Multiple bool
	// Placeholder the input placeholder.
	Placeholder string
	// Prompt the prompt message.(use for single line input)
	Prompt string
	// Validate the input validation function.
	Validate func(string) error
}

type ChoiceOption struct {
	// Default the default value for the input.
	Default string
	// Description the input description.
	Description string
	// Validate the input validation function.
	Validate func(string) error
}

type ConfirmOption struct {
	// Affirmative label for the affirmative button.
	Affirmative string
	// Default the default value for the input.
	Default bool
	// Description the input description.
	Description string
	// Negative label for the negative button.
	Negative string
}

type SecretOption struct {
	// Default the default value for the input.
	Default string
	// Description the input description.
	Description string
	// Limit the character limit for the input.
	Limit int
	// Placeholder the input placeholder.
	Placeholder string
	// Validate the input validation function.
	Validate func(string) error
}

type MultiSelectOption struct {
	// Default the default value for the input.
	Default []string
	// Description the input description.
	Description string
	// Filterable determines if the choices can be filtered.
	Filterable bool
	// Limit the number of choices that can be selected.
	Limit int
	// Validate the input validation function.
	Validate func([]string) error
}

type SpinnerOption struct {
	// Ctx the context for the spinner.
	Ctx context.Context
	// Action the action to execute.
	Action func() error
}
