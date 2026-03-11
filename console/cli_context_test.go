package console

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
)

func TestArgumentString(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue string
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentString{Name: "string", Default: "default"}},
			args:          []string{"test"},
			expectedValue: "default",
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentString{Name: "string", Required: true}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg string not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentString{Name: "string", Required: true}},
			args:          []string{"test", "hello"},
			expectedValue: "hello",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentString("string"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentStringSlice(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue []string
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentStringSlice{Name: "string", Default: []string{"default"}, Min: 0, Max: 2}},
			args:          []string{"test"},
			expectedValue: []string{"default"},
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentStringSlice{Name: "string", Default: []string{"default"}, Min: 1, Max: 2}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg string not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentStringSlice{Name: "string", Default: []string{"default"}, Min: 1, Max: 2}},
			args:          []string{"test", "hello", "world"},
			expectedValue: []string{"hello", "world"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentStringSlice("string"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentInt(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue int
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentInt{Name: "int", Default: 42}},
			args:          []string{"test"},
			expectedValue: 42,
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentInt{Name: "int", Required: true}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg int not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentInt{Name: "int", Required: true}},
			args:          []string{"test", "100"},
			expectedValue: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentInt("int"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentIntSlice(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue []int
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentIntSlice{Name: "int", Default: []int{1, 2}, Min: 0, Max: 3}},
			args:          []string{"test"},
			expectedValue: []int{1, 2},
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentIntSlice{Name: "int", Default: []int{1}, Min: 1, Max: 3}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg int not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentIntSlice{Name: "int", Default: []int{1}, Min: 1, Max: 3}},
			args:          []string{"test", "10", "20", "30"},
			expectedValue: []int{10, 20, 30},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentIntSlice("int"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentFloat64(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue float64
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentFloat64{Name: "float", Default: 3.14}},
			args:          []string{"test"},
			expectedValue: 3.14,
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentFloat64{Name: "float", Required: true}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg float not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentFloat64{Name: "float", Required: true}},
			args:          []string{"test", "2.718"},
			expectedValue: 2.718,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentFloat64("float"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentFloat64Slice(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue []float64
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentFloat64Slice{Name: "float", Default: []float64{1.1, 2.2}, Min: 0, Max: 3}},
			args:          []string{"test"},
			expectedValue: []float64{1.1, 2.2},
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentFloat64Slice{Name: "float", Default: []float64{1.1}, Min: 1, Max: 3}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg float not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentFloat64Slice{Name: "float", Default: []float64{1.1}, Min: 1, Max: 3}},
			args:          []string{"test", "3.14", "2.718", "1.414"},
			expectedValue: []float64{3.14, 2.718, 1.414},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentFloat64Slice("float"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentUint(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue uint
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentUint{Name: "uint", Default: 42}},
			args:          []string{"test"},
			expectedValue: 42,
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentUint{Name: "uint", Required: true}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg uint not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentUint{Name: "uint", Required: true}},
			args:          []string{"test", "100"},
			expectedValue: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentUint("uint"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentUintSlice(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue []uint
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentUintSlice{Name: "uint", Default: []uint{1, 2}, Min: 0, Max: 3}},
			args:          []string{"test"},
			expectedValue: []uint{1, 2},
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentUintSlice{Name: "uint", Default: []uint{1}, Min: 1, Max: 3}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg uint not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentUintSlice{Name: "uint", Default: []uint{1}, Min: 1, Max: 3}},
			args:          []string{"test", "10", "20", "30"},
			expectedValue: []uint{10, 20, 30},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentUintSlice("uint"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentInt8(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue int8
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentInt8{Name: "int8", Default: 42}},
			args:          []string{"test"},
			expectedValue: 42,
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentInt8{Name: "int8", Required: true}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg int8 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentInt8{Name: "int8", Required: true}},
			args:          []string{"test", "127"},
			expectedValue: 127,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentInt8("int8"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentInt8Slice(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue []int8
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentInt8Slice{Name: "int8", Default: []int8{1, 2}, Min: 0, Max: 3}},
			args:          []string{"test"},
			expectedValue: []int8{1, 2},
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentInt8Slice{Name: "int8", Default: []int8{1}, Min: 1, Max: 3}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg int8 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentInt8Slice{Name: "int8", Default: []int8{1}, Min: 1, Max: 3}},
			args:          []string{"test", "10", "20", "30"},
			expectedValue: []int8{10, 20, 30},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentInt8Slice("int8"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentInt16(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue int16
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentInt16{Name: "int16", Default: 1000}},
			args:          []string{"test"},
			expectedValue: 1000,
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentInt16{Name: "int16", Required: true}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg int16 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentInt16{Name: "int16", Required: true}},
			args:          []string{"test", "32767"},
			expectedValue: 32767,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentInt16("int16"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentInt16Slice(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue []int16
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentInt16Slice{Name: "int16", Default: []int16{100, 200}, Min: 0, Max: 3}},
			args:          []string{"test"},
			expectedValue: []int16{100, 200},
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentInt16Slice{Name: "int16", Default: []int16{100}, Min: 1, Max: 3}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg int16 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentInt16Slice{Name: "int16", Default: []int16{100}, Min: 1, Max: 3}},
			args:          []string{"test", "1000", "2000", "3000"},
			expectedValue: []int16{1000, 2000, 3000},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentInt16Slice("int16"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentInt32(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue int32
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentInt32{Name: "int32", Default: 100000}},
			args:          []string{"test"},
			expectedValue: 100000,
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentInt32{Name: "int32", Required: true}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg int32 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentInt32{Name: "int32", Required: true}},
			args:          []string{"test", "2147483647"},
			expectedValue: 2147483647,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentInt32("int32"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentInt32Slice(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue []int32
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentInt32Slice{Name: "int32", Default: []int32{1000, 2000}, Min: 0, Max: 3}},
			args:          []string{"test"},
			expectedValue: []int32{1000, 2000},
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentInt32Slice{Name: "int32", Default: []int32{1000}, Min: 1, Max: 3}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg int32 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentInt32Slice{Name: "int32", Default: []int32{1000}, Min: 1, Max: 3}},
			args:          []string{"test", "10000", "20000", "30000"},
			expectedValue: []int32{10000, 20000, 30000},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentInt32Slice("int32"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentInt64(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue int64
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentInt64{Name: "int64", Default: 1000000}},
			args:          []string{"test"},
			expectedValue: 1000000,
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentInt64{Name: "int64", Required: true}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg int64 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentInt64{Name: "int64", Required: true}},
			args:          []string{"test", "9223372036854775807"},
			expectedValue: 9223372036854775807,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentInt64("int64"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentInt64Slice(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue []int64
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentInt64Slice{Name: "int64", Default: []int64{10000, 20000}, Min: 0, Max: 3}},
			args:          []string{"test"},
			expectedValue: []int64{10000, 20000},
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentInt64Slice{Name: "int64", Default: []int64{10000}, Min: 1, Max: 3}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg int64 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentInt64Slice{Name: "int64", Default: []int64{10000}, Min: 1, Max: 3}},
			args:          []string{"test", "100000", "200000", "300000"},
			expectedValue: []int64{100000, 200000, 300000},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentInt64Slice("int64"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentFloat32(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue float32
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentFloat32{Name: "float32", Default: 3.14}},
			args:          []string{"test"},
			expectedValue: 3.14,
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentFloat32{Name: "float32", Required: true}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg float32 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentFloat32{Name: "float32", Required: true}},
			args:          []string{"test", "2.718"},
			expectedValue: 2.718,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentFloat32("float32"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentFloat32Slice(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue []float32
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentFloat32Slice{Name: "float32", Default: []float32{1.1, 2.2}, Min: 0, Max: 3}},
			args:          []string{"test"},
			expectedValue: []float32{1.1, 2.2},
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentFloat32Slice{Name: "float32", Default: []float32{1.1}, Min: 1, Max: 3}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg float32 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentFloat32Slice{Name: "float32", Default: []float32{1.1}, Min: 1, Max: 3}},
			args:          []string{"test", "3.14", "2.718", "1.414"},
			expectedValue: []float32{3.14, 2.718, 1.414},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentFloat32Slice("float32"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentUint8(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue uint8
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentUint8{Name: "uint8", Default: 100}},
			args:          []string{"test"},
			expectedValue: 100,
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentUint8{Name: "uint8", Required: true}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg uint8 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentUint8{Name: "uint8", Required: true}},
			args:          []string{"test", "255"},
			expectedValue: 255,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentUint8("uint8"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentUint8Slice(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue []uint8
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentUint8Slice{Name: "uint8", Default: []uint8{10, 20}, Min: 0, Max: 3}},
			args:          []string{"test"},
			expectedValue: []uint8{10, 20},
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentUint8Slice{Name: "uint8", Default: []uint8{10}, Min: 1, Max: 3}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg uint8 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentUint8Slice{Name: "uint8", Default: []uint8{10}, Min: 1, Max: 3}},
			args:          []string{"test", "100", "200", "255"},
			expectedValue: []uint8{100, 200, 255},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentUint8Slice("uint8"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentUint16(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue uint16
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentUint16{Name: "uint16", Default: 1000}},
			args:          []string{"test"},
			expectedValue: 1000,
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentUint16{Name: "uint16", Required: true}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg uint16 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentUint16{Name: "uint16", Required: true}},
			args:          []string{"test", "65535"},
			expectedValue: 65535,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentUint16("uint16"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentUint16Slice(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue []uint16
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentUint16Slice{Name: "uint16", Default: []uint16{100, 200}, Min: 0, Max: 3}},
			args:          []string{"test"},
			expectedValue: []uint16{100, 200},
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentUint16Slice{Name: "uint16", Default: []uint16{100}, Min: 1, Max: 3}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg uint16 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentUint16Slice{Name: "uint16", Default: []uint16{100}, Min: 1, Max: 3}},
			args:          []string{"test", "1000", "2000", "3000"},
			expectedValue: []uint16{1000, 2000, 3000},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentUint16Slice("uint16"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentUint32(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue uint32
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentUint32{Name: "uint32", Default: 100000}},
			args:          []string{"test"},
			expectedValue: 100000,
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentUint32{Name: "uint32", Required: true}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg uint32 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentUint32{Name: "uint32", Required: true}},
			args:          []string{"test", "4294967295"},
			expectedValue: 4294967295,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentUint32("uint32"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentUint32Slice(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue []uint32
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentUint32Slice{Name: "uint32", Default: []uint32{1000, 2000}, Min: 0, Max: 3}},
			args:          []string{"test"},
			expectedValue: []uint32{1000, 2000},
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentUint32Slice{Name: "uint32", Default: []uint32{1000}, Min: 1, Max: 3}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg uint32 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentUint32Slice{Name: "uint32", Default: []uint32{1000}, Min: 1, Max: 3}},
			args:          []string{"test", "10000", "20000", "30000"},
			expectedValue: []uint32{10000, 20000, 30000},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentUint32Slice("uint32"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentUint64(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue uint64
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentUint64{Name: "uint64", Default: 1000000}},
			args:          []string{"test"},
			expectedValue: 1000000,
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentUint64{Name: "uint64", Required: true}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg uint64 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentUint64{Name: "uint64", Required: true}},
			args:          []string{"test", "18446744073709551615"},
			expectedValue: 18446744073709551615,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentUint64("uint64"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentUint64Slice(t *testing.T) {
	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue []uint64
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentUint64Slice{Name: "uint64", Default: []uint64{10000, 20000}, Min: 0, Max: 3}},
			args:          []string{"test"},
			expectedValue: []uint64{10000, 20000},
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentUint64Slice{Name: "uint64", Default: []uint64{10000}, Min: 1, Max: 3}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg uint64 not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentUint64Slice{Name: "uint64", Default: []uint64{10000}, Min: 1, Max: 3}},
			args:          []string{"test", "100000", "200000", "300000"},
			expectedValue: []uint64{100000, 200000, 300000},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					assert.Equal(t, tc.expectedValue, cliCtx.ArgumentUint64Slice("uint64"))
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentTimestamp(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue time.Time
		expectedError string
	}{
		{
			name:          "get default value",
			arguments:     []command.Argument{&command.ArgumentTimestamp{Name: "timestamp", Value: now, Layouts: []string{time.RFC3339}}},
			args:          []string{"test"},
			expectedValue: now,
		},
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentTimestamp{Name: "timestamp", Required: true, Layouts: []string{time.RFC3339}}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg timestamp not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentTimestamp{Name: "timestamp", Required: true, Layouts: []string{time.RFC3339}}},
			args:          []string{"test", now.Format(time.RFC3339)},
			expectedValue: now,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					result := cliCtx.ArgumentTimestamp("timestamp")
					assert.Equal(t, tc.expectedValue.Unix(), result.Unix())
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgumentTimestampSlice(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)

	testCases := []struct {
		name          string
		arguments     []command.Argument
		args          []string
		expectedValue []time.Time
		expectedError string
	}{
		{
			name:          "failed when require",
			arguments:     []command.Argument{&command.ArgumentTimestampSlice{Name: "timestamp", Value: now, Min: 1, Max: 3, Layouts: []string{time.RFC3339}}},
			args:          []string{"test"},
			expectedError: "sufficient count of arg timestamp not provided, given 0 expected 1",
		},
		{
			name:          "get real value",
			arguments:     []command.Argument{&command.ArgumentTimestampSlice{Name: "timestamp", Value: now, Min: 1, Max: 3, Layouts: []string{time.RFC3339}}},
			args:          []string{"test", now.Format(time.RFC3339), later.Format(time.RFC3339)},
			expectedValue: []time.Time{now, later},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cliArguments, err := argumentsToCliArgs(tc.arguments)
			require.NoError(t, err)

			cmd := &cli.Command{
				Name:      "test",
				Arguments: cliArguments,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cliCtx := NewCliContext(cmd, tc.arguments)
					result := cliCtx.ArgumentTimestampSlice("timestamp")
					if tc.expectedValue != nil {
						require.Equal(t, len(tc.expectedValue), len(result))
						for i := range tc.expectedValue {
							assert.Equal(t, tc.expectedValue[i].Unix(), result[i].Unix())
						}
					}
					return nil
				},
			}

			err = cmd.Run(context.Background(), tc.args)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

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
					cliCtx := NewCliContext(cmd, nil)
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
		termWidth      int
		expectedOutput string
	}{
		{
			name: "test Divider default",
			testFunc: func(ctx *CliContext) {
				ctx.Divider()
			},
			termWidth:      20,
			expectedOutput: color.Default().Sprintln("--------------------"),
		},
		{
			name: "test Divider empty",
			testFunc: func(ctx *CliContext) {
				ctx.Divider("")
			},
			termWidth:      20,
			expectedOutput: color.Default().Sprintln("--------------------"),
		},
		{
			name: "test Divider char",
			testFunc: func(ctx *CliContext) {
				ctx.Divider("=")
			},
			termWidth:      20,
			expectedOutput: color.Default().Sprintln("===================="),
		},
		{
			name: "test Divider multiple",
			testFunc: func(ctx *CliContext) {
				ctx.Divider("=->")
			},
			termWidth:      20,
			expectedOutput: color.Default().Sprintln("=->=->=->=->=->=->=-"),
		},
		{
			name: "test Divider multibyte",
			testFunc: func(ctx *CliContext) {
				ctx.Divider("♠")
			},
			termWidth:      20,
			expectedOutput: color.Default().Sprintln("♠♠♠♠♠♠♠♠♠♠♠♠♠♠♠♠♠♠♠♠"),
		},
		{
			name: "test Divider multibyte multiple",
			testFunc: func(ctx *CliContext) {
				ctx.Divider("♠♣♥")
			},
			termWidth:      20,
			expectedOutput: color.Default().Sprintln("♠♣♥♠♣♥♠♣♥♠♣♥♠♣♥♠♣♥♠♣"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := CliContext{}
			got := color.CaptureOutput(func(io.Writer) {
				pterm.SetForcedTerminalSize(tt.termWidth, 10)
				tt.testFunc(&ctx)
			})

			assert.Equal(t, tt.expectedOutput, got)
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

			assert.Equal(t, tt.expectedOutput, got)
		})
	}
}
