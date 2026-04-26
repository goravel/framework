package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	contractsai "github.com/goravel/framework/contracts/ai"
	contractshttp "github.com/goravel/framework/contracts/http"
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
			name:     "sets provider while preserving model",
			initial:  &contractsai.Options{Model: "gpt-4"},
			args:     []string{"openai"},
			expected: &contractsai.Options{Provider: "openai", Model: "gpt-4"},
		},
		{
			name:     "overrides previous value",
			initial:  &contractsai.Options{Provider: "initial-provider"},
			args:     []string{"openai", "anthropic"},
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
			name:     "sets model while preserving provider",
			initial:  &contractsai.Options{Provider: "openai"},
			args:     []string{"gpt-4"},
			expected: &contractsai.Options{Provider: "openai", Model: "gpt-4"},
		},
		{
			name:     "overrides previous value",
			initial:  &contractsai.Options{Model: "initial-model"},
			args:     []string{"gpt-4", "gpt-4o"},
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

func TestWithMiddleware(t *testing.T) {
	middlewareA := &optionTestMiddleware{}
	middlewareB := &optionTestMiddleware{}

	tests := []struct {
		name     string
		initial  *contractsai.Options
		apply    func(*contractsai.Options)
		expected *contractsai.Options
		nilOpts  bool
	}{
		{
			name:    "appends middleware while preserving options",
			initial: &contractsai.Options{Provider: "openai", Model: "gpt-4"},
			apply: func(options *contractsai.Options) {
				WithMiddleware(middlewareA, middlewareB)(options)
			},
			expected: &contractsai.Options{
				Provider:    "openai",
				Model:       "gpt-4",
				Middlewares: []contractsai.Middleware{middlewareA, middlewareB},
			},
		},
		{
			name: "appends to existing middleware",
			initial: &contractsai.Options{
				Middlewares: []contractsai.Middleware{middlewareA},
			},
			apply: func(options *contractsai.Options) {
				WithMiddleware(middlewareB)(options)
			},
			expected: &contractsai.Options{Middlewares: []contractsai.Middleware{middlewareA, middlewareB}},
		},
		{
			name:    "skips typed nil middleware",
			initial: &contractsai.Options{},
			apply: func(options *contractsai.Options) {
				var middleware *optionNilTestMiddleware
				WithMiddleware(middleware, middlewareA)(options)
			},
			expected: &contractsai.Options{Middlewares: []contractsai.Middleware{middlewareA}},
		},
		{
			name:    "panics on nil options",
			nilOpts: true,
			apply: func(options *contractsai.Options) {
				WithMiddleware(middlewareA)(options)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.nilOpts {
				assert.Panics(t, func() {
					tt.apply(nil)
				})
				return
			}

			tt.apply(tt.initial)
			assert.Equal(t, tt.expected, tt.initial)
		})
	}
}

func TestWithStreamCode(t *testing.T) {
	tests := []struct {
		name     string
		initial  *contractsai.StreamOptions
		args     []int
		expected *contractsai.StreamOptions
		nilOpts  bool
	}{
		{
			name:    "sets stream code",
			initial: &contractsai.StreamOptions{},
			args:    []int{204},
			expected: &contractsai.StreamOptions{
				Code: 204,
			},
		},
		{
			name:    "overrides previous stream code",
			initial: &contractsai.StreamOptions{Code: 200},
			args:    []int{201, 202},
			expected: &contractsai.StreamOptions{
				Code: 202,
			},
		},
		{
			name:    "panics on nil options",
			args:    []int{200},
			nilOpts: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.nilOpts {
				assert.Panics(t, func() {
					for _, arg := range tt.args {
						WithStreamCode(arg)(nil)
					}
				})
				return
			}

			for _, arg := range tt.args {
				WithStreamCode(arg)(tt.initial)
			}
			assert.Equal(t, tt.expected, tt.initial)
		})
	}
}

func TestWithStreamRender(t *testing.T) {
	tests := []struct {
		name        string
		nilOpts     bool
		render      contractsai.RenderFunc
		expectError error
	}{
		{
			name: "sets stream render",
			render: func(w contractshttp.StreamWriter, event contractsai.StreamEvent) error {
				return nil
			},
		},
		{
			name: "overrides previous stream render",
			render: func(w contractshttp.StreamWriter, event contractsai.StreamEvent) error {
				return assert.AnError
			},
			expectError: assert.AnError,
		},
		{
			name:    "panics on nil options",
			nilOpts: true,
			render: func(w contractshttp.StreamWriter, event contractsai.StreamEvent) error {
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.nilOpts {
				assert.Panics(t, func() {
					WithStreamRender(tt.render)(nil)
				})
				return
			}

			options := &contractsai.StreamOptions{
				Render: func(w contractshttp.StreamWriter, event contractsai.StreamEvent) error { return nil },
			}

			WithStreamRender(tt.render)(options)
			err := options.Render(nil, contractsai.StreamEvent{Type: contractsai.StreamEventTypeDone})

			assert.Equal(t, tt.expectError, err)
		})
	}
}

type optionTestMiddleware struct{}

func (m *optionTestMiddleware) Handle(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
	return next(ctx, prompt)
}

type optionNilTestMiddleware struct{}

func (m *optionNilTestMiddleware) Handle(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.Response, error) {
	return next(ctx, prompt)
}
