package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

var testCommand = 0

func TestRun(t *testing.T) {
	cli := NewApplication()
	cli.Register([]console.Command{
		&TestCommand{},
	})

	cli.Call("test")
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
			cliFloat64Flag := cliFlags[i].(*cli.Float64Flag)
			assert.Equal(t, float64Flag.Name, cliFloat64Flag.Name)
			assert.Equal(t, float64Flag.Aliases, cliFloat64Flag.Aliases)
			assert.Equal(t, float64Flag.Usage, cliFloat64Flag.Usage)
			assert.Equal(t, float64Flag.Required, cliFloat64Flag.Required)
			assert.Equal(t, float64Flag.Value, cliFloat64Flag.Value)
		case command.FlagTypeFloat64Slice:
			float64SliceFlag := flag.(*command.Float64SliceFlag)
			cliFloat64SliceFlag := cliFlags[i].(*cli.Float64SliceFlag)
			assert.Equal(t, float64SliceFlag.Name, cliFloat64SliceFlag.Name)
			assert.Equal(t, float64SliceFlag.Aliases, cliFloat64SliceFlag.Aliases)
			assert.Equal(t, float64SliceFlag.Usage, cliFloat64SliceFlag.Usage)
			assert.Equal(t, float64SliceFlag.Required, cliFloat64SliceFlag.Required)
			assert.Equal(t, cli.NewFloat64Slice(float64SliceFlag.Value...), cliFloat64SliceFlag.Value)
		case command.FlagTypeInt:
			intFlag := flag.(*command.IntFlag)
			cliIntFlag := cliFlags[i].(*cli.IntFlag)
			assert.Equal(t, intFlag.Name, cliIntFlag.Name)
			assert.Equal(t, intFlag.Aliases, cliIntFlag.Aliases)
			assert.Equal(t, intFlag.Usage, cliIntFlag.Usage)
			assert.Equal(t, intFlag.Required, cliIntFlag.Required)
			assert.Equal(t, intFlag.Value, cliIntFlag.Value)
		case command.FlagTypeIntSlice:
			intSliceFlag := flag.(*command.IntSliceFlag)
			cliIntSliceFlag := cliFlags[i].(*cli.IntSliceFlag)
			assert.Equal(t, intSliceFlag.Name, cliIntSliceFlag.Name)
			assert.Equal(t, intSliceFlag.Aliases, cliIntSliceFlag.Aliases)
			assert.Equal(t, intSliceFlag.Usage, cliIntSliceFlag.Usage)
			assert.Equal(t, intSliceFlag.Required, cliIntSliceFlag.Required)
			assert.Equal(t, cli.NewIntSlice(intSliceFlag.Value...), cliIntSliceFlag.Value)
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
			assert.Equal(t, cli.NewInt64Slice(int64SliceFlag.Value...), cliInt64SliceFlag.Value)
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
			assert.Equal(t, cli.NewStringSlice(stringSliceFlag.Value...), cliStringSliceFlag.Value)
		}
	}
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
