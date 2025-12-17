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
		want log.Value
	}{
		{
			name: "bool true",
			arg:  true,
			want: log.BoolValue(true),
		},
		{
			name: "string",
			arg:  "goravel",
			want: log.StringValue("goravel"),
		},
		{
			name: "int",
			arg:  int(42),
			want: log.Int64Value(42),
		},
		{
			name: "int64",
			arg:  int64(9000),
			want: log.Int64Value(9000),
		},
		{
			name: "float64",
			arg:  3.14159,
			want: log.Float64Value(3.14159),
		},
		{
			name: "time.Time (RFC3339Nano)",
			arg:  fixedTime,
			want: log.StringValue(fixedTimeStr),
		},
		{
			name: "[]byte",
			arg:  []byte("secret"),
			want: log.BytesValue([]byte("secret")),
		},
		{
			name: "error",
			arg:  errors.New("database connection failed"),
			want: log.StringValue("database connection failed"),
		},
		{
			name: "fmt.Stringer",
			arg:  attribute.Key("custom_key"),
			want: log.StringValue("custom_key"),
		},
		{
			name: "nil interface",
			arg:  nil,
			want: log.Value{},
		},
		{
			name: "nil pointer (typed)",
			arg:  (*string)(nil),
			want: log.Value{},
		},
		{
			name: "map[string]any (Structured Log)",
			arg: map[string]any{
				"role": "admin",
			},
			want: log.MapValue(
				log.String("role", "admin"),
			),
		},
		{
			name: "map[string]int (Typed Map)",
			arg: map[string]int{
				"retries": 3,
			},
			want: log.MapValue(
				log.Int64("retries", 3),
			),
		},
		{
			name: "[]string (Tags)",
			arg:  []string{"api", "v1"},
			want: log.SliceValue(
				log.StringValue("api"),
				log.StringValue("v1"),
			),
		},
		{
			name: "[]int",
			arg:  []int{1, 2, 3},
			want: log.SliceValue(
				log.Int64Value(1),
				log.Int64Value(2),
				log.Int64Value(3),
			),
		},
		{
			name: "complex64",
			arg:  complex(float32(1.5), float32(2.5)),
			want: log.MapValue(
				log.Float64("r", 1.5),
				log.Float64("i", 2.5),
			),
		},
		{
			name: "struct (simple)",
			arg: struct {
				ID   int
				Name string
			}{1, "User"},
			want: log.StringValue("{ID:1 Name:User}"),
		},
		{
			name: "pointer to struct",
			arg: &struct {
				Active bool
			}{true},
			want: log.StringValue("{Active:true}"),
		},
		{
			name: "context.Context",
			arg:  context.Background(),
			want: log.StringValue("context.Background"),
		},
		{
			name: "log.Value Pass-through",
			arg:  log.BoolValue(false),
			want: log.BoolValue(false),
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
	assert.Equal(t, log.Int64Value(100), toValue(val))

	hugeVal := uint64(18446744073709551615)
	assert.Equal(t, log.StringValue("18446744073709551615"), toValue(hugeVal))
}
