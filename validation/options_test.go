package validation

import (
	"context"
	"testing"

	"github.com/gookit/validate"
	"github.com/stretchr/testify/assert"

	contractsvalidation "github.com/goravel/framework/contracts/validation"
)

func TestRules(t *testing.T) {
	tests := []struct {
		name     string
		rules    map[string]string
		expected map[string]any
	}{
		{
			name:     "with rules",
			rules:    map[string]string{"name": "required", "age": "numeric"},
			expected: map[string]any{"rules": map[string]string{"name": "required", "age": "numeric"}},
		},
		{
			name:     "with empty rules",
			rules:    map[string]string{},
			expected: map[string]any{},
		},
		{
			name:     "with nil rules",
			rules:    nil,
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := make(map[string]any)
			Rules(tt.rules)(options)
			assert.Equal(t, tt.expected, options)
		})
	}
}

func TestFiltersOption(t *testing.T) {
	tests := []struct {
		name     string
		filters  map[string]string
		expected map[string]any
	}{
		{
			name:     "with filters",
			filters:  map[string]string{"name": "trim", "email": "lower"},
			expected: map[string]any{"filters": map[string]string{"name": "trim", "email": "lower"}},
		},
		{
			name:     "with empty filters",
			filters:  map[string]string{},
			expected: map[string]any{},
		},
		{
			name:     "with nil filters",
			filters:  nil,
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := make(map[string]any)
			Filters(tt.filters)(options)
			assert.Equal(t, tt.expected, options)
		})
	}
}

func TestCustomFilters(t *testing.T) {
	mockFilter := &mockFilter{signature: "custom_filter"}

	tests := []struct {
		name     string
		filters  []contractsvalidation.Filter
		expected int
	}{
		{
			name:     "with custom filters",
			filters:  []contractsvalidation.Filter{mockFilter},
			expected: 1,
		},
		{
			name:     "with empty filters",
			filters:  []contractsvalidation.Filter{},
			expected: 0,
		},
		{
			name:     "with nil filters",
			filters:  nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := make(map[string]any)
			CustomFilters(tt.filters)(options)

			if tt.expected > 0 {
				assert.NotNil(t, options["customFilters"])
				assert.Len(t, options["customFilters"], tt.expected)
			} else {
				assert.Nil(t, options["customFilters"])
			}
		})
	}
}

func TestCustomRules(t *testing.T) {
	mockRule := &mockRule{signature: "custom_rule"}

	tests := []struct {
		name     string
		rules    []contractsvalidation.Rule
		expected int
	}{
		{
			name:     "with custom rules",
			rules:    []contractsvalidation.Rule{mockRule},
			expected: 1,
		},
		{
			name:     "with empty rules",
			rules:    []contractsvalidation.Rule{},
			expected: 0,
		},
		{
			name:     "with nil rules",
			rules:    nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := make(map[string]any)
			CustomRules(tt.rules)(options)

			if tt.expected > 0 {
				assert.NotNil(t, options["customRules"])
				assert.Len(t, options["customRules"], tt.expected)
			} else {
				assert.Nil(t, options["customRules"])
			}
		})
	}
}

func TestMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages map[string]string
		expected map[string]any
	}{
		{
			name:     "with messages",
			messages: map[string]string{"required": "Field is required", "email": "Invalid email"},
			expected: map[string]any{"messages": map[string]string{"required": "Field is required", "email": "Invalid email"}},
		},
		{
			name:     "with empty messages",
			messages: map[string]string{},
			expected: map[string]any{},
		},
		{
			name:     "with nil messages",
			messages: nil,
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := make(map[string]any)
			Messages(tt.messages)(options)
			assert.Equal(t, tt.expected, options)
		})
	}
}

func TestAttributes(t *testing.T) {
	tests := []struct {
		name       string
		attributes map[string]string
		expected   map[string]any
	}{
		{
			name:       "with attributes",
			attributes: map[string]string{"name": "Full Name", "email": "Email Address"},
			expected:   map[string]any{"attributes": map[string]string{"name": "Full Name", "email": "Email Address"}},
		},
		{
			name:       "with empty attributes",
			attributes: map[string]string{},
			expected:   map[string]any{},
		},
		{
			name:       "with nil attributes",
			attributes: nil,
			expected:   map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := make(map[string]any)
			Attributes(tt.attributes)(options)
			assert.Equal(t, tt.expected, options)
		})
	}
}

func TestPrepareForValidation(t *testing.T) {
	prepareFunc := func(ctx context.Context, data contractsvalidation.Data) error {
		return nil
	}

	options := make(map[string]any)
	PrepareForValidation(prepareFunc)(options)

	assert.NotNil(t, options["prepareForValidation"])
}

func TestGenerateOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  []contractsvalidation.Option
		expected map[string]any
	}{
		{
			name: "with multiple options",
			options: []contractsvalidation.Option{
				Rules(map[string]string{"name": "required"}),
				Filters(map[string]string{"name": "trim"}),
				Messages(map[string]string{"required": "Field is required"}),
			},
			expected: map[string]any{
				"rules":    map[string]string{"name": "required"},
				"filters":  map[string]string{"name": "trim"},
				"messages": map[string]string{"required": "Field is required"},
			},
		},
		{
			name:     "with no options",
			options:  []contractsvalidation.Option{},
			expected: map[string]any{},
		},
		{
			name: "with empty options",
			options: []contractsvalidation.Option{
				Rules(map[string]string{}),
				Filters(map[string]string{}),
			},
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateOptions(tt.options)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppendOptions(t *testing.T) {
	ctx := context.Background()

	t.Run("append rules", func(t *testing.T) {
		v := validate.Map(map[string]any{"name": "goravel"})
		options := map[string]any{
			"rules": map[string]string{
				"name": "required|minLen:3",
			},
		}

		AppendOptions(ctx, v, options)
		assert.True(t, v.Validate())
	})

	t.Run("append filters", func(t *testing.T) {
		v := validate.Map(map[string]any{"name": "  goravel  "})
		options := map[string]any{
			"filters": map[string]string{
				"name": "trim",
			},
		}

		AppendOptions(ctx, v, options)
		v.StringRule("name", "required")
		assert.True(t, v.Validate())
		assert.Equal(t, "goravel", v.SafeVal("name"))
	})

	t.Run("append messages", func(t *testing.T) {
		v := validate.Map(map[string]any{"name": ""})
		v.StringRule("name", "required")
		options := map[string]any{
			"messages": map[string]string{
				"name.required": "The :attribute field is mandatory",
			},
		}

		AppendOptions(ctx, v, options)
		assert.False(t, v.Validate())
		// The :attribute is replaced with {field} in AppendOptions
		assert.Contains(t, v.Errors.One(), "field is mandatory")
	})

	t.Run("append attributes", func(t *testing.T) {
		v := validate.Map(map[string]any{"name": ""})
		v.StringRule("name", "required")
		options := map[string]any{
			"attributes": map[string]string{
				"name": "Full Name",
			},
		}

		AppendOptions(ctx, v, options)
		assert.False(t, v.Validate())
	})

	t.Run("append custom rules", func(t *testing.T) {
		v := validate.Map(map[string]any{"name": "test"})
		mockRule := &mockRule{
			signature: "custom_rule",
			message:   "The :attribute is invalid",
			passes:    false,
		}
		options := map[string]any{
			"customRules": []contractsvalidation.Rule{mockRule},
		}

		AppendOptions(ctx, v, options)
		v.StringRule("name", "required|custom_rule")
		assert.False(t, v.Validate())
	})

	t.Run("append custom filters", func(t *testing.T) {
		v := validate.Map(map[string]any{"name": "GORAVEL"})
		mockFilter := &mockFilter{
			signature: "custom_lower",
			handle: func(val any) (any, error) {
				if str, ok := val.(string); ok {
					return str + "_filtered", nil
				}
				return val, nil
			},
		}
		options := map[string]any{
			"customFilters": []contractsvalidation.Filter{mockFilter},
		}

		AppendOptions(ctx, v, options)
		v.FilterRule("name", "custom_lower")
		v.StringRule("name", "required")
		assert.True(t, v.Validate())
		assert.Equal(t, "GORAVEL_filtered", v.SafeVal("name"))
	})

	t.Run("append custom filter with nil handle", func(t *testing.T) {
		v := validate.Map(map[string]any{"name": "test"})
		mockFilter := &mockFilter{
			signature: "nil_filter",
			handle:    nil,
		}
		options := map[string]any{
			"customFilters": []contractsvalidation.Filter{mockFilter},
		}

		AppendOptions(ctx, v, options)
		v.StringRule("name", "required")
		assert.True(t, v.Validate())
	})

	t.Run("with empty attributes map", func(t *testing.T) {
		v := validate.Map(map[string]any{"name": "test"})
		options := map[string]any{
			"attributes": map[string]string{},
		}

		AppendOptions(ctx, v, options)
		v.StringRule("name", "required")
		assert.True(t, v.Validate())
	})

	t.Run("with invalid filters type", func(t *testing.T) {
		v := validate.Map(map[string]any{"name": "test"})
		options := map[string]any{
			"filters": "invalid",
		}

		AppendOptions(ctx, v, options)
		v.StringRule("name", "required")
		assert.True(t, v.Validate())
	})

	t.Run("with nil options", func(t *testing.T) {
		v := validate.Map(map[string]any{"name": "test"})
		options := map[string]any{}

		AppendOptions(ctx, v, options)
		v.StringRule("name", "required")
		assert.True(t, v.Validate())
	})
}

// Mock implementations for testing

type mockRule struct {
	signature string
	message   string
	passes    bool
}

func (m *mockRule) Signature() string {
	return m.signature
}

func (m *mockRule) Passes(ctx context.Context, data contractsvalidation.Data, val any, options ...any) bool {
	return m.passes
}

func (m *mockRule) Message(ctx context.Context) string {
	return m.message
}

type mockFilter struct {
	signature string
	handle    any
}

func (m *mockFilter) Signature() string {
	return m.signature
}

func (m *mockFilter) Handle(ctx context.Context) any {
	return m.handle
}
