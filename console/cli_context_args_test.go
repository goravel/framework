package console

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

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
