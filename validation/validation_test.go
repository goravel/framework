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
		expectValidator bool
		expectErr       error
	}{
		{
			description:     "success when data is map[string]any",
			data:            map[string]any{"a": "b"},
			rules:           map[string]any{"a": "some_rule"},
			expectValidator: true,
		},
		{
			description:     "success when data is struct",
			data:            &Data{A: "b"},
			rules:           map[string]any{"a": "some_rule"},
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
			rules:       map[string]any{"a": "some_rule"},
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
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			validation := NewValidation()
			validator, err := validation.Make(context.Background(), test.data, test.rules, test.options...)
			assert.Equal(t, test.expectValidator, validator != nil, test.description)
			if test.expectErr != nil {
				assert.ErrorIs(t, err, test.expectErr, test.description)
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

func TestCustomFilterIntegration(t *testing.T) {
	validation := NewValidation()
	err := validation.AddFilters([]httpvalidate.Filter{&DefaultFilter{}})
	assert.Nil(t, err)

	validator, err := validation.Make(context.Background(), map[string]any{
		"name":  "krishan ",
		"empty": "",
	}, map[string]any{
		"name":  "some_rule",
		"empty": "some_rule",
	}, Filters(map[string]any{
		"empty": "default:emptyDefault",
	}))
	assert.Nil(t, err)

	var mp map[string]any
	assert.Nil(t, validator.Bind(&mp))
	assert.Equal(t, "emptyDefault", mp["empty"])
}

func TestValidated(t *testing.T) {
	validation := NewValidation()
	validator, err := validation.Make(context.Background(), map[string]any{
		"name":  "goravel",
		"email": "test@example.com",
		"extra": "not in rules",
	}, map[string]any{
		"name":  "some_rule",
		"email": "some_rule",
	})
	assert.Nil(t, err)
	assert.False(t, validator.Fails())

	validated := validator.Validated()
	assert.Equal(t, "goravel", validated["name"])
	assert.Equal(t, "test@example.com", validated["email"])
	_, exists := validated["extra"]
	assert.False(t, exists)
}

func TestSliceRuleSyntax(t *testing.T) {
	validation := NewValidation()

	t.Run("slice syntax accepted", func(t *testing.T) {
		validator, err := validation.Make(context.Background(), map[string]any{
			"name": "goravel",
		}, map[string]any{
			"name": []string{"some_rule", "another_rule"},
		})
		assert.Nil(t, err)
		assert.NotNil(t, validator)
		assert.False(t, validator.Fails())
	})

	t.Run("invalid rule type returns error", func(t *testing.T) {
		_, err := validation.Make(context.Background(), map[string]any{
			"name": "goravel",
		}, map[string]any{
			"name": 123,
		})
		assert.ErrorIs(t, err, errors.ValidationInvalidRuleType)
	})
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
