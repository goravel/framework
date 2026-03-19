package validation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsvalidation "github.com/goravel/framework/contracts/validation"
)

func TestBuiltinFilters(t *testing.T) {
	tests := []struct {
		name     string
		filter   string
		input    any
		expected any
	}{
		// String cleaning
		{"trim", "trim", "  hello  ", "hello"},
		{"ltrim", "ltrim", "  hello  ", "hello  "},
		{"rtrim", "rtrim", "  hello  ", "  hello"},

		// Case conversion
		{"lower", "lower", "HELLO", "hello"},
		{"upper", "upper", "hello", "HELLO"},
		{"title", "title", "hello world", "Hello World"},
		{"ucfirst", "ucfirst", "hello", "Hello"},
		{"ucfirst empty", "ucfirst", "", ""},
		{"lcfirst", "lcfirst", "Hello", "hello"},
		{"lcfirst empty", "lcfirst", "", ""},

		// Naming style
		{"camel from snake", "camel", "hello_world", "helloWorld"},
		{"camel from words", "camel", "hello world", "helloWorld"},
		{"snake from camel", "snake", "helloWorld", "hello_world"},
		{"snake from words", "snake", "hello world", "hello_world"},

		// Type conversion
		{"to_int from string", "to_int", "42", 42},
		{"to_int from float", "to_int", 42.9, 42},
		{"to_int64 from string", "to_int64", "9999999999", int64(9999999999)},
		{"to_int64 from int", "to_int64", 42, int64(42)},
		{"to_uint from string", "to_uint", "42", uint(42)},
		{"to_float from string", "to_float", "3.14", 3.14},
		{"to_bool true", "to_bool", "true", true},
		{"to_bool false", "to_bool", "false", false},
		{"to_string from int", "to_string", 42, "42"},
		{"to_time", "to_time", "2024-01-01", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"int alias", "int", "42", 42},
		{"int64 alias", "int64", "9999999999", int64(9999999999)},
		{"uint alias", "uint", "42", uint(42)},
		{"float alias", "float", "3.14", 3.14},
		{"bool alias", "bool", "true", true},

		// Encoding
		{"strip_tags", "strip_tags", "<p>Hello <b>World</b></p>", "Hello World"},
		{"strip_tags no tags", "strip_tags", "Hello World", "Hello World"},
		{"escape_js", "escape_js", `<script>alert("xss")</script>`, `\x3cscript\x3ealert(\"xss\")\x3c\/script\x3e`},
		{"escape_js newlines", "escape_js", "line1\nline2", `line1\nline2`},
		{"escape_html", "escape_html", "<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"url_encode", "url_encode", "hello world&foo=bar", "hello+world%26foo%3Dbar"},
		{"url_decode", "url_decode", "hello+world%26foo%3Dbar", "hello world&foo=bar"},
		{"url_decode invalid", "url_decode", "%zz", "%zz"},

		// String splitting
		{"str_to_ints", "str_to_ints", "1,2,3", []int{1, 2, 3}},
		{"str_to_ints with spaces", "str_to_ints", "1, 2, 3", []int{1, 2, 3}},
		{"str_to_ints single", "str_to_ints", "42", []int{42}},
		{"str_to_array", "str_to_array", "a,b,c", []string{"a", "b", "c"}},
		{"str_to_array with spaces", "str_to_array", "a, b, c", []string{"a", "b", "c"}},
		{"str_to_array single", "str_to_array", "hello", []string{"hello"}},
		{"str_to_time", "str_to_time", "2024-01-01", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},

		// Deprecated aliases
		{"trimSpace", "trimSpace", "  hello  ", "hello"},
		{"trimLeft", "trimLeft", "  hello  ", "hello  "},
		{"trimRight", "trimRight", "  hello  ", "  hello"},
		{"lowercase", "lowercase", "HELLO", "hello"},
		{"uppercase", "uppercase", "hello", "HELLO"},
		{"lowerFirst", "lowerFirst", "Hello", "hello"},
		{"upperFirst", "upperFirst", "hello", "Hello"},
		{"ucWord", "ucWord", "hello world", "Hello World"},
		{"upperWord", "upperWord", "hello world", "Hello World"},
		{"camelCase", "camelCase", "hello_world", "helloWorld"},
		{"snakeCase", "snakeCase", "helloWorld", "hello_world"},
		{"toInt", "toInt", "42", 42},
		{"toUint", "toUint", "42", uint(42)},
		{"toInt64", "toInt64", "100", int64(100)},
		{"toFloat", "toFloat", "3.14", 3.14},
		{"toBool", "toBool", "true", true},
		{"toString", "toString", 42, "42"},
		{"toTime", "toTime", "2024-01-01", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"str2time", "str2time", "2024-01-01", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"strToTime", "strToTime", "2024-01-01", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"escapeHtml", "escapeHtml", "<b>hi</b>", "&lt;b&gt;hi&lt;/b&gt;"},
		{"escapeHTML", "escapeHTML", "<b>hi</b>", "&lt;b&gt;hi&lt;/b&gt;"},
		{"escapeJs", "escapeJs", "alert('x')", `alert(\'x\')`},
		{"escapeJS", "escapeJS", "alert('x')", `alert(\'x\')`},
		{"urlEncode", "urlEncode", "a b", "a+b"},
		{"urlDecode", "urlDecode", "a+b", "a b"},
		{"stripTags", "stripTags", "<p>hi</p>", "hi"},
		{"str2ints", "str2ints", "1,2", []int{1, 2}},
		{"strToInts", "strToInts", "1,2", []int{1, 2}},
		{"str2arr", "str2arr", "a,b", []string{"a", "b"}},
		{"str2array", "str2array", "a,b", []string{"a", "b"}},
		{"strToArray", "strToArray", "a,b", []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := builtinFilters[tt.filter]
			require.True(t, ok, "builtin filter %s not found", tt.filter)
			result := fn(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyFilters(t *testing.T) {
	t.Run("apply builtin filters", func(t *testing.T) {
		bag, err := NewDataBag(map[string]any{
			"name":  "  Hello World  ",
			"email": "TEST@EXAMPLE.COM",
		})
		require.NoError(t, err)

		err = applyFilters(context.Background(), bag, map[string]any{
			"name":  "trim|lower",
			"email": "lower",
		}, nil)
		require.NoError(t, err)

		val, _ := bag.Get("name")
		assert.Equal(t, "hello world", val)

		val, _ = bag.Get("email")
		assert.Equal(t, "test@example.com", val)
	})

	t.Run("skip non-existent field", func(t *testing.T) {
		bag, err := NewDataBag(map[string]any{
			"name": "hello",
		})
		require.NoError(t, err)

		err = applyFilters(context.Background(), bag, map[string]any{
			"nonexistent": "trim",
		}, nil)
		require.NoError(t, err)

		val, _ := bag.Get("name")
		assert.Equal(t, "hello", val)
	})

	t.Run("empty filter rules", func(t *testing.T) {
		bag, err := NewDataBag(map[string]any{
			"name": "hello",
		})
		require.NoError(t, err)

		err = applyFilters(context.Background(), bag, map[string]any{}, nil)
		require.NoError(t, err)

		val, _ := bag.Get("name")
		assert.Equal(t, "hello", val)
	})

	t.Run("apply custom filter", func(t *testing.T) {
		bag, err := NewDataBag(map[string]any{
			"name": "hello",
		})
		require.NoError(t, err)

		customFilter := &mockFilter{
			signature: "reverse",
			handler: func(val string) string {
				runes := []rune(val)
				for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
					runes[i], runes[j] = runes[j], runes[i]
				}
				return string(runes)
			},
		}

		err = applyFilters(context.Background(), bag, map[string]any{
			"name": "reverse",
		}, []contractsvalidation.Filter{customFilter})
		require.NoError(t, err)

		val, _ := bag.Get("name")
		assert.Equal(t, "olleh", val)
	})

	t.Run("chain builtin and custom filters", func(t *testing.T) {
		bag, err := NewDataBag(map[string]any{
			"name": "  Hello  ",
		})
		require.NoError(t, err)

		customFilter := &mockFilter{
			signature: "append_exclaim",
			handler: func(val string) string {
				return val + "!"
			},
		}

		err = applyFilters(context.Background(), bag, map[string]any{
			"name": "trim|lower|append_exclaim",
		}, []contractsvalidation.Filter{customFilter})
		require.NoError(t, err)

		val, _ := bag.Get("name")
		assert.Equal(t, "hello!", val)
	})

	t.Run("invalid filter type returns error", func(t *testing.T) {
		bag, err := NewDataBag(map[string]any{
			"name": "hello",
		})
		require.NoError(t, err)

		err = applyFilters(context.Background(), bag, map[string]any{
			"name": 123,
		}, nil)
		assert.Error(t, err)
	})
}

func TestCallFilterFunc(t *testing.T) {
	t.Run("single return value", func(t *testing.T) {
		filter := &mockFilter{
			signature: "double",
			handler: func(val int) int {
				return val * 2
			},
		}

		result, err := callFilterFunc(context.Background(), filter, 5, nil)
		assert.NoError(t, err)
		assert.Equal(t, 10, result)
	})

	t.Run("return value with error", func(t *testing.T) {
		filter := &mockFilter{
			signature: "safe_parse",
			handler: func(val string) (int, error) {
				if val == "bad" {
					return 0, fmt.Errorf("bad value")
				}
				return 42, nil
			},
		}

		result, err := callFilterFunc(context.Background(), filter, "good", nil)
		assert.NoError(t, err)
		assert.Equal(t, 42, result)

		_, err = callFilterFunc(context.Background(), filter, "bad", nil)
		assert.Error(t, err)
	})

	t.Run("variadic arguments", func(t *testing.T) {
		filter := &mockFilter{
			signature: "default_val",
			handler: func(val string, def ...string) string {
				if val == "" && len(def) > 0 {
					return def[0]
				}
				return val
			},
		}

		result, err := callFilterFunc(context.Background(), filter, "", []string{"default"})
		assert.NoError(t, err)
		assert.Equal(t, "default", result)

		result, err = callFilterFunc(context.Background(), filter, "actual", nil)
		assert.NoError(t, err)
		assert.Equal(t, "actual", result)
	})

	t.Run("nil handler", func(t *testing.T) {
		filter := &mockFilter{
			signature: "nil_handler",
			handler:   nil,
		}

		_, err := callFilterFunc(context.Background(), filter, "test", nil)
		assert.Error(t, err)
	})

	t.Run("non-function handler", func(t *testing.T) {
		filter := &mockFilter{
			signature: "not_func",
			handler:   "not a function",
		}

		_, err := callFilterFunc(context.Background(), filter, "test", nil)
		assert.Error(t, err)
	})
}

func TestFiltersIntegration(t *testing.T) {
	t.Run("filters applied before validation", func(t *testing.T) {
		validation := NewValidation()
		validator, err := validation.Make(context.Background(), map[string]any{
			"name": "  goravel  ",
		}, map[string]any{
			"name": "required|string",
		}, Filters(map[string]any{
			"name": "trim",
		}))
		assert.NoError(t, err)
		assert.False(t, validator.Fails())

		val := validator.Validated()
		assert.Equal(t, "goravel", val["name"])
	})

	t.Run("filters with type conversion", func(t *testing.T) {
		validation := NewValidation()
		validator, err := validation.Make(context.Background(), map[string]any{
			"age": "25",
		}, map[string]any{
			"age": "required",
		}, Filters(map[string]any{
			"age": "to_int",
		}))
		assert.NoError(t, err)
		assert.False(t, validator.Fails())

		val := validator.Validated()
		assert.Equal(t, 25, val["age"])
	})

	t.Run("filters with custom filter via AddFilters", func(t *testing.T) {
		validation := NewValidation()
		err := validation.AddFilters([]contractsvalidation.Filter{
			&mockFilter{
				signature: "prefix_hello",
				handler: func(val string) string {
					return "hello_" + val
				},
			},
		})
		require.NoError(t, err)

		validator, err := validation.Make(context.Background(), map[string]any{
			"name": "world",
		}, map[string]any{
			"name": "required",
		}, Filters(map[string]any{
			"name": "prefix_hello",
		}))
		assert.NoError(t, err)
		assert.False(t, validator.Fails())

		val := validator.Validated()
		assert.Equal(t, "hello_world", val["name"])
	})

	t.Run("filters with custom filter via option", func(t *testing.T) {
		validation := NewValidation()

		customFilter := &mockFilter{
			signature: "suffix_test",
			handler: func(val string) string {
				return val + "_test"
			},
		}

		validator, err := validation.Make(context.Background(), map[string]any{
			"name": "hello",
		}, map[string]any{
			"name": "required",
		},
			Filters(map[string]any{
				"name": "suffix_test",
			}),
			CustomFilters([]contractsvalidation.Filter{customFilter}),
		)
		assert.NoError(t, err)
		assert.False(t, validator.Fails())

		val := validator.Validated()
		assert.Equal(t, "hello_test", val["name"])
	})
}

func TestApplyFiltersWithWildcards(t *testing.T) {
	t.Run("wildcard filter on nested slice", func(t *testing.T) {
		bag, err := NewDataBag(map[string]any{
			"users": []any{
				map[string]any{"name": "  Alice  ", "email": "ALICE@EXAMPLE.COM"},
				map[string]any{"name": "  Bob  ", "email": "BOB@EXAMPLE.COM"},
			},
		})
		require.NoError(t, err)

		err = applyFilters(context.Background(), bag, map[string]any{
			"users.*.name":  "trim",
			"users.*.email": "lower",
		}, nil)
		require.NoError(t, err)

		val, _ := bag.Get("users.0.name")
		assert.Equal(t, "Alice", val)
		val, _ = bag.Get("users.1.name")
		assert.Equal(t, "Bob", val)
		val, _ = bag.Get("users.0.email")
		assert.Equal(t, "alice@example.com", val)
		val, _ = bag.Get("users.1.email")
		assert.Equal(t, "bob@example.com", val)
	})

	t.Run("wildcard filter with chained filters", func(t *testing.T) {
		bag, err := NewDataBag(map[string]any{
			"items": []any{
				map[string]any{"title": "  HELLO WORLD  "},
				map[string]any{"title": "  FOO BAR  "},
			},
		})
		require.NoError(t, err)

		err = applyFilters(context.Background(), bag, map[string]any{
			"items.*.title": "trim|lower",
		}, nil)
		require.NoError(t, err)

		val, _ := bag.Get("items.0.title")
		assert.Equal(t, "hello world", val)
		val, _ = bag.Get("items.1.title")
		assert.Equal(t, "foo bar", val)
	})

	t.Run("wildcard filter with no matching data", func(t *testing.T) {
		bag, err := NewDataBag(map[string]any{
			"name": "  hello  ",
		})
		require.NoError(t, err)

		err = applyFilters(context.Background(), bag, map[string]any{
			"users.*.name": "trim",
		}, nil)
		require.NoError(t, err)

		// Original data should not be affected
		val, _ := bag.Get("name")
		assert.Equal(t, "  hello  ", val)
	})

	t.Run("wildcard filter with slice syntax", func(t *testing.T) {
		bag, err := NewDataBag(map[string]any{
			"tags": []any{
				map[string]any{"value": "  Go  "},
				map[string]any{"value": "  Rust  "},
			},
		})
		require.NoError(t, err)

		err = applyFilters(context.Background(), bag, map[string]any{
			"tags.*.value": []string{"trim", "upper"},
		}, nil)
		require.NoError(t, err)

		val, _ := bag.Get("tags.0.value")
		assert.Equal(t, "GO", val)
		val, _ = bag.Get("tags.1.value")
		assert.Equal(t, "RUST", val)
	})

	t.Run("mixed wildcard and direct fields", func(t *testing.T) {
		bag, err := NewDataBag(map[string]any{
			"title": "  ADMIN  ",
			"users": []any{
				map[string]any{"name": "  Alice  "},
			},
		})
		require.NoError(t, err)

		err = applyFilters(context.Background(), bag, map[string]any{
			"title":        "trim|lower",
			"users.*.name": "trim",
		}, nil)
		require.NoError(t, err)

		val, _ := bag.Get("title")
		assert.Equal(t, "admin", val)
		val, _ = bag.Get("users.0.name")
		assert.Equal(t, "Alice", val)
	})

	t.Run("wildcard filter with custom filter", func(t *testing.T) {
		bag, err := NewDataBag(map[string]any{
			"items": []any{
				map[string]any{"code": "abc"},
				map[string]any{"code": "def"},
			},
		})
		require.NoError(t, err)

		customFilter := &mockFilter{
			signature: "prefix_item",
			handler: func(val string) string {
				return "item_" + val
			},
		}

		err = applyFilters(context.Background(), bag, map[string]any{
			"items.*.code": "prefix_item",
		}, []contractsvalidation.Filter{customFilter})
		require.NoError(t, err)

		val, _ := bag.Get("items.0.code")
		assert.Equal(t, "item_abc", val)
		val, _ = bag.Get("items.1.code")
		assert.Equal(t, "item_def", val)
	})
}

func TestAddFilters(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		validation := NewValidation()
		err := validation.AddFilters([]contractsvalidation.Filter{
			&mockFilter{signature: "test_filter"},
		})
		assert.NoError(t, err)
		assert.Len(t, validation.Filters(), 1)
	})

	t.Run("duplicate filter", func(t *testing.T) {
		validation := NewValidation()
		err := validation.AddFilters([]contractsvalidation.Filter{
			&mockFilter{signature: "test_filter"},
		})
		assert.NoError(t, err)

		err = validation.AddFilters([]contractsvalidation.Filter{
			&mockFilter{signature: "test_filter"},
		})
		assert.Error(t, err)
	})

	t.Run("duplicate builtin filter", func(t *testing.T) {
		validation := NewValidation()
		err := validation.AddFilters([]contractsvalidation.Filter{
			&mockFilter{signature: "trim"},
		})
		assert.Error(t, err)
	})
}

// mockFilter is a test helper that implements the Filter interface.
type mockFilter struct {
	signature string
	handler   any
}

func (m *mockFilter) Signature() string {
	return m.signature
}

func (m *mockFilter) Handle(_ context.Context) any {
	return m.handler
}
