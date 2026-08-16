package log

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"

	contractslog "github.com/goravel/framework/contracts/log"
)

func TestToSeverity(t *testing.T) {
	var UnknowLevel contractslog.Level = 45
	tests := []struct {
		name  string
		level contractslog.Level
		want  log.Severity
	}{
		{"Debug", contractslog.LevelDebug, log.SeverityDebug},
		{"Info", contractslog.LevelInfo, log.SeverityInfo},
		{"Warning", contractslog.LevelWarning, log.SeverityWarn},
		{"Error", contractslog.LevelError, log.SeverityError},
		{"Fatal", contractslog.LevelFatal, log.SeverityFatal},
		{"Panic", contractslog.LevelPanic, log.SeverityFatal4},
		{"Unknown", UnknowLevel, log.SeverityInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, toSeverity(tt.level))
		})
	}
}

func TestToValue(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 45, 123456000, time.UTC)
	fixedTimeStr := "2024-01-15T10:30:45.123456Z"

	tests := []struct {
		name string
		arg  any
		want attribute.Value
	}{
		{
			name: "bool true",
			arg:  true,
			want: attribute.BoolValue(true),
		},
		{
			name: "string",
			arg:  "goravel",
			want: attribute.StringValue("goravel"),
		},
		{
			name: "int",
			arg:  int(42),
			want: attribute.Int64Value(42),
		},
		{
			name: "int64",
			arg:  int64(9000),
			want: attribute.Int64Value(9000),
		},
		{
			name: "float64",
			arg:  3.14159,
			want: attribute.Float64Value(3.14159),
		},
		{
			name: "time.Time (RFC3339Nano)",
			arg:  fixedTime,
			want: attribute.StringValue(fixedTimeStr),
		},
		{
			name: "[]byte",
			arg:  []byte("secret"),
			want: attribute.ByteSliceValue([]byte("secret")),
		},
		{
			name: "error",
			arg:  errors.New("database connection failed"),
			want: attribute.StringValue("database connection failed"),
		},
		{
			name: "fmt.Stringer",
			arg:  attribute.Key("custom_key"),
			want: attribute.StringValue("custom_key"),
		},
		{
			name: "nil interface",
			arg:  nil,
			want: attribute.Value{},
		},
		{
			name: "nil pointer (typed)",
			arg:  (*string)(nil),
			want: attribute.Value{},
		},
		{
			name: "map[string]any (Structured Log)",
			arg: map[string]any{
				"role": "admin",
			},
			want: attribute.MapValue(
				attribute.String("role", "admin"),
			),
		},
		{
			name: "map[string]int (Typed Map)",
			arg: map[string]int{
				"retries": 3,
			},
			want: attribute.MapValue(
				attribute.Int64("retries", 3),
			),
		},
		{
			name: "[]string (Tags)",
			arg:  []string{"api", "v1"},
			want: attribute.SliceValue(
				attribute.StringValue("api"),
				attribute.StringValue("v1"),
			),
		},
		{
			name: "[]int",
			arg:  []int{1, 2, 3},
			want: attribute.SliceValue(
				attribute.Int64Value(1),
				attribute.Int64Value(2),
				attribute.Int64Value(3),
			),
		},
		{
			name: "complex64",
			arg:  complex(float32(1.5), float32(2.5)),
			want: attribute.MapValue(
				attribute.Float64("r", 1.5),
				attribute.Float64("i", 2.5),
			),
		},
		{
			name: "struct (simple)",
			arg: struct {
				ID   int
				Name string
			}{1, "User"},
			want: attribute.StringValue("{ID:1 Name:User}"),
		},
		{
			name: "pointer to struct",
			arg: &struct {
				Active bool
			}{true},
			want: attribute.StringValue("{Active:true}"),
		},
		{
			name: "context.Context",
			arg:  context.Background(),
			want: attribute.StringValue("context.Background"),
		},
		{
			name: "attribute.Value Pass-through",
			arg:  attribute.BoolValue(false),
			want: attribute.BoolValue(false),
		},
		{
			name: "attribute.Value",
			arg:  attribute.StringValue("from_attribute"),
			want: attribute.StringValue("from_attribute"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toValue(tt.arg)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToValue_Uint_Overflow(t *testing.T) {
	val := uint64(100)
	assert.Equal(t, attribute.Int64Value(100), toValue(val))

	hugeVal := uint64(18446744073709551615)
	assert.Equal(t, attribute.StringValue("18446744073709551615"), toValue(hugeVal))
}
