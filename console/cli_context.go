package console

import (
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
)

type CliContext struct {
	instance *cli.Command
}

func NewCliContext(instance *cli.Command) *CliContext {
	return &CliContext{instance}
}

func (r *CliContext) Ask(question string, option ...console.AskOption) (string, error) {
	var answer string
	multiple := false

	if len(option) > 0 {
		multiple = option[0].Multiple
		answer = option[0].Default
	}

	if multiple {
		input := huh.NewText().Title(question)
		if len(option) > 0 {
			input.CharLimit(option[0].Limit).Description(option[0].Description).Placeholder(option[0].Placeholder).Lines(option[0].Lines)
			if option[0].Validate != nil {
				input.Validate(option[0].Validate)
			}
		}

		err := input.Value(&answer).Run()
		if err != nil {
			return "", err
		}
	} else {
		input := huh.NewInput().Title(question)

		if len(option) > 0 {
			input.CharLimit(option[0].Limit).Description(option[0].Description).Placeholder(option[0].Placeholder).Prompt(option[0].Prompt)
			if option[0].Validate != nil {
				input.Validate(option[0].Validate)
			}
		}

		err := input.Value(&answer).Run()
		if err != nil {
			return "", err
		}
	}

	return answer, nil
}

func (r *CliContext) Argument(index int) string {
	return r.instance.Args().Get(index)
}

func (r *CliContext) Arguments() []string {
	return r.instance.Args().Slice()
}

func (r *CliContext) CreateProgressBar(total int) console.Progress {
	return NewProgressBar(total)
}

func (r *CliContext) Choice(question string, choices []console.Choice, option ...console.ChoiceOption) (string, error) {
	var answer string

	if len(option) > 0 {
		answer = option[0].Default
	}

	options := make([]huh.Option[string], len(choices))
	for i, choice := range choices {
		options[i] = huh.NewOption[string](choice.Key, choice.Value).Selected(choice.Selected)
	}

	input := huh.NewSelect[string]().Title(question).Options(options...)
	if len(option) > 0 {
		input.Description(option[0].Description)
		if option[0].Validate != nil {
			input.Validate(option[0].Validate)
		}
	}

	err := huh.NewForm(huh.NewGroup(input.Value(&answer))).Run()
	if err != nil {
		return "", err
	}
	return answer, err
}

func (r *CliContext) Comment(message string) {
	color.Debugln(message)
}

func (r *CliContext) Confirm(question string, option ...console.ConfirmOption) bool {
	input := huh.NewConfirm().Title(question)
	answer := false
	if len(option) > 0 {
		if len(option[0].Description) > 0 {
			input.Description(option[0].Description)
		}
		if len(option[0].Affirmative) > 0 {
			input.Affirmative(option[0].Affirmative)
		}
		if len(option[0].Negative) > 0 {
			input.Negative(option[0].Negative)
		}
		answer = option[0].Default
	}

	if err := input.Value(&answer).Run(); err != nil {
		r.Error(err.Error())

		return false
	}

	return answer
}

func (r *CliContext) Error(message string) {
	color.Errorln(message)
}

func (r *CliContext) Info(message string) {
	color.Infoln(message)
}

func (r *CliContext) Line(message string) {
	color.Default().Println(message)
}

func (r *CliContext) MultiSelect(question string, choices []console.Choice, option ...console.MultiSelectOption) ([]string, error) {
	var answer []string

	if len(option) > 0 {
		answer = option[0].Default
	}

	options := make([]huh.Option[string], len(choices))
	for i, choice := range choices {
		options[i] = huh.NewOption(choice.Key, choice.Value).Selected(choice.Selected)
	}

	input := huh.NewMultiSelect[string]().Title(question).Options(options...)
	if len(option) > 0 {
		input.Description(option[0].Description).Limit(option[0].Limit).Filterable(option[0].Filterable)
		if option[0].Validate != nil {
			input.Validate(option[0].Validate)
		}
	}

	err := huh.NewForm(huh.NewGroup(input.Value(&answer))).Run()
	if err != nil {
		return nil, err
	}

	return answer, err
}

func (r *CliContext) NewLine(times ...int) {
	numLines := 1
	if len(times) > 0 && times[0] > 0 {
		numLines = times[0]
	}
	for i := 0; i < numLines; i++ {
		color.Default().Println()
	}
}

func (r *CliContext) Option(key string) string {
	return r.instance.String(key)
}

func (r *CliContext) OptionSlice(key string) []string {
	return r.instance.StringSlice(key)
}

func (r *CliContext) OptionBool(key string) bool {
	return r.instance.Bool(key)
}

func (r *CliContext) OptionFloat64(key string) float64 {
	return r.instance.Float(key)
}

func (r *CliContext) OptionFloat64Slice(key string) []float64 {
	return r.instance.FloatSlice(key)
}

func (r *CliContext) OptionInt(key string) int {
	return r.instance.Int(key)
}

func (r *CliContext) OptionIntSlice(key string) []int {
	return r.instance.IntSlice(key)
}

func (r *CliContext) OptionInt64(key string) int64 {
	return r.instance.Int64(key)
}

func (r *CliContext) OptionInt64Slice(key string) []int64 {
	return r.instance.Int64Slice(key)
}

func (r *CliContext) ArgumentString(key string) string {
	ret := r.instance.StringArgs(key)
	if len(ret) > 0 {
		return ret[0]
	}
	var zero string
	return zero
}

func (r *CliContext) ArgumentStringSlice(key string) []string {
	return r.instance.StringArgs(key)
}

func (r *CliContext) ArgumentFloat32(key string) float32 {
	ret := r.instance.Float32Args(key)
	if len(ret) > 0 {
		return ret[0]
	}
	var zero float32
	return zero
}

func (r *CliContext) ArgumentFloat32Slice(key string) []float32 {
	return r.instance.Float32Args(key)
}

func (r *CliContext) ArgumentFloat64(key string) float64 {
	ret := r.instance.Float64Args(key)
	if len(ret) > 0 {
		return ret[0]
	}
	var zero float64
	return zero
}

func (r *CliContext) ArgumentFloat64Slice(key string) []float64 {
	return r.instance.Float64Args(key)
}

func (r *CliContext) ArgumentInt(key string) int {
	ret := r.instance.IntArgs(key)
	if len(ret) > 0 {
		return ret[0]
	}
	var zero int
	return zero
}

func (r *CliContext) ArgumentIntSlice(key string) []int {
	return r.instance.IntArgs(key)
}

func (r *CliContext) ArgumentInt8(key string) int8 {
	ret := r.instance.Int8Args(key)
	if len(ret) > 0 {
		return ret[0]
	}
	var zero int8
	return zero
}

func (r *CliContext) ArgumentInt8Slice(key string) []int8 {
	return r.instance.Int8Args(key)
}

func (r *CliContext) ArgumentInt16(key string) int16 {
	ret := r.instance.Int16Args(key)
	if len(ret) > 0 {
		return ret[0]
	}
	var zero int16
	return zero
}

func (r *CliContext) ArgumentInt16Slice(key string) []int16 {
	return r.instance.Int16Args(key)
}

func (r *CliContext) ArgumentInt32(key string) int32 {
	ret := r.instance.Int32Args(key)
	if len(ret) > 0 {
		return ret[0]
	}
	var zero int32
	return zero
}

func (r *CliContext) ArgumentInt32Slice(key string) []int32 {
	return r.instance.Int32Args(key)
}

func (r *CliContext) ArgumentInt64(key string) int64 {
	ret := r.instance.Int64Args(key)
	if len(ret) > 0 {
		return ret[0]
	}
	var zero int64
	return zero
}

func (r *CliContext) ArgumentInt64Slice(key string) []int64 {
	return r.instance.Int64Args(key)
}

func (r *CliContext) ArgumentUint(key string) uint {
	ret := r.instance.UintArgs(key)
	if len(ret) > 0 {
		return ret[0]
	}
	var zero uint
	return zero
}

func (r *CliContext) ArgumentUintSlice(key string) []uint {
	return r.instance.UintArgs(key)
}

func (r *CliContext) ArgumentUint8(key string) uint8 {
	ret := r.instance.Uint8Args(key)
	if len(ret) > 0 {
		return ret[0]
	}
	var zero uint8
	return zero
}

func (r *CliContext) ArgumentUint8Slice(key string) []uint8 {
	return r.instance.Uint8Args(key)
}

func (r *CliContext) ArgumentUint16(key string) uint16 {
	ret := r.instance.Uint16Args(key)
	if len(ret) > 0 {
		return ret[0]
	}
	var zero uint16
	return zero
}

func (r *CliContext) ArgumentUint16Slice(key string) []uint16 {
	return r.instance.Uint16Args(key)
}

func (r *CliContext) ArgumentUint32(key string) uint32 {
	ret := r.instance.Uint32Args(key)
	if len(ret) > 0 {
		return ret[0]
	}
	var zero uint32
	return zero
}

func (r *CliContext) ArgumentUint32Slice(key string) []uint32 {
	return r.instance.Uint32Args(key)
}

func (r *CliContext) ArgumentUint64(key string) uint64 {
	ret := r.instance.Uint64Args(key)
	if len(ret) > 0 {
		return ret[0]
	}
	var zero uint64
	return zero
}

func (r *CliContext) ArgumentUint64Slice(key string) []uint64 {
	return r.instance.Uint64Args(key)
}

func (r *CliContext) ArgumentTimestamp(key string) time.Time {
	ret := r.instance.TimestampArgs(key)
	if len(ret) > 0 {
		return ret[0]
	}
	return time.Time{}
}

func (r *CliContext) ArgumentTimestampSlice(key string) []time.Time {
	return r.instance.TimestampArgs(key)
}

func (r *CliContext) Secret(question string, option ...console.SecretOption) (string, error) {
	var answer string
	if len(option) > 0 {
		answer = option[0].Default
	}

	input := huh.NewInput().Title(question)

	if len(option) > 0 {
		input.CharLimit(option[0].Limit).Description(option[0].Description).Placeholder(option[0].Placeholder).EchoMode(huh.EchoModePassword)
		if option[0].Validate != nil {
			input.Validate(option[0].Validate)
		}
	}

	err := input.Value(&answer).Run()
	if err != nil {
		return "", err
	}

	return answer, nil
}

func (r *CliContext) Spinner(message string, option console.SpinnerOption) error {
	style := lipgloss.NewStyle().Foreground(lipgloss.CompleteColor{TrueColor: "#3D8C8D", ANSI256: "30", ANSI: "6"})
	spin := spinner.New().Title(message).Style(style).TitleStyle(style)

	var err error
	spin.Context(option.Ctx).Action(func() {
		err = option.Action()
	})
	if err := spin.Run(); err != nil {
		return err
	}

	return err
}

func (r *CliContext) Success(message string) {
	color.Successln(message)
}

func (r *CliContext) Warning(message string) {
	color.Warningln(message)
}

func (r *CliContext) WithProgressBar(items []any, callback func(any) error) ([]any, error) {
	bar := r.CreateProgressBar(len(items))
	err := bar.Start()
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		err := callback(item)
		if err != nil {
			return nil, err
		}
		bar.Advance()
	}

	err = bar.Finish()
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *CliContext) TwoColumnDetail(first, second string, filler ...rune) {
	r.Line(supportconsole.TwoColumnDetail(first, second, filler...))
}

func (r *CliContext) Divider(filler ...rune) {
	width := pterm.GetTerminalWidth()
	char := string(append(filler, '-')[0])

	r.Line(strings.Repeat(char, width))
}

func (r *CliContext) Green(message string) {
	color.Green().Print(message)
}

func (r *CliContext) Greenln(message string) {
	color.Green().Println(message)
}

func (r *CliContext) Red(message string) {
	color.Red().Print(message)
}

func (r *CliContext) Redln(message string) {
	color.Red().Println(message)
}

func (r *CliContext) Yellow(message string) {
	color.Yellow().Print(message)
}

func (r *CliContext) Yellowln(message string) {
	color.Yellow().Println(message)
}

func (r *CliContext) Black(message string) {
	color.Black().Print(message)
}

func (r *CliContext) Blackln(message string) {
	color.Black().Println(message)
}
