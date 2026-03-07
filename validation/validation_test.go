package validation

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	httpvalidate "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/errors"
)

func TestMake(t *testing.T) {
	type Data struct {
		A string `form:"a"`
	}

	tests := []struct {
		description     string
		data            any
		rules           map[string]any
		options         []httpvalidate.Option
		customRules     []httpvalidate.Rule
		expectValidator bool
		expectErr       error
	}{
		{
			description:     "success when data is map[string]any",
			data:            map[string]any{"a": "b"},
			rules:           map[string]any{"a": "custom_uppercase"},
			customRules:     []httpvalidate.Rule{&CustomUppercase{}},
			expectValidator: true,
		},
		{
			description:     "success when data is struct",
			data:            &Data{A: "b"},
			rules:           map[string]any{"a": "custom_uppercase"},
			customRules:     []httpvalidate.Rule{&CustomUppercase{}},
			expectValidator: true,
		},
		{
			description: "error when data isn't map[string]any or map[string][]string or struct",
			data:        "1   ",
			rules:       map[string]any{"a": "some_rule"},
			expectErr:   errors.ValidationDataInvalidType,
		},
		{
			description: "error when data is nil",
			data:        nil,
			rules:       map[string]any{"a": "some_rule"},
			expectErr:   errors.ValidationEmptyData,
		},
		{
			description: "error when rule is empty map",
			data:        map[string]any{"a": "b"},
			rules:       map[string]any{},
			expectErr:   errors.ValidationEmptyRules,
		},
		{
			description: "error when PrepareForValidation returns error",
			data:        map[string]any{"a": "b"},
			rules:       map[string]any{"a": "some_rule"},
			options: []httpvalidate.Option{
				PrepareForValidation(func(ctx context.Context, data httpvalidate.Data) error {
					return assert.AnError
				}),
			},
			expectErr: assert.AnError,
		},
		{
			description: "success with PrepareForValidation modifying data",
			data:        map[string]any{"a": "b"},
			rules:       map[string]any{"a": "custom_uppercase"},
			customRules: []httpvalidate.Rule{&CustomUppercase{}},
			options: []httpvalidate.Option{
				PrepareForValidation(func(ctx context.Context, data httpvalidate.Data) error {
					return data.Set("a", "c")
				}),
			},
			expectValidator: true,
		},
		{
			description: "error when rule type is invalid",
			data:        map[string]any{"a": "b"},
			rules:       map[string]any{"a": 123},
			expectErr:   errors.ValidationInvalidRuleType,
		},
		{
			description: "error when rule is unknown",
			data:        map[string]any{"a": "b"},
			rules:       map[string]any{"a": "unknown_rule"},
			expectErr:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			validation := NewValidation()
			if len(test.customRules) > 0 {
				err := validation.AddRules(test.customRules)
				assert.Nil(t, err)
			}
			validator, err := validation.Make(context.Background(), test.data, test.rules, test.options...)
			assert.Equal(t, test.expectValidator, validator != nil, test.description)
			if test.expectErr != nil {
				assert.ErrorIs(t, err, test.expectErr, test.description)
			} else if test.description == "error when rule is unknown" {
				assert.Error(t, err, test.description)
				assert.Contains(t, err.Error(), "unknown validation rule")
			} else {
				assert.Nil(t, err, test.description)
			}
		})
	}
}

func TestAddRules(t *testing.T) {
	validation := NewValidation()

	t.Run("success", func(t *testing.T) {
		err := validation.AddRules([]httpvalidate.Rule{&CustomUppercase{}})
		assert.Nil(t, err)
	})

	t.Run("duplicate custom rule", func(t *testing.T) {
		err := validation.AddRules([]httpvalidate.Rule{&CustomUppercase{}})
		assert.Error(t, err)
	})
}

func TestAddFilters(t *testing.T) {
	validation := NewValidation()

	t.Run("success", func(t *testing.T) {
		err := validation.AddFilters([]httpvalidate.Filter{&DefaultFilter{}})
		assert.Nil(t, err)
	})

	t.Run("duplicate filter", func(t *testing.T) {
		err := validation.AddFilters([]httpvalidate.Filter{&DefaultFilter{}})
		assert.Error(t, err)
	})
}

func TestCustomRule(t *testing.T) {
	validation := NewValidation()
	err := validation.AddRules([]httpvalidate.Rule{&CustomUppercase{}, &CustomLowercase{}})
	assert.Nil(t, err)

	t.Run("success", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"name":    "ABC",
			"address": "de",
		}, map[string]any{
			"name":    "custom_uppercase:3",
			"address": "custom_lowercase:2",
		})
		assert.Nil(t, err)
		assert.NotNil(t, validator)
		assert.False(t, validator.Fails())
	})

	t.Run("error", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"name":    "abc",
			"address": "DE",
		}, map[string]any{
			"name":    "custom_uppercase:3",
			"address": "custom_lowercase:2",
		})
		assert.Nil(t, err)
		assert.NotNil(t, validator)
		assert.Equal(t, map[string]string{"custom_uppercase": "name must be upper"}, validator.Errors().Get("name"))
		assert.Equal(t, map[string]string{"custom_lowercase": "address must be lower"}, validator.Errors().Get("address"))
	})
}

func TestCustomFilter(t *testing.T) {
	validation := NewValidation()
	err := validation.AddFilters([]httpvalidate.Filter{&DefaultFilter{}})
	assert.Nil(t, err)

	filters := validation.Filters()
	defaultFilterFunc := filters[0].Handle(context.Background()).(func(string, ...string) string)
	assert.Equal(t, "default", defaultFilterFunc("", "default"))
	assert.Equal(t, "a", defaultFilterFunc("a"))
}

// --- Test fixtures ---

type CustomUppercase struct{}

func (receiver *CustomUppercase) Signature() string {
	return "custom_uppercase"
}

func (receiver *CustomUppercase) Passes(ctx context.Context, data httpvalidate.Data, val any, options ...any) bool {
	name, exist := data.Get("name")
	if len(options) > 0 {
		return strings.ToUpper(val.(string)) == val.(string) && len(val.(string)) == cast.ToInt(options[0]) && name == val && exist
	}
	return false
}

func (receiver *CustomUppercase) Message(ctx context.Context) string {
	return ":attribute must be upper"
}

type CustomLowercase struct{}

func (receiver *CustomLowercase) Signature() string {
	return "custom_lowercase"
}

func (receiver *CustomLowercase) Passes(ctx context.Context, data httpvalidate.Data, val any, options ...any) bool {
	address, exist := data.Get("address")
	if len(options) > 0 {
		return strings.ToLower(val.(string)) == val.(string) && len(val.(string)) == cast.ToInt(options[0]) && address == val && exist
	}
	return false
}

func (receiver *CustomLowercase) Message(ctx context.Context) string {
	return ":attribute must be lower"
}

type DefaultFilter struct{}

func (receiver *DefaultFilter) Signature() string {
	return "default"
}

func (receiver *DefaultFilter) Handle(ctx context.Context) any {
	return func(val string, def ...string) string {
		if val == "" {
			if len(def) > 0 {
				return def[0]
			}
		}
		return val
	}
}
