package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRules(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []ParsedRule
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:  "single rule",
			input: "required",
			expected: []ParsedRule{
				{Name: "required"},
			},
		},
		{
			name:  "multiple rules",
			input: "required|string|max:255",
			expected: []ParsedRule{
				{Name: "required"},
				{Name: "string"},
				{Name: "max", Parameters: []string{"255"}},
			},
		},
		{
			name:  "rule with multiple parameters",
			input: "in:a,b,c",
			expected: []ParsedRule{
				{Name: "in", Parameters: []string{"a", "b", "c"}},
			},
		},
		{
			name:  "regex at end",
			input: `required|regex:^\S+@\S+\.\S+$`,
			expected: []ParsedRule{
				{Name: "required"},
				{Name: "regex", Parameters: []string{`^\S+@\S+\.\S+$`}},
			},
		},
		{
			name:  "regex with pipe consumes rest",
			input: "required|regex:^(foo|bar)$",
			expected: []ParsedRule{
				{Name: "required"},
				{Name: "regex", Parameters: []string{"^(foo|bar)$"}},
			},
		},
		{
			name:  "regex with pipe and trailing rule is consumed",
			input: "required|regex:^(foo|bar)$|string",
			expected: []ParsedRule{
				{Name: "required"},
				{Name: "regex", Parameters: []string{"^(foo|bar)$|string"}},
			},
		},
		{
			name:  "not_regex with pipe",
			input: "not_regex:^(bad|worse)$",
			expected: []ParsedRule{
				{Name: "not_regex", Parameters: []string{"^(bad|worse)$"}},
			},
		},
		{
			name:  "escaped pipe",
			input: `required\|string`,
			expected: []ParsedRule{
				{Name: "required|string"},
			},
		},
		{
			name:  "whitespace around rules",
			input: " required | string ",
			expected: []ParsedRule{
				{Name: "required"},
				{Name: "string"},
			},
		},
		{
			name:  "empty segments skipped",
			input: "required||string",
			expected: []ParsedRule{
				{Name: "required"},
				{Name: "string"},
			},
		},
		{
			name:  "rule with empty parameter",
			input: "regex:",
			expected: []ParsedRule{
				{Name: "regex", Parameters: []string{""}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseRules(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseRuleSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []ParsedRule
	}{
		{
			name:  "simple rules",
			input: []string{"required", "string"},
			expected: []ParsedRule{
				{Name: "required"},
				{Name: "string"},
			},
		},
		{
			name:  "regex with pipe in pattern",
			input: []string{"required", "regex:^(foo|bar)$", "string"},
			expected: []ParsedRule{
				{Name: "required"},
				{Name: "regex", Parameters: []string{"^(foo|bar)$"}},
				{Name: "string"},
			},
		},
		{
			name:  "regex with pipe NOT at end preserves following rules",
			input: []string{"required", "regex:^(a|b|c)$", "string", "max:10"},
			expected: []ParsedRule{
				{Name: "required"},
				{Name: "regex", Parameters: []string{"^(a|b|c)$"}},
				{Name: "string"},
				{Name: "max", Parameters: []string{"10"}},
			},
		},
		{
			name:  "not_regex with pipe in pattern",
			input: []string{"required", "not_regex:^(bad|worse)$"},
			expected: []ParsedRule{
				{Name: "required"},
				{Name: "not_regex", Parameters: []string{"^(bad|worse)$"}},
			},
		},
		{
			name:  "rules with parameters",
			input: []string{"required", "max:255", "in:a,b,c"},
			expected: []ParsedRule{
				{Name: "required"},
				{Name: "max", Parameters: []string{"255"}},
				{Name: "in", Parameters: []string{"a", "b", "c"}},
			},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []ParsedRule{},
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: []ParsedRule{},
		},
		{
			name:     "whitespace entries skipped",
			input:    []string{"required", " ", "", "string"},
			expected: []ParsedRule{{Name: "required"}, {Name: "string"}},
		},
		{
			name:  "single rule with colon",
			input: []string{"min:5"},
			expected: []ParsedRule{
				{Name: "min", Parameters: []string{"5"}},
			},
		},
		{
			name:  "escaped comma in parameter",
			input: []string{`in:a\,b,c`},
			expected: []ParsedRule{
				{Name: "in", Parameters: []string{"a,b", "c"}},
			},
		},
		{
			name:  "complex regex pattern",
			input: []string{"regex:^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"},
			expected: []ParsedRule{
				{Name: "regex", Parameters: []string{"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseRuleSlice(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitRules(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple pipe split",
			input:    "required|string",
			expected: []string{"required", "string"},
		},
		{
			name:     "single rule",
			input:    "required",
			expected: []string{"required"},
		},
		{
			name:     "escaped pipe",
			input:    `required\|string`,
			expected: []string{"required|string"},
		},
		{
			name:     "regex consumes rest",
			input:    "required|regex:^(foo|bar)$|string",
			expected: []string{"required", "regex:^(foo|bar)$|string"},
		},
		{
			name:     "not_regex consumes rest",
			input:    "required|not_regex:^(x|y)$|min:3",
			expected: []string{"required", "not_regex:^(x|y)$|min:3"},
		},
		{
			name:     "regex at end with pipe in pattern",
			input:    "required|regex:^(a|b)$",
			expected: []string{"required", "regex:^(a|b)$"},
		},
		{
			name:     "no regex normal split",
			input:    "required|string|max:255",
			expected: []string{"required", "string", "max:255"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "trailing pipe",
			input:    "required|",
			expected: []string{"required"},
		},
		{
			name:     "leading pipe",
			input:    "|required",
			expected: []string{"required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitRules(tt.input)
			if tt.expected == nil {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestExtractRuleName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"required", "required"},
		{"max:255", "max"},
		{"regex:^(foo|bar)$", "regex"},
		{"not_regex:pattern", "not_regex"},
		{"in:a,b,c", "in"},
		{"  required  ", "required"},
		{"  max:255  ", "max"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractRuleName(tt.input))
		})
	}
}

func TestParseOneRule(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ParsedRule
	}{
		{
			name:     "simple rule",
			input:    "required",
			expected: ParsedRule{Name: "required"},
		},
		{
			name:     "rule with single parameter",
			input:    "max:255",
			expected: ParsedRule{Name: "max", Parameters: []string{"255"}},
		},
		{
			name:     "rule with multiple parameters",
			input:    "in:a,b,c",
			expected: ParsedRule{Name: "in", Parameters: []string{"a", "b", "c"}},
		},
		{
			name:     "regex rule",
			input:    "regex:^(foo|bar)$",
			expected: ParsedRule{Name: "regex", Parameters: []string{"^(foo|bar)$"}},
		},
		{
			name:     "not_regex rule",
			input:    "not_regex:^(bad|worse)$",
			expected: ParsedRule{Name: "not_regex", Parameters: []string{"^(bad|worse)$"}},
		},
		{
			name:     "regex with empty pattern",
			input:    "regex:",
			expected: ParsedRule{Name: "regex", Parameters: []string{""}},
		},
		{
			name:     "rule with escaped comma",
			input:    `in:a\,b,c`,
			expected: ParsedRule{Name: "in", Parameters: []string{"a,b", "c"}},
		},
		{
			name:     "rule with colon in parameter",
			input:    "between:1,10",
			expected: ParsedRule{Name: "between", Parameters: []string{"1", "10"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseOneRule(tt.input))
		})
	}
}

func TestSplitParameters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single parameter",
			input:    "255",
			expected: []string{"255"},
		},
		{
			name:     "multiple parameters",
			input:    "a,b,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "escaped comma",
			input:    `a\,b,c`,
			expected: []string{"a,b", "c"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single empty parameter from trailing comma",
			input:    "a,",
			expected: []string{"a", ""},
		},
		{
			name:     "multiple escaped commas",
			input:    `a\,b\,c,d`,
			expected: []string{"a,b,c", "d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitParameters(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
