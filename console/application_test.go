package console

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

var testCommand = 0

func TestRun(t *testing.T) {
	cliApp := NewApplication("test", "test", "test", "test", true)
	cliApp.Register([]console.Command{
		&TestCommand{},
	})

	assert.NoError(t, cliApp.Call("test"))
	assert.Equal(t, 1, testCommand)
}

func TestFlagsToCliFlags(t *testing.T) {
	// Mock flags of different types
	boolFlag := &command.BoolFlag{Name: "boolFlag", Aliases: []string{"bf"}, Usage: "bool flag", Required: false, Value: false}
	float64Flag := &command.Float64Flag{Name: "float64Flag", Aliases: []string{"ff"}, Usage: "float64 flag", Required: true, Value: 1.0}
	float64SliceFlag := &command.Float64SliceFlag{Name: "float64SliceFlag", Aliases: []string{"fsf"}, Usage: "float64 slice flag", Required: false, Value: []float64{1.0, 2.0, 3.0}}
	intFlag := &command.IntFlag{Name: "intFlag", Aliases: []string{"if"}, Usage: "int flag", Required: true, Value: 1}
	intSliceFlag := &command.IntSliceFlag{Name: "intSliceFlag", Aliases: []string{"isf"}, Usage: "int slice flag", Required: false, Value: []int{1, 2, 3}}
	int64Flag := &command.Int64Flag{Name: "int64Flag", Aliases: []string{"i64f"}, Usage: "int64 flag", Required: false, Value: 1}
	int64SliceFlag := &command.Int64SliceFlag{Name: "int64SliceFlag", Aliases: []string{"i64sf"}, Usage: "int64 slice flag", Required: false, Value: []int64{1, 2, 3}}
	stringFlag := &command.StringFlag{Name: "stringFlag", Aliases: []string{"sf"}, Usage: "string flag", Required: false, Value: "default"}
	stringSliceFlag := &command.StringSliceFlag{Name: "stringSliceFlag", Aliases: []string{"ssf"}, Usage: "string slice flag", Required: false, Value: []string{"a", "b", "c"}}

	// Create a slice of command flags
	flags := []command.Flag{boolFlag, float64Flag, float64SliceFlag, intFlag, intSliceFlag, int64Flag, int64SliceFlag, stringFlag, stringSliceFlag}

	// Convert command flags to CLI flags
	cliFlags := flagsToCliFlags(flags)
	assert.NotNil(t, cliFlags)

	// Assert that the number of CLI flags matches the number of command flags
	assert.Equal(t, len(cliFlags), len(flags))

	// Assert that each CLI flag has the expected name, aliases, usage, required, and value
	for i, flag := range flags {
		switch flag.Type() {
		case command.FlagTypeBool:
			boolFlag := flag.(*command.BoolFlag)
			cliBoolFlag := cliFlags[i].(*cli.BoolFlag)

			assert.Equal(t, boolFlag.Name, cliBoolFlag.Name)
			assert.Equal(t, boolFlag.Aliases, cliBoolFlag.Aliases)
			assert.Equal(t, boolFlag.Usage, cliBoolFlag.Usage)
			assert.Equal(t, boolFlag.Required, cliBoolFlag.Required)
			assert.Equal(t, boolFlag.Value, cliBoolFlag.Value)
		case command.FlagTypeFloat64:
			float64Flag := flag.(*command.Float64Flag)
			cliFloat64Flag := cliFlags[i].(*cli.FloatFlag)
			assert.Equal(t, float64Flag.Name, cliFloat64Flag.Name)
			assert.Equal(t, float64Flag.Aliases, cliFloat64Flag.Aliases)
			assert.Equal(t, float64Flag.Usage, cliFloat64Flag.Usage)
			assert.Equal(t, float64Flag.Required, cliFloat64Flag.Required)
			assert.Equal(t, float64Flag.Value, cliFloat64Flag.Value)
		case command.FlagTypeFloat64Slice:
			float64SliceFlag := flag.(*command.Float64SliceFlag)
			cliFloat64SliceFlag := cliFlags[i].(*cli.FloatSliceFlag)
			assert.Equal(t, float64SliceFlag.Name, cliFloat64SliceFlag.Name)
			assert.Equal(t, float64SliceFlag.Aliases, cliFloat64SliceFlag.Aliases)
			assert.Equal(t, float64SliceFlag.Usage, cliFloat64SliceFlag.Usage)
			assert.Equal(t, float64SliceFlag.Required, cliFloat64SliceFlag.Required)
			assert.Equal(t, float64SliceFlag.Value, cliFloat64SliceFlag.Value)
		case command.FlagTypeInt:
			intFlag := flag.(*command.IntFlag)
			cliIntFlag := cliFlags[i].(*cli.IntFlag)
			assert.Equal(t, intFlag.Name, cliIntFlag.Name)
			assert.Equal(t, intFlag.Aliases, cliIntFlag.Aliases)
			assert.Equal(t, intFlag.Usage, cliIntFlag.Usage)
			assert.Equal(t, intFlag.Required, cliIntFlag.Required)
			assert.Equal(t, intFlag.Value, int(cliIntFlag.Value))
		case command.FlagTypeIntSlice:
			intSliceFlag := flag.(*command.IntSliceFlag)
			cliIntSliceFlag := cliFlags[i].(*cli.IntSliceFlag)
			assert.Equal(t, intSliceFlag.Name, cliIntSliceFlag.Name)
			assert.Equal(t, intSliceFlag.Aliases, cliIntSliceFlag.Aliases)
			assert.Equal(t, intSliceFlag.Usage, cliIntSliceFlag.Usage)
			assert.Equal(t, intSliceFlag.Required, cliIntSliceFlag.Required)
			assert.Equal(t, intSliceFlag.Value, cliIntSliceFlag.Value)
		case command.FlagTypeInt64:
			int64Flag := flag.(*command.Int64Flag)
			cliInt64Flag := cliFlags[i].(*cli.Int64Flag)
			assert.Equal(t, int64Flag.Name, cliInt64Flag.Name)
			assert.Equal(t, int64Flag.Aliases, cliInt64Flag.Aliases)
			assert.Equal(t, int64Flag.Usage, cliInt64Flag.Usage)
			assert.Equal(t, int64Flag.Required, cliInt64Flag.Required)
			assert.Equal(t, int64Flag.Value, cliInt64Flag.Value)
		case command.FlagTypeInt64Slice:
			int64SliceFlag := flag.(*command.Int64SliceFlag)
			cliInt64SliceFlag := cliFlags[i].(*cli.Int64SliceFlag)
			assert.Equal(t, int64SliceFlag.Name, cliInt64SliceFlag.Name)
			assert.Equal(t, int64SliceFlag.Aliases, cliInt64SliceFlag.Aliases)
			assert.Equal(t, int64SliceFlag.Usage, cliInt64SliceFlag.Usage)
			assert.Equal(t, int64SliceFlag.Required, cliInt64SliceFlag.Required)
			assert.Equal(t, int64SliceFlag.Value, cliInt64SliceFlag.Value)
		case command.FlagTypeString:
			stringFlag := flag.(*command.StringFlag)
			cliStringFlag := cliFlags[i].(*cli.StringFlag)
			assert.Equal(t, stringFlag.Name, cliStringFlag.Name)
			assert.Equal(t, stringFlag.Aliases, cliStringFlag.Aliases)
			assert.Equal(t, stringFlag.Usage, cliStringFlag.Usage)
			assert.Equal(t, stringFlag.Required, cliStringFlag.Required)
			assert.Equal(t, stringFlag.Value, cliStringFlag.Value)
		case command.FlagTypeStringSlice:
			stringSliceFlag := flag.(*command.StringSliceFlag)
			cliStringSliceFlag := cliFlags[i].(*cli.StringSliceFlag)
			assert.Equal(t, stringSliceFlag.Name, cliStringSliceFlag.Name)
			assert.Equal(t, stringSliceFlag.Aliases, cliStringSliceFlag.Aliases)
			assert.Equal(t, stringSliceFlag.Usage, cliStringSliceFlag.Usage)
			assert.Equal(t, stringSliceFlag.Required, cliStringSliceFlag.Required)
			assert.Equal(t, stringSliceFlag.Value, cliStringSliceFlag.Value)
		}
	}
}

func TestArgumentsToCliArguments(t *testing.T) {
	float32Arg := &command.Float32Argument{Name: "float32Arg", Usage: "float32 argument", Value: float32(1.0)}
	float64Arg := &command.Float64Argument{Name: "float64Arg", Usage: "float64 flag", Value: float64(1.0)}
	intArg := &command.IntArgument{Name: "intArg", Usage: "int argument", Value: 1}
	int8Arg := &command.Int8Argument{Name: "int8Arg", Usage: "int8 argument", Value: int8(1)}
	int16Arg := &command.Int16Argument{Name: "int16Arg", Usage: "int16 argument", Value: int16(1)}
	int32Arg := &command.Int32Argument{Name: "int32Arg", Usage: "int32 argument", Value: int32(1)}
	int64Arg := &command.Int64Argument{Name: "int64Arg", Usage: "int64 argument", Value: int64(1)}
	stringArg := &command.StringArgument{Name: "stringArg", Usage: "string argument", Value: "default"}
	timestampArg := &command.TimestampArgument{Name: "timestampArg", Usage: "timestamp argument", Value: time.Now()}
	uintArg := &command.UintArgument{Name: "uintArg", Usage: "uint argument", Value: uint(1)}
	uint8Arg := &command.Uint8Argument{Name: "uint8Arg", Usage: "uint8 argument", Value: uint8(1)}
	uint16Arg := &command.Uint16Argument{Name: "uint16Arg", Usage: "uint16 argument", Value: uint16(1)}
	uint32Arg := &command.Uint32Argument{Name: "uint32Arg", Usage: "uint32 flag", Value: uint32(1)}
	uint64Arg := &command.Uint64Argument{Name: "uint64Arg", Usage: "uint64 flag", Value: uint64(1)}

	float32SliceArg := &command.Float32SliceArgument{Name: "float32SliceArg", Usage: "float32 slice argument", Value: float32(2.0)}
	float64SliceArg := &command.Float64SliceArgument{Name: "float64SliceArg", Usage: "float64 slice argument", Value: float64(2.0)}
	intSliceArg := &command.IntSliceArgument{Name: "intSliceArg", Usage: "int slice argument", Value: int(2)}
	int8SliceArg := &command.Int8SliceArgument{Name: "int8SliceArg", Usage: "int8 slice argument", Value: int8(2)}
	int16SliceArg := &command.Int16SliceArgument{Name: "int16SliceArg", Usage: "int16 slice argument", Value: int16(2)}
	int32SliceArg := &command.Int32SliceArgument{Name: "int32SliceArg", Usage: "int32 slice argument", Value: int32(2)}
	int64SliceArg := &command.Int64SliceArgument{Name: "int64SliceArg", Usage: "int64 slice argument", Value: int64(2)}
	stringSliceArg := &command.StringSliceArgument{Name: "stringSliceArg", Usage: "string slice argument", Value: "b"}
	timestampSliceArg := &command.TimestampSliceArgument{Name: "timestampSliceArg", Usage: "timestamp slice argument", Value: time.Now().Add(time.Hour)}
	uintSliceArg := &command.UintSliceArgument{Name: "uintSliceArg", Usage: "uint slice argument", Value: uint(2)}
	uint8SliceArg := &command.Uint8SliceArgument{Name: "uint8SliceArg", Usage: "uint8 slice argument", Value: uint8(2)}
	uint16SliceArg := &command.Uint16SliceArgument{Name: "uint16SliceArg", Usage: "uint16 slice argument", Value: uint16(2)}
	uint32SliceArg := &command.Uint32SliceArgument{Name: "uint32SliceArg", Usage: "uint32 slice argument", Value: uint32(2)}
	uint64SliceArg := &command.Uint64SliceArgument{Name: "uint64SliceArg", Usage: "uint64 slice argument", Value: uint64(2)}

	// Create a slice of command flags
	arguments := []command.Argument{
		float32Arg,
		float64Arg,
		intArg,
		int8Arg,
		int16Arg,
		int32Arg,
		int64Arg,
		stringArg,
		timestampArg,
		uintArg,
		uint8Arg,
		uint16Arg,
		uint32Arg,
		uint64Arg,
		float32SliceArg,
		float64SliceArg,
		intSliceArg,
		int8SliceArg,
		int16SliceArg,
		int32SliceArg,
		int64SliceArg,
		stringSliceArg,
		timestampSliceArg,
		uintSliceArg,
		uint8SliceArg,
		uint16SliceArg,
		uint32SliceArg,
		uint64SliceArg,
	}

	// Convert command flags to CLI arguments
	cliArguments := argumentsToCliArgs(arguments)
	assert.NotNil(t, cliArguments)

	// Assert that the number of CLI flags matches the number of command flags
	assert.Equal(t, len(cliArguments), len(arguments))

	// Assert that each CLI argument has the expected name, aliases, usage, required, and value
	for i, v := range arguments {
		switch arg := v.(type) {
		case *command.Float32Argument:
			cliArg := cliArguments[i].(*cli.Float32Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Float64Argument:
			cliArg := cliArguments[i].(*cli.Float64Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.IntArgument:
			cliArg := cliArguments[i].(*cli.IntArgs)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Int8Argument:
			cliArg := cliArguments[i].(*cli.Int8Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Int16Argument:
			cliArg := cliArguments[i].(*cli.Int16Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Int32Argument:
			cliArg := cliArguments[i].(*cli.Int32Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Int64Argument:
			cliArg := cliArguments[i].(*cli.Int64Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.StringArgument:
			cliArg := cliArguments[i].(*cli.StringArgs)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.TimestampArgument:
			cliArg := cliArguments[i].(*cli.TimestampArgs)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value.Unix(), cliArg.Value.Unix())
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.UintArgument:
			cliArg := cliArguments[i].(*cli.UintArgs)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Uint8Argument:
			cliArg := cliArguments[i].(*cli.Uint8Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Uint16Argument:
			cliArg := cliArguments[i].(*cli.Uint16Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Uint32Argument:
			cliArg := cliArguments[i].(*cli.Uint32Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Uint64Argument:
			cliArg := cliArguments[i].(*cli.Uint64Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Float32SliceArgument:
			cliArg := cliArguments[i].(*cli.Float32Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Float64SliceArgument:
			cliArg := cliArguments[i].(*cli.Float64Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.IntSliceArgument:
			cliArg := cliArguments[i].(*cli.IntArgs)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Int8SliceArgument:
			cliArg := cliArguments[i].(*cli.Int8Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Int16SliceArgument:
			cliArg := cliArguments[i].(*cli.Int16Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Int32SliceArgument:
			cliArg := cliArguments[i].(*cli.Int32Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Int64SliceArgument:
			cliArg := cliArguments[i].(*cli.Int64Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.StringSliceArgument:
			cliArg := cliArguments[i].(*cli.StringArgs)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.TimestampSliceArgument:
			cliArg := cliArguments[i].(*cli.TimestampArgs)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value.Unix(), cliArg.Value.Unix())
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.UintSliceArgument:
			cliArg := cliArguments[i].(*cli.UintArgs)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Uint8SliceArgument:
			cliArg := cliArguments[i].(*cli.Uint8Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Uint16SliceArgument:
			cliArg := cliArguments[i].(*cli.Uint16Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Uint32SliceArgument:
			cliArg := cliArguments[i].(*cli.Uint32Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		case *command.Uint64SliceArgument:
			cliArg := cliArguments[i].(*cli.Uint64Args)
			assert.Equal(t, arg.Name, cliArg.Name)
			assert.Equal(t, arg.Usage, cliArg.UsageText)
			assert.Equal(t, arg.Value, cliArg.Value)
			assert.Equal(t, arg.MinOccurrences(), cliArg.Min)
			assert.Equal(t, arg.MaxOccurrences(), cliArg.Max)
		default:
			t.Fatalf("unhandled argument type: %T, value %+v", arg, arg)
		}
	}

}

func TestArgumentsToCliArgumentsPanic(t *testing.T) {
	arguments := []command.Argument{
		&command.StringArgument{
			Name: "string_arg",
		},
		&command.StringArgument{
			Name:     "string_arg_required",
			Required: true,
		},
	}
	assert.Panics(t, func() {
		argumentsToCliArgs(arguments)
	})
}

type TestCommand struct {
}

func (receiver *TestCommand) Signature() string {
	return "test"
}

func (receiver *TestCommand) Description() string {
	return "Test command"
}

func (receiver *TestCommand) Extend() command.Extend {
	return command.Extend{}
}

func (receiver *TestCommand) Handle(ctx console.Context) error {
	testCommand++

	return nil
}
