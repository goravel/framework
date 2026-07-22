package broadcasting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBroadcastJob_Signature(t *testing.T) {
	job := &BroadcastJob{}
	assert.Equal(t, "goravel_broadcast", job.Signature())
}

func TestBroadcastJob_Handle_InvalidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []any
	}{
		{
			name: "no arguments",
			args: nil,
		},
		{
			name: "empty arguments",
			args: []any{},
		},
		{
			name: "non-string argument",
			args: []any{42},
		},
		{
			name: "wrong type argument",
			args: []any{map[string]any{"key": "value"}},
		},
		{
			name: "invalid JSON",
			args: []any{"not-json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &BroadcastJob{}
			err := job.Handle(tt.args...)
			assert.Error(t, err)
		})
	}
}
