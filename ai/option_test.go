package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractsai "github.com/goravel/framework/contracts/ai"
)

func TestWithProvider(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]any
		args     []string
		expected map[string]any
		nilMap   bool
	}{
		{
			name:    "sets provider while preserving existing keys",
			initial: map[string]any{"existing-key": "preserve"},
			args:    []string{"openai"},
			expected: map[string]any{
				"existing-key":             "preserve",
				contractsai.OptionProvider: "openai",
			},
		},
		{
			name: "overrides previous value",
			initial: map[string]any{
				contractsai.OptionProvider: "initial-provider",
			},
			args: []string{"openai", "anthropic"},
			expected: map[string]any{
				contractsai.OptionProvider: "anthropic",
			},
		},
		{
			name:   "panics on nil map",
			args:   []string{"openai"},
			nilMap: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.nilMap {
				assert.PanicsWithError(t, "assignment to entry in nil map", func() {
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
		initial  map[string]any
		args     []string
		expected map[string]any
		nilMap   bool
	}{
		{
			name:    "sets model while preserving existing keys",
			initial: map[string]any{"existing-key": "preserve"},
			args:    []string{"gpt-4"},
			expected: map[string]any{
				"existing-key":          "preserve",
				contractsai.OptionModel: "gpt-4",
			},
		},
		{
			name: "overrides previous value",
			initial: map[string]any{
				contractsai.OptionModel: "initial-model",
			},
			args: []string{"gpt-4", "gpt-4o"},
			expected: map[string]any{
				contractsai.OptionModel: "gpt-4o",
			},
		},
		{
			name:   "panics on nil map",
			args:   []string{"gpt-4"},
			nilMap: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.nilMap {
				assert.PanicsWithError(t, "assignment to entry in nil map", func() {
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
