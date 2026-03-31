package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractsai "github.com/goravel/framework/contracts/ai"
)

func TestWithProvider(t *testing.T) {
	tests := []struct {
		name     string
		initial  *contractsai.Options
		args     []string
		expected *contractsai.Options
		nilOpts  bool
	}{
		{
			name:    "sets provider while preserving model",
			initial: &contractsai.Options{Model: "gpt-4"},
			args:    []string{"openai"},
			expected: &contractsai.Options{Provider: "openai", Model: "gpt-4"},
		},
		{
			name: "overrides previous value",
			initial: &contractsai.Options{Provider: "initial-provider"},
			args: []string{"openai", "anthropic"},
			expected: &contractsai.Options{Provider: "anthropic"},
		},
		{
			name:    "panics on nil options",
			args:    []string{"openai"},
			nilOpts: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.nilOpts {
				assert.Panics(t, func() {
					for _, arg := range tt.args {
						WithProvider(arg)(nil)
					}
				})
				return
			}
			for _, arg := range tt.args {
				WithProvider(arg)(tt.initial)
			}
			assert.Equal(t, tt.expected, tt.initial)
		})
	}
}

func TestWithModel(t *testing.T) {
	tests := []struct {
		name     string
		initial  *contractsai.Options
		args     []string
		expected *contractsai.Options
		nilOpts  bool
	}{
		{
			name:    "sets model while preserving provider",
			initial: &contractsai.Options{Provider: "openai"},
			args:    []string{"gpt-4"},
			expected: &contractsai.Options{Provider: "openai", Model: "gpt-4"},
		},
		{
			name: "overrides previous value",
			initial: &contractsai.Options{Model: "initial-model"},
			args: []string{"gpt-4", "gpt-4o"},
			expected: &contractsai.Options{Model: "gpt-4o"},
		},
		{
			name:    "panics on nil options",
			args:    []string{"gpt-4"},
			nilOpts: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.nilOpts {
				assert.Panics(t, func() {
					for _, arg := range tt.args {
						WithModel(arg)(nil)
					}
				})
				return
			}
			for _, arg := range tt.args {
				WithModel(arg)(tt.initial)
			}
			assert.Equal(t, tt.expected, tt.initial)
		})
	}
}
