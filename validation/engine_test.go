package validation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsvalidation "github.com/goravel/framework/contracts/validation"
)

// testRule is a simple Rule implementation for testing.
type testRule struct {
	signature string
	passes    func(ctx context.Context, data contractsvalidation.Data, val any, options ...any) bool
	message   string
}

func (r *testRule) Signature() string { return r.signature }
func (r *testRule) Passes(ctx context.Context, data contractsvalidation.Data, val any, options ...any) bool {
	return r.passes(ctx, data, val, options...)
}
func (r *testRule) Message(_ context.Context) string { return r.message }

func newAlwaysPassRule(name string) *testRule {
	return &testRule{
		signature: name,
		passes:    func(_ context.Context, _ contractsvalidation.Data, _ any, _ ...any) bool { return true },
		message:   "The :attribute field is invalid.",
	}
}

func newAlwaysFailRule(name string, msg string) *testRule {
	return &testRule{
		signature: name,
		passes:    func(_ context.Context, _ contractsvalidation.Data, _ any, _ ...any) bool { return false },
		message:   msg,
	}
}

func TestEngine_Validate_NoRules(t *testing.T) {
	bag, _ := NewDataBag(map[string]any{"name": "Alice"})
	engine := NewEngine(context.Background(), bag, map[string][]ParsedRule{}, engineOptions{})

	errors := engine.Validate()
	assert.True(t, errors.IsEmpty())
}

func TestEngine_Validate_CustomRulePasses(t *testing.T) {
	bag, _ := NewDataBag(map[string]any{"name": "Alice"})
	rules := map[string][]ParsedRule{
		"name": {{Name: "custom_pass"}},
	}
	engine := NewEngine(context.Background(), bag, rules, engineOptions{
		customRules: map[string]contractsvalidation.Rule{
			"custom_pass": newAlwaysPassRule("custom_pass"),
		},
	})

	errors := engine.Validate()
	assert.True(t, errors.IsEmpty())
}

func TestEngine_Validate_CustomRuleFails(t *testing.T) {
	bag, _ := NewDataBag(map[string]any{"name": "Alice"})
	rules := map[string][]ParsedRule{
		"name": {{Name: "custom_fail"}},
	}
	engine := NewEngine(context.Background(), bag, rules, engineOptions{
		customRules: map[string]contractsvalidation.Rule{
			"custom_fail": newAlwaysFailRule("custom_fail", "The :attribute field failed custom validation."),
		},
	})

	errors := engine.Validate()
	assert.True(t, errors.Has("name"))
	assert.Equal(t, "The name field failed custom validation.", errors.One("name"))
}

func TestEngine_Validate_UnknownRuleFails(t *testing.T) {
	bag, _ := NewDataBag(map[string]any{"name": "Alice"})
	rules := map[string][]ParsedRule{
		"name": {{Name: "nonexistent_rule"}},
	}
	engine := NewEngine(context.Background(), bag, rules, engineOptions{})

	errors := engine.Validate()
	assert.True(t, errors.Has("name"))
}

func TestEngine_Validate_MissingField_SkipsNonImplicit(t *testing.T) {
	bag, _ := NewDataBag(map[string]any{})
	rules := map[string][]ParsedRule{
		"missing_field": {{Name: "custom_fail"}},
	}
	engine := NewEngine(context.Background(), bag, rules, engineOptions{
		customRules: map[string]contractsvalidation.Rule{
			"custom_fail": newAlwaysFailRule("custom_fail", "failed"),
		},
	})

	errors := engine.Validate()
	// Non-implicit rules are skipped when field is missing
	assert.True(t, errors.IsEmpty())
}

func TestEngine_Validate_Sometimes(t *testing.T) {
	t.Run("field present - validates", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"name": "Alice"})
		rules := map[string][]ParsedRule{
			"name": {{Name: "sometimes"}, {Name: "custom_fail"}},
		}
		engine := NewEngine(context.Background(), bag, rules, engineOptions{
			customRules: map[string]contractsvalidation.Rule{
				"custom_fail": newAlwaysFailRule("custom_fail", "failed"),
			},
		})

		errors := engine.Validate()
		assert.True(t, errors.Has("name"))
	})

	t.Run("field absent - skips", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{})
		rules := map[string][]ParsedRule{
			"name": {{Name: "sometimes"}, {Name: "custom_fail"}},
		}
		engine := NewEngine(context.Background(), bag, rules, engineOptions{
			customRules: map[string]contractsvalidation.Rule{
				"custom_fail": newAlwaysFailRule("custom_fail", "failed"),
			},
		})

		errors := engine.Validate()
		assert.True(t, errors.IsEmpty())
	})
}

func TestEngine_Validate_Nullable(t *testing.T) {
	bag, _ := NewDataBag(map[string]any{"name": nil})
	rules := map[string][]ParsedRule{
		"name": {{Name: "nullable"}, {Name: "custom_fail"}},
	}
	engine := NewEngine(context.Background(), bag, rules, engineOptions{
		customRules: map[string]contractsvalidation.Rule{
			"custom_fail": newAlwaysFailRule("custom_fail", "failed"),
		},
	})

	errors := engine.Validate()
	// Nullable + nil value => non-implicit rules are skipped
	assert.True(t, errors.IsEmpty())
}

func TestEngine_Validate_Bail(t *testing.T) {
	bag, _ := NewDataBag(map[string]any{"name": "Alice"})
	rules := map[string][]ParsedRule{
		"name": {
			{Name: "bail"},
			{Name: "fail_one"},
			{Name: "fail_two"},
		},
	}
	engine := NewEngine(context.Background(), bag, rules, engineOptions{
		customRules: map[string]contractsvalidation.Rule{
			"fail_one": newAlwaysFailRule("fail_one", "first error"),
			"fail_two": newAlwaysFailRule("fail_two", "second error"),
		},
	})

	errors := engine.Validate()
	// With bail, only first error should be recorded
	fieldErrors := errors.Get("name")
	assert.Len(t, fieldErrors, 1)
	assert.Equal(t, "first error", fieldErrors["fail_one"])
}

func TestEngine_Validate_DeterministicFieldOrder(t *testing.T) {
	bag, _ := NewDataBag(map[string]any{"b": "val", "a": "val", "c": "val"})
	rules := map[string][]ParsedRule{
		"b": {{Name: "custom_fail"}},
		"a": {{Name: "custom_fail"}},
		"c": {{Name: "custom_fail"}},
	}
	engine := NewEngine(context.Background(), bag, rules, engineOptions{
		customRules: map[string]contractsvalidation.Rule{
			"custom_fail": newAlwaysFailRule("custom_fail", "failed :attribute"),
		},
	})

	errors := engine.Validate()
	assert.True(t, errors.Has("a"))
	assert.True(t, errors.Has("b"))
	assert.True(t, errors.Has("c"))
}

func TestEngine_ValidatedData(t *testing.T) {
	t.Run("returns only ruled fields", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"name": "Alice", "extra": "data"})
		rules := map[string][]ParsedRule{
			"name": {{Name: "custom_pass"}},
		}
		engine := NewEngine(context.Background(), bag, rules, engineOptions{
			customRules: map[string]contractsvalidation.Rule{
				"custom_pass": newAlwaysPassRule("custom_pass"),
			},
		})
		engine.Validate()

		data := engine.ValidatedData()
		assert.Equal(t, "Alice", data["name"])
		assert.NotContains(t, data, "extra")
	})

	t.Run("excludes excluded fields", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"name": "Alice", "secret": "hide"})
		rules := map[string][]ParsedRule{
			"name":   {{Name: "custom_pass"}},
			"secret": {{Name: "custom_pass"}},
		}
		engine := NewEngine(context.Background(), bag, rules, engineOptions{
			customRules: map[string]contractsvalidation.Rule{
				"custom_pass": newAlwaysPassRule("custom_pass"),
			},
		})
		engine.Validate()
		engine.excludes["secret"] = true

		data := engine.ValidatedData()
		assert.Equal(t, "Alice", data["name"])
		assert.NotContains(t, data, "secret")
	})

	t.Run("wildcard arrays are reconstructed as slices", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{
			"tags":   []any{"tag1", "tag2"},
			"scores": []int{1, 2},
		})
		rules := map[string][]ParsedRule{
			"tags.*":   {{Name: "custom_pass"}},
			"scores.*": {{Name: "custom_pass"}},
		}
		engine := NewEngine(context.Background(), bag, rules, engineOptions{
			customRules: map[string]contractsvalidation.Rule{
				"custom_pass": newAlwaysPassRule("custom_pass"),
			},
		})
		engine.Validate()

		data := engine.ValidatedData()

		tags, ok := data["tags"].([]any)
		assert.True(t, ok)
		assert.Equal(t, []any{"tag1", "tag2"}, tags)

		scores, ok := data["scores"].([]int)
		assert.True(t, ok)
		assert.Equal(t, []int{1, 2}, scores)
	})

	t.Run("sparse indexed wildcard falls back to []any with nil gaps", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{
			"items": []int{10, 20, 30},
		})
		rules := map[string][]ParsedRule{
			"items.2": {{Name: "custom_pass"}},
		}
		engine := NewEngine(context.Background(), bag, rules, engineOptions{
			customRules: map[string]contractsvalidation.Rule{
				"custom_pass": newAlwaysPassRule("custom_pass"),
			},
		})
		engine.Validate()

		data := engine.ValidatedData()
		items, ok := data["items"].([]any)
		assert.True(t, ok)
		assert.Equal(t, []any{nil, nil, 30}, items)
	})
}

func TestEngine_HandleExcludeRule(t *testing.T) {
	t.Run("exclude always excludes", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"field": "val"})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{})

		engine.handleExcludeRule("field", ParsedRule{Name: "exclude"})
		assert.True(t, engine.isExcluded("field"))
	})

	t.Run("exclude_if matches", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"type": "admin", "secret": "val"})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{})

		engine.handleExcludeRule("secret", ParsedRule{
			Name:       "exclude_if",
			Parameters: []string{"type", "admin"},
		})
		assert.True(t, engine.isExcluded("secret"))
	})

	t.Run("exclude_if does not match", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"type": "user", "secret": "val"})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{})

		engine.handleExcludeRule("secret", ParsedRule{
			Name:       "exclude_if",
			Parameters: []string{"type", "admin"},
		})
		assert.False(t, engine.isExcluded("secret"))
	})

	t.Run("exclude_unless excludes when not matching", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"type": "guest", "field": "val"})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{})

		engine.handleExcludeRule("field", ParsedRule{
			Name:       "exclude_unless",
			Parameters: []string{"type", "admin"},
		})
		assert.True(t, engine.isExcluded("field"))
	})

	t.Run("exclude_unless keeps when matching", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"type": "admin", "field": "val"})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{})

		engine.handleExcludeRule("field", ParsedRule{
			Name:       "exclude_unless",
			Parameters: []string{"type", "admin"},
		})
		assert.False(t, engine.isExcluded("field"))
	})

	t.Run("exclude_with excludes when other field present", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"other": "val", "field": "val"})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{})

		engine.handleExcludeRule("field", ParsedRule{
			Name:       "exclude_with",
			Parameters: []string{"other"},
		})
		assert.True(t, engine.isExcluded("field"))
	})

	t.Run("exclude_with keeps when other field absent", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"field": "val"})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{})

		engine.handleExcludeRule("field", ParsedRule{
			Name:       "exclude_with",
			Parameters: []string{"other"},
		})
		assert.False(t, engine.isExcluded("field"))
	})

	t.Run("exclude_without excludes when other field absent", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"field": "val"})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{})

		engine.handleExcludeRule("field", ParsedRule{
			Name:       "exclude_without",
			Parameters: []string{"other"},
		})
		assert.True(t, engine.isExcluded("field"))
	})

	t.Run("exclude_without keeps when other field present", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"other": "val", "field": "val"})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{})

		engine.handleExcludeRule("field", ParsedRule{
			Name:       "exclude_without",
			Parameters: []string{"other"},
		})
		assert.False(t, engine.isExcluded("field"))
	})
}

func TestEngine_ExpandWildcardRules(t *testing.T) {
	bag, _ := NewDataBag(map[string]any{
		"users": []any{
			map[string]any{"name": "Alice"},
			map[string]any{"name": "Bob"},
		},
	})
	rules := map[string][]ParsedRule{
		"users.*.name": {{Name: "custom_pass"}},
	}
	engine := NewEngine(context.Background(), bag, rules, engineOptions{
		customRules: map[string]contractsvalidation.Rule{
			"custom_pass": newAlwaysPassRule("custom_pass"),
		},
	})

	expanded := engine.expandWildcardRules()
	assert.Contains(t, expanded, "users.0.name")
	assert.Contains(t, expanded, "users.1.name")

	// Results should be cached
	expanded2 := engine.expandWildcardRules()
	assert.Equal(t, expanded, expanded2)
}

func TestEngine_ExpandWildcardRules_TypedSlice(t *testing.T) {
	bag, _ := NewDataBag(map[string]any{
		"scores": []int{1, 2},
	})
	rules := map[string][]ParsedRule{
		"scores.*": {{Name: "required"}},
	}
	engine := NewEngine(context.Background(), bag, rules, engineOptions{})

	expanded := engine.expandWildcardRules()
	assert.Contains(t, expanded, "scores.0")
	assert.Contains(t, expanded, "scores.1")
}

func TestEngine_TrackDistinct(t *testing.T) {
	rules := map[string][]ParsedRule{
		"items.*.id": {{Name: "distinct"}},
	}
	bag, _ := NewDataBag(map[string]any{
		"items": []any{
			map[string]any{"id": 1},
			map[string]any{"id": 2},
			map[string]any{"id": 1},
		},
	})
	engine := NewEngine(context.Background(), bag, rules, engineOptions{})

	assert.False(t, engine.trackDistinct("items.0.id", 1)) // first occurrence
	assert.False(t, engine.trackDistinct("items.1.id", 2)) // different value
	assert.True(t, engine.trackDistinct("items.2.id", 1))  // duplicate
}

func TestEngine_FormatErrorMessage(t *testing.T) {
	t.Run("custom rule message without custom message override", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{
			customRules: map[string]contractsvalidation.Rule{
				"my_rule": newAlwaysFailRule("my_rule", "The :attribute is bad."),
			},
			attributes: map[string]string{"name": "Full Name"},
		})

		msg := engine.formatErrorMessage("name", ParsedRule{Name: "my_rule"}, "string")
		assert.Equal(t, "The Full Name is bad.", msg)
	})

	t.Run("custom field+rule message overrides custom rule message", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{
			customRules: map[string]contractsvalidation.Rule{
				"custom_exists": newAlwaysFailRule("custom_exists", "The :attribute does not exist in custom rule."),
			},
			messages: map[string]string{
				"f.custom_exists": "custom_exists failed for :attribute",
			},
		})

		msg := engine.formatErrorMessage("f", ParsedRule{Name: "custom_exists"}, "string")
		assert.Equal(t, "custom_exists failed for f", msg)
	})

	t.Run("custom rule message override overrides custom rule message", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{
			customRules: map[string]contractsvalidation.Rule{
				"custom_exists": newAlwaysFailRule("custom_exists", "The :attribute does not exist in custom rule."),
			},
			messages: map[string]string{
				"custom_exists": "custom_exists failed",
			},
		})

		msg := engine.formatErrorMessage("f", ParsedRule{Name: "custom_exists"}, "string")
		assert.Equal(t, "custom_exists failed", msg)
	})

	t.Run("custom message override", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{
			messages: map[string]string{
				"name.required": "Please enter your name.",
			},
		})

		msg := engine.formatErrorMessage("name", ParsedRule{Name: "required"}, "string")
		assert.Equal(t, "Please enter your name.", msg)
	})

	t.Run("min rule with parameter", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{})

		msg := engine.formatErrorMessage("name", ParsedRule{
			Name:       "min",
			Parameters: []string{"3"},
		}, "string")
		assert.Equal(t, "The name field must be at least 3 characters.", msg)
	})

	t.Run("between rule with parameters", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{})

		msg := engine.formatErrorMessage("age", ParsedRule{
			Name:       "between",
			Parameters: []string{"1", "100"},
		}, "numeric")
		assert.Equal(t, "The age field must be between 1 and 100.", msg)
	})

	t.Run("custom attribute name", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{
			attributes: map[string]string{"first_name": "First Name"},
		})

		msg := engine.formatErrorMessage("first_name", ParsedRule{Name: "required"}, "string")
		assert.Equal(t, "The First Name field is required.", msg)
	})
}

func TestEngine_ExecuteRule(t *testing.T) {
	t.Run("custom rule", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"name": "Alice"})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{
			customRules: map[string]contractsvalidation.Rule{
				"custom_pass": newAlwaysPassRule("custom_pass"),
			},
		})

		passed := engine.executeRule("name", ParsedRule{Name: "custom_pass"}, "Alice", nil)
		assert.True(t, passed)
	})

	t.Run("unknown rule returns false", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"name": "Alice"})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{})

		passed := engine.executeRule("name", ParsedRule{Name: "nonexistent"}, "Alice", nil)
		assert.False(t, passed)
	})

	t.Run("custom rule with parameters", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"name": "Alice"})
		engine := NewEngine(context.Background(), bag, nil, engineOptions{
			customRules: map[string]contractsvalidation.Rule{
				"min_len": &testRule{
					signature: "min_len",
					passes: func(_ context.Context, _ contractsvalidation.Data, val any, options ...any) bool {
						s, ok := val.(string)
						if !ok {
							return false
						}
						require.Len(t, options, 1)
						return len(s) >= 3
					},
					message: "too short",
				},
			},
		})

		passed := engine.executeRule("name", ParsedRule{
			Name:       "min_len",
			Parameters: []string{"3"},
		}, "Alice", nil)
		assert.True(t, passed)
	})
}

func TestAnySlice(t *testing.T) {
	result := anySlice([]string{"a", "b", "c"})
	assert.Equal(t, []any{"a", "b", "c"}, result)

	result = anySlice(nil)
	assert.Equal(t, []any{}, result)
}
