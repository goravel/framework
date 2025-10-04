package console

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/goravel/framework/support/color"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestAsk(_ *testing.T) {
	/*
			ctx := &CliContext{}
			// single line input text
			question := "How are you feeling today?"
			answer, err := ctx.Ask(question, console.AskOption{
				Default:     "Good",
				Description: "Please enter your feeling",
				Limit:       10,
				Placeholder: "Good",
				Prompt:      ">",
				Validate: func(s string) error {
					if s == "" {
						return fmt.Errorf("please enter your feeling")
					}
					return nil
				},
			})
			if err != nil {
				ctx.Error(err.Error())
				return
			}
			ctx.Info(fmt.Sprintf("You said: %s", answer))

			// multiple lines input text
			question = "tell me about yourself"
			answer, err = ctx.Ask(question, console.AskOption{
				Default:     "I am a software engineer",
				Description: "Please enter your bio",
		        Multiple:    true,
				Lines:       5,
				Placeholder: "Bio",
				Validate: func(s string) error {
					if s == "" {
						return fmt.Errorf("please enter your bio")
					}
					return nil
				},
			})

			if err != nil {
				ctx.Error(err.Error())
				return
			}

			ctx.Info(fmt.Sprintf("You said: %s", answer))
	*/
}

func TestCreateProgressBar(_ *testing.T) {
	/*
		ctx := &CliContext{}
		bar := ctx.CreateProgressBar(100)
		err := bar.Start()
		if err != nil {
			ctx.Error(err.Error())
			return
		}

		for i := 1; i < 100; i++ {
			// performTask()
			if i%2 == 0 {
				bar.Advance(2)
			} else {
				bar.Advance()
			}
			time.Sleep(time.Millisecond * 50)
		}

		err = bar.Finish()
		if err != nil {
			ctx.Error(err.Error())
			return
		}
	*/
}

func TestChoice(_ *testing.T) {
	/*
		ctx := &CliContext{}
		question := "What is your favorite programming language?"
		options := []console.Choice{
			{Key: "go", Value: "Go"},
			{Key: "php", Value: "PHP"},
			{Key: "python", Value: "Python"},
			{Key: "cpp", Value: "C++", Selected: true},
		}

		answer, err := ctx.Choice(question, options, console.ChoiceOption{
			Default:     "cpp",
			Description: "Please select your favorite programming language",
			Validate: func(s string) error {
				if s == "Python" {
					return fmt.Errorf("you can't have Python as your favorite programming language")
				}

				return nil
			},
		})

		if err != nil {
			ctx.Error(err.Error())
			return
		}

		ctx.Info(fmt.Sprintf("You selected: %s", answer))
	*/
}

func TestConfirm(_ *testing.T) {
	/*
		ctx := &CliContext{}
		question := "Are you sure you want to continue?"
		confirmation := false
		confirmed, err := ctx.Confirm(question, console.ConfirmOption{
			Affirmative: "Hell Yeah",
			Default:     &confirmation,
			Description: "Please confirm to proceed",
			Negative:    "Nah",
		})

		if err != nil {
			ctx.Error(err.Error())
			return
		}

		if confirmed {
			ctx.Info("You confirmed to proceed.")
		} else {
			ctx.Info("You declined to proceed.")
		}
	*/
}

func TestMultiSelect(_ *testing.T) {
	/*
		ctx := &CliContext{}
		question := "What are your favorite colors?"
		options := []console.Choice{
			{Key: "red", Value: "Red"},
			{Key: "blue", Value: "Blue"},
			{Key: "green", Value: "Green"},
			{Key: "yellow", Value: "Yellow", Selected: true},
			{Key: "purple", Value: "Purple"},
		}
		filterable := true
		answers, err := ctx.MultiSelect(question, options, console.MultiSelectOption{
			Default:     []string{"yellow"},
			Description: "Please select your favorite colors",
			Filterable:  &filterable,
			Limit:       3,
			Validate: func(s []string) error {
				if len(s) == 0 {
					return fmt.Errorf("please select at least one color")
				}
				return nil
			},
		})
		if err != nil {
			ctx.Error(err.Error())
			return
		}

		ctx.Info(fmt.Sprintf("You selected: %v", answers))
	*/
}

func TestNewLine(_ *testing.T) {
	/*
		ctx := &CliContext{}
		ctx.NewLine()
		ctx.NewLine(3)
	*/
}

func TestQuestion(_ *testing.T) {
	/*
		ctx := &CliContext{}
		ctx.Question("What is your name?")
	*/
}

func TestSecret(_ *testing.T) {
	/*
		ctx := &CliContext{}
		question := "What is your password?"
		password, err := ctx.Secret(question, console.SecretOption{
			Default:     "password",
			Description: "Please enter your password",
			Limit:       15,
			Placeholder: "password",
			Validate: func(s string) error {
				if len(s) < 8 {
					return fmt.Errorf("password must be at least 8 characters")
				}

				return nil
			},
		})

		if err != nil {
			ctx.Error(err.Error())
			return
		}

		ctx.Info(fmt.Sprintf("You entered: %s", password))
	*/
}

func TestSpinner(_ *testing.T) {
	/*
		ctx := &CliContext{}
		err := ctx.Spinner("Loading...", console.SpinnerOption{
			Action: func() error {
				// when to stop the spinner
				time.Sleep(2 * time.Second)
				return nil
			},
		})
		if err != nil {
			ctx.Error(err.Error())
			return
		}

		ctx.Info("Task completed successfully.")
	*/
}

func TestWarn(_ *testing.T) {
	/*
		ctx := &CliContext{}
		ctx.Warn("This is a warning message.")
	*/
}

func TestWithProgressBar(_ *testing.T) {
	/*
		ctx := &CliContext{}
		items := []any{"item1", "item2", "item3"}
		_, err := ctx.WithProgressBar(items, func(item any) error {
			// performTask(item)
			return nil
		})

		if err != nil {
			ctx.Error(err.Error())
			return
		}

		ctx.Info("Task completed successfully.")
	*/
}

func TestCliContextArguments(t *testing.T) {
	now := time.Now()
	nowUnix := now.Unix()

	testCases := []struct {
		name      string
		args      []string
		arguments []cli.Argument
		testFunc  func(t *testing.T, ctx *CliContext)
	}{
		{
			name: "Single Arguments",
			args: []string{
				"test", "string", "3.14", "3.14159", "42", "127", "32767", "2147483647", "9223372036854775807",
				"100", "255", "65535", "4294967295", "18446744073709551615", now.Format(time.RFC3339),
			},
			arguments: []cli.Argument{
				&cli.StringArgs{Name: "string-arg", Max: 1},
				&cli.Float32Args{Name: "float32-arg", Max: 1},
				&cli.Float64Args{Name: "float64-arg", Max: 1},
				&cli.IntArgs{Name: "int-arg", Max: 1},
				&cli.Int8Args{Name: "int8-arg", Max: 1},
				&cli.Int16Args{Name: "int16-arg", Max: 1},
				&cli.Int32Args{Name: "int32-arg", Max: 1},
				&cli.Int64Args{Name: "int64-arg", Max: 1},
				&cli.UintArgs{Name: "uint-arg", Max: 1},
				&cli.Uint8Args{Name: "uint8-arg", Max: 1},
				&cli.Uint16Args{Name: "uint16-arg", Max: 1},
				&cli.Uint32Args{Name: "uint32-arg", Max: 1},
				&cli.Uint64Args{Name: "uint64-arg", Max: 1},
				&cli.TimestampArgs{
					Name: "timestamp-arg",
					Max:  1,
					Config: cli.TimestampConfig{
						Layouts: []string{time.RFC3339},
					},
				},
			},
			testFunc: func(t *testing.T, ctx *CliContext) {
				assert.Equal(t, "string", ctx.ArgumentString("string-arg"))
				assert.Equal(t, float32(3.14), ctx.ArgumentFloat32("float32-arg"))
				assert.Equal(t, 3.14159, ctx.ArgumentFloat64("float64-arg"))
				assert.Equal(t, 42, ctx.ArgumentInt("int-arg"))
				assert.Equal(t, int8(127), ctx.ArgumentInt8("int8-arg"))
				assert.Equal(t, int16(32767), ctx.ArgumentInt16("int16-arg"))
				assert.Equal(t, int32(2147483647), ctx.ArgumentInt32("int32-arg"))
				assert.Equal(t, int64(9223372036854775807), ctx.ArgumentInt64("int64-arg"))
				assert.Equal(t, uint(100), ctx.ArgumentUint("uint-arg"))
				assert.Equal(t, uint8(255), ctx.ArgumentUint8("uint8-arg"))
				assert.Equal(t, uint16(65535), ctx.ArgumentUint16("uint16-arg"))
				assert.Equal(t, uint32(4294967295), ctx.ArgumentUint32("uint32-arg"))
				assert.Equal(t, uint64(18446744073709551615), ctx.ArgumentUint64("uint64-arg"))
				assert.Equal(t, nowUnix, ctx.ArgumentTimestamp("timestamp-arg").Unix())

				// Test zero values for non-existent keys
				assert.Equal(t, "", ctx.ArgumentString("non-existent"))
				assert.Equal(t, float32(0), ctx.ArgumentFloat32("non-existent"))
				assert.Equal(t, float64(0), ctx.ArgumentFloat64("non-existent"))
				assert.Equal(t, 0, ctx.ArgumentInt("non-existent"))
				assert.Equal(t, int8(0), ctx.ArgumentInt8("non-existent"))
				assert.Equal(t, int16(0), ctx.ArgumentInt16("non-existent"))
				assert.Equal(t, int32(0), ctx.ArgumentInt32("non-existent"))
				assert.Equal(t, int64(0), ctx.ArgumentInt64("non-existent"))
				assert.Equal(t, uint(0), ctx.ArgumentUint("non-existent"))
				assert.Equal(t, uint8(0), ctx.ArgumentUint8("non-existent"))
				assert.Equal(t, uint16(0), ctx.ArgumentUint16("non-existent"))
				assert.Equal(t, uint32(0), ctx.ArgumentUint32("non-existent"))
				assert.Equal(t, uint64(0), ctx.ArgumentUint64("non-existent"))
				assert.True(t, ctx.ArgumentTimestamp("non-existent").IsZero())
			},
		},
		{
			name: "Slice Arguments",
			args: []string{
				"test", "a", "b", "c",
				"1.1", "2.2", "3.3",
				"4.4", "5.5", "6.6",
				"10", "20", "30",
				"11", "22", "33",
				"12", "23", "34",
				"13", "24", "35",
				"14", "25", "36",
				"100", "200",
				"101", "201",
				"102", "202",
				"103", "203",
				"104", "204",
				now.Format(time.RFC3339), now.Add(time.Hour).Format(time.RFC3339),
			},
			arguments: []cli.Argument{
				&cli.StringArgs{Name: "string-slice-arg", Min: 1, Max: 3},
				&cli.Float32Args{Name: "float32-slice-arg", Min: 1, Max: 3},
				&cli.Float64Args{Name: "float64-slice-arg", Min: 1, Max: 3},
				&cli.IntArgs{Name: "int-slice-arg", Min: 1, Max: 3},
				&cli.Int8Args{Name: "int8-slice-arg", Min: 1, Max: 3},
				&cli.Int16Args{Name: "int16-slice-arg", Min: 1, Max: 3},
				&cli.Int32Args{Name: "int32-slice-arg", Min: 1, Max: 3},
				&cli.Int64Args{Name: "int64-slice-arg", Min: 1, Max: 3},
				&cli.UintArgs{Name: "uint-slice-arg", Min: 1, Max: 2},
				&cli.Uint8Args{Name: "uint8-slice-arg", Min: 1, Max: 2},
				&cli.Uint16Args{Name: "uint16-slice-arg", Min: 1, Max: 2},
				&cli.Uint32Args{Name: "uint32-slice-arg", Min: 1, Max: 2},
				&cli.Uint64Args{Name: "uint64-slice-arg", Min: 1, Max: 2},
				&cli.TimestampArgs{
					Name: "timestamp-slice-arg",
					Min:  1,
					Max:  2,
					Config: cli.TimestampConfig{
						Layouts: []string{time.RFC3339},
					},
				},
			},
			testFunc: func(t *testing.T, ctx *CliContext) {
				assert.Equal(t, []string{"a", "b", "c"}, ctx.ArgumentStringSlice("string-slice-arg"))
				assert.Equal(t, []float32{1.1, 2.2, 3.3}, ctx.ArgumentFloat32Slice("float32-slice-arg"))
				assert.Equal(t, []float64{4.4, 5.5, 6.6}, ctx.ArgumentFloat64Slice("float64-slice-arg"))
				assert.Equal(t, []int{10, 20, 30}, ctx.ArgumentIntSlice("int-slice-arg"))
				assert.Equal(t, []int8{11, 22, 33}, ctx.ArgumentInt8Slice("int8-slice-arg"))
				assert.Equal(t, []int16{12, 23, 34}, ctx.ArgumentInt16Slice("int16-slice-arg"))
				assert.Equal(t, []int32{13, 24, 35}, ctx.ArgumentInt32Slice("int32-slice-arg"))
				assert.Equal(t, []int64{14, 25, 36}, ctx.ArgumentInt64Slice("int64-slice-arg"))
				assert.Equal(t, []uint{100, 200}, ctx.ArgumentUintSlice("uint-slice-arg"))
				assert.Equal(t, []uint8{101, 201}, ctx.ArgumentUint8Slice("uint8-slice-arg"))
				assert.Equal(t, []uint16{102, 202}, ctx.ArgumentUint16Slice("uint16-slice-arg"))
				assert.Equal(t, []uint32{103, 203}, ctx.ArgumentUint32Slice("uint32-slice-arg"))
				assert.Equal(t, []uint64{104, 204}, ctx.ArgumentUint64Slice("uint64-slice-arg"))
				assert.Equal(t, nowUnix, ctx.ArgumentTimestampSlice("timestamp-slice-arg")[0].Unix())
				assert.Equal(t, now.Add(time.Hour).Unix(), ctx.ArgumentTimestampSlice("timestamp-slice-arg")[1].Unix())

				// Test nil values for non-existent keys
				assert.Nil(t, ctx.ArgumentStringSlice("non-existent"))
				assert.Nil(t, ctx.ArgumentFloat32Slice("non-existent"))
				assert.Nil(t, ctx.ArgumentFloat64Slice("non-existent"))
				assert.Nil(t, ctx.ArgumentIntSlice("non-existent"))
				assert.Nil(t, ctx.ArgumentInt8Slice("non-existent"))
				assert.Nil(t, ctx.ArgumentInt16Slice("non-existent"))
				assert.Nil(t, ctx.ArgumentInt32Slice("non-existent"))
				assert.Nil(t, ctx.ArgumentInt64Slice("non-existent"))
				assert.Nil(t, ctx.ArgumentUintSlice("non-existent"))
				assert.Nil(t, ctx.ArgumentUint8Slice("non-existent"))
				assert.Nil(t, ctx.ArgumentUint16Slice("non-existent"))
				assert.Nil(t, ctx.ArgumentUint32Slice("non-existent"))
				assert.Nil(t, ctx.ArgumentUint64Slice("non-existent"))
				assert.Nil(t, ctx.ArgumentTimestampSlice("non-existent"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cli.Command{
				Name:      "test",
				Arguments: tc.arguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd)
					tc.testFunc(t, cliCtx)
					return nil
				},
			}

			err := cmd.Run(context.Background(), tc.args)
			assert.NoError(t, err)
		})
	}
}

func TestDivider(t *testing.T) {
	testCases := []struct {
		name           string
		testFunc       func(ctx *CliContext)
		expectedOutput string
	}{
		{
			name: "test Divider default",
			testFunc: func(ctx *CliContext) {
				ctx.Divider()
			},
			expectedOutput: color.Default().Sprintln(strings.Repeat("-", pterm.GetTerminalWidth())),
		},
		{
			name: "test Divider empty",
			testFunc: func(ctx *CliContext) {
				ctx.Divider("")
			},
			expectedOutput: color.Default().Sprintln(strings.Repeat("-", pterm.GetTerminalWidth())),
		},
		{
			name: "test Divider char",
			testFunc: func(ctx *CliContext) {
				ctx.Divider("=")
			},
			expectedOutput: color.Default().Sprintln(strings.Repeat("=", pterm.GetTerminalWidth())),
		},
		{
			name: "test Divider multiple",
			testFunc: func(ctx *CliContext) {
				ctx.Divider("=->")
			},
			expectedOutput: color.Default().Sprintln(
				strings.Repeat("=->", pterm.GetTerminalWidth()/3) + "=->"[0:pterm.GetTerminalWidth()%3],
			),
		},
		{
			name: "test Divider multibyte",
			testFunc: func(ctx *CliContext) {
				ctx.Divider("♠")
			},
			expectedOutput: color.Default().Sprintln(strings.Repeat("♠", pterm.GetTerminalWidth())),
		},
		{
			name: "test Divider multibyte multiple",
			testFunc: func(ctx *CliContext) {
				ctx.Divider("♠♣♥")
			},
			expectedOutput: color.Default().Sprintln(
				strings.Repeat("♠♣♥", pterm.GetTerminalWidth()/3) + string([]rune("♠♣♥")[0:pterm.GetTerminalWidth()%3]),
			),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := CliContext{}
			got := color.CaptureOutput(func(io.Writer) {
				tt.testFunc(&ctx)
			})

			assert.Equal(t, got, tt.expectedOutput)
		})
	}
}

func TestColors(t *testing.T) {
	testCases := []struct {
		name           string
		testFunc       func(ctx *CliContext)
		expectedOutput string
	}{
		{
			name: "test Green",
			testFunc: func(ctx *CliContext) {
				ctx.Green("Green text")
			},
			expectedOutput: color.Green().Sprint("Green text"),
		},
		{
			name: "test Greenln",
			testFunc: func(ctx *CliContext) {
				ctx.Greenln("Green line")
			},
			expectedOutput: color.Green().Sprintln("Green line"),
		},
		{
			name: "test Red",
			testFunc: func(ctx *CliContext) {
				ctx.Red("Red text")
			},
			expectedOutput: color.Red().Sprint("Red text"),
		},
		{
			name: "test Redln",
			testFunc: func(ctx *CliContext) {
				ctx.Redln("Red line")
			},
			expectedOutput: color.Red().Sprintln("Red line"),
		},
		{
			name: "test Yellow",
			testFunc: func(ctx *CliContext) {
				ctx.Yellow("Yellow text")
			},
			expectedOutput: color.Yellow().Sprint("Yellow text"),
		},
		{
			name: "test Yellowln",
			testFunc: func(ctx *CliContext) {
				ctx.Yellowln("Yellow line")
			},
			expectedOutput: color.Yellow().Sprintln("Yellow line"),
		},
		{
			name: "test Black",
			testFunc: func(ctx *CliContext) {
				ctx.Black("Black text")
			},
			expectedOutput: color.Black().Sprint("Black text"),
		},
		{
			name: "test Blackln",
			testFunc: func(ctx *CliContext) {
				ctx.Blackln("Black line")
			},
			expectedOutput: color.Black().Sprintln("Black line"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := CliContext{}
			got := color.CaptureOutput(func(io.Writer) {
				tt.testFunc(&ctx)
			})

			assert.Equal(t, got, tt.expectedOutput)
		})
	}
}
