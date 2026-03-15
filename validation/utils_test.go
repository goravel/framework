package validation

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValueEmpty(t *testing.T) {
	tests := []struct {
		name     string
		val      any
		expected bool
	}{
		{"nil", nil, true},
		{"empty string", "", true},
		{"whitespace string", "   ", true},
		{"non-empty string", "hello", false},
		{"zero int", 0, false},
		{"positive int", 42, false},
		{"false bool", false, false},
		{"empty slice", []any{}, true},
		{"non-empty slice", []any{1}, false},
		{"empty typed slice", []string{}, true},
		{"non-empty typed slice", []string{"a"}, false},
		{"empty map", map[string]any{}, true},
		{"non-empty map", map[string]any{"a": 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isValueEmpty(tt.val))
		})
	}
}

func TestGetAttributeType(t *testing.T) {
	t.Run("numeric from runtime value", func(t *testing.T) {
		assert.Equal(t, "numeric", getAttributeType("age", 42, nil))
		assert.Equal(t, "numeric", getAttributeType("price", 3.14, nil))
		assert.Equal(t, "numeric", getAttributeType("count", int64(100), nil))
	})

	t.Run("array from runtime value", func(t *testing.T) {
		assert.Equal(t, "array", getAttributeType("items", []any{1, 2}, nil))
		assert.Equal(t, "array", getAttributeType("data", map[string]any{}, nil))
	})

	t.Run("string fallback", func(t *testing.T) {
		assert.Equal(t, "string", getAttributeType("name", "hello", nil))
		assert.Equal(t, "string", getAttributeType("name", nil, nil))
	})
}

func TestMatchesOtherValue(t *testing.T) {
	t.Run("string match", func(t *testing.T) {
		assert.True(t, matchesOtherValue("yes", []string{"yes", "no"}))
		assert.False(t, matchesOtherValue("maybe", []string{"yes", "no"}))
	})

	t.Run("int match via Sprint", func(t *testing.T) {
		assert.True(t, matchesOtherValue(42, []string{"42"}))
		assert.False(t, matchesOtherValue(42, []string{"43"}))
	})

	t.Run("bool true match", func(t *testing.T) {
		assert.True(t, matchesOtherValue(true, []string{"true"}))
		assert.True(t, matchesOtherValue(true, []string{"1"}))
		assert.False(t, matchesOtherValue(true, []string{"false"}))
	})

	t.Run("bool false match", func(t *testing.T) {
		assert.True(t, matchesOtherValue(false, []string{"false"}))
		assert.True(t, matchesOtherValue(false, []string{"0"}))
		assert.False(t, matchesOtherValue(false, []string{"true"}))
	})
}

func TestIsControlRule(t *testing.T) {
	assert.True(t, isControlRule("bail"))
	assert.True(t, isControlRule("nullable"))
	assert.True(t, isControlRule("sometimes"))
	assert.False(t, isControlRule("required"))
	assert.False(t, isControlRule("string"))
}

func TestDotGet(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"name": "Alice",
			"addresses": []any{
				map[string]any{"city": "Beijing"},
				map[string]any{"city": "Shanghai"},
			},
		},
		"tags": []any{"go", "php"},
	}

	t.Run("empty segments returns data", func(t *testing.T) {
		val, ok := dotGet(data, []string{})
		assert.True(t, ok)
		assert.Equal(t, data, val)
	})

	t.Run("nested map", func(t *testing.T) {
		val, ok := dotGet(data, []string{"user", "name"})
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)
	})

	t.Run("array index", func(t *testing.T) {
		val, ok := dotGet(data, []string{"tags", "0"})
		assert.True(t, ok)
		assert.Equal(t, "go", val)
	})

	t.Run("nested array of maps", func(t *testing.T) {
		val, ok := dotGet(data, []string{"user", "addresses", "1", "city"})
		assert.True(t, ok)
		assert.Equal(t, "Shanghai", val)
	})

	t.Run("missing key", func(t *testing.T) {
		_, ok := dotGet(data, []string{"user", "missing"})
		assert.False(t, ok)
	})

	t.Run("invalid array index", func(t *testing.T) {
		_, ok := dotGet(data, []string{"tags", "abc"})
		assert.False(t, ok)
	})

	t.Run("out of range array index", func(t *testing.T) {
		_, ok := dotGet(data, []string{"tags", "99"})
		assert.False(t, ok)
	})

	t.Run("[]map[string]any type", func(t *testing.T) {
		data := map[string]any{
			"items": []map[string]any{
				{"id": 1},
				{"id": 2},
			},
		}
		val, ok := dotGet(data, []string{"items", "0", "id"})
		assert.True(t, ok)
		assert.Equal(t, 1, val)
	})

	t.Run("unsupported type returns false", func(t *testing.T) {
		_, ok := dotGet("string", []string{"key"})
		assert.False(t, ok)
	})
}

func TestDotSet(t *testing.T) {
	t.Run("single segment", func(t *testing.T) {
		data := map[string]any{}
		dotSet(data, []string{"name"}, "Alice")
		assert.Equal(t, "Alice", data["name"])
	})

	t.Run("nested creates intermediate maps", func(t *testing.T) {
		data := map[string]any{}
		dotSet(data, []string{"user", "name"}, "Alice")
		user, ok := data["user"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "Alice", user["name"])
	})

	t.Run("into existing slice", func(t *testing.T) {
		data := map[string]any{
			"items": []any{"a", "b", "c"},
		}
		dotSet(data, []string{"items", "1"}, "B")
		items := data["items"].([]any)
		assert.Equal(t, "B", items[1])
	})

	t.Run("into nested slice of maps", func(t *testing.T) {
		data := map[string]any{
			"items": []any{
				map[string]any{"name": "old"},
			},
		}
		dotSet(data, []string{"items", "0", "name"}, "new")
		items := data["items"].([]any)
		item := items[0].(map[string]any)
		assert.Equal(t, "new", item["name"])
	})

	t.Run("empty segments does nothing", func(t *testing.T) {
		data := map[string]any{"a": 1}
		dotSet(data, []string{}, "val")
		assert.Equal(t, map[string]any{"a": 1}, data)
	})

	t.Run("overwrite non-map with new nested path", func(t *testing.T) {
		data := map[string]any{"user": "old-string"}
		dotSet(data, []string{"user", "name"}, "Alice")
		user, ok := data["user"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "Alice", user["name"])
	})
}

func TestCollectKeys(t *testing.T) {
	t.Run("flat map", func(t *testing.T) {
		data := map[string]any{"a": 1, "b": 2}
		var keys []string
		collectKeys(data, "", &keys)
		assert.Contains(t, keys, "a")
		assert.Contains(t, keys, "b")
	})

	t.Run("nested map", func(t *testing.T) {
		data := map[string]any{
			"user": map[string]any{"name": "Alice"},
		}
		var keys []string
		collectKeys(data, "", &keys)
		assert.Contains(t, keys, "user")
		assert.Contains(t, keys, "user.name")
	})

	t.Run("slice", func(t *testing.T) {
		data := map[string]any{
			"tags": []any{"a", "b"},
		}
		var keys []string
		collectKeys(data, "", &keys)
		assert.Contains(t, keys, "tags")
		assert.Contains(t, keys, "tags.0")
		assert.Contains(t, keys, "tags.1")
	})

	t.Run("[]map[string]any", func(t *testing.T) {
		data := map[string]any{
			"items": []map[string]any{
				{"id": 1},
			},
		}
		var keys []string
		collectKeys(data, "", &keys)
		assert.Contains(t, keys, "items")
		assert.Contains(t, keys, "items.0")
		assert.Contains(t, keys, "items.0.id")
	})

	t.Run("with prefix", func(t *testing.T) {
		data := []any{"x", "y"}
		var keys []string
		collectKeys(data, "arr", &keys)
		assert.Contains(t, keys, "arr.0")
		assert.Contains(t, keys, "arr.1")
	})
}

func TestExpandWildcardFields(t *testing.T) {
	t.Run("no wildcards", func(t *testing.T) {
		fields := map[string]string{"name": "required", "email": "required"}
		dataKeys := []string{"name", "email"}
		result := expandWildcardFields(fields, dataKeys, false)
		assert.Equal(t, fields, result)
	})

	t.Run("expand wildcard", func(t *testing.T) {
		fields := map[string]string{"users.*.name": "required"}
		dataKeys := []string{"users", "users.0", "users.0.name", "users.1", "users.1.name"}
		result := expandWildcardFields(fields, dataKeys, false)
		assert.Equal(t, "required", result["users.0.name"])
		assert.Equal(t, "required", result["users.1.name"])
		assert.Empty(t, result["users.*.name"])
	})

	t.Run("keep unmatched when flag is true", func(t *testing.T) {
		fields := map[string]string{"items.*.id": "required"}
		dataKeys := []string{"name"}
		result := expandWildcardFields(fields, dataKeys, true)
		assert.Equal(t, "required", result["items.*.id"])
	})

	t.Run("discard unmatched when flag is false", func(t *testing.T) {
		fields := map[string]string{"items.*.id": "required"}
		dataKeys := []string{"name"}
		result := expandWildcardFields(fields, dataKeys, false)
		assert.Empty(t, result)
	})
}

func TestUrlValuesToMap(t *testing.T) {
	t.Run("single values unwrapped", func(t *testing.T) {
		vals := url.Values{"name": {"Alice"}, "age": {"30"}}
		result := urlValuesToMap(vals)
		assert.Equal(t, "Alice", result["name"])
		assert.Equal(t, "30", result["age"])
	})

	t.Run("multiple values kept as slice", func(t *testing.T) {
		vals := url.Values{"tags": {"a", "b", "c"}}
		result := urlValuesToMap(vals)
		assert.Equal(t, []any{"a", "b", "c"}, result["tags"])
	})
}

func TestStructToMap(t *testing.T) {
	t.Run("basic struct with form tags", func(t *testing.T) {
		type User struct {
			Name  string `form:"name"`
			Email string `form:"email"`
		}
		rv := reflect.ValueOf(User{Name: "Alice", Email: "alice@example.com"})
		result := structToMap(rv)
		assert.Equal(t, "Alice", result["name"])
		assert.Equal(t, "alice@example.com", result["email"])
	})

	t.Run("falls back to json tag", func(t *testing.T) {
		type Data struct {
			Value string `json:"val"`
		}
		rv := reflect.ValueOf(Data{Value: "test"})
		result := structToMap(rv)
		assert.Equal(t, "test", result["val"])
	})

	t.Run("falls back to field name", func(t *testing.T) {
		type Data struct {
			Name string
		}
		rv := reflect.ValueOf(Data{Name: "test"})
		result := structToMap(rv)
		assert.Equal(t, "test", result["Name"])
	})

	t.Run("skips dash tags", func(t *testing.T) {
		type Data struct {
			Internal string `form:"-" json:"-"`
			Public   string `form:"public"`
		}
		rv := reflect.ValueOf(Data{Internal: "secret", Public: "visible"})
		result := structToMap(rv)
		assert.NotContains(t, result, "Internal")
		assert.NotContains(t, result, "-")
		assert.Equal(t, "visible", result["public"])
	})

	t.Run("handles omitempty tag option", func(t *testing.T) {
		type Data struct {
			Name string `form:"name,omitempty"`
		}
		rv := reflect.ValueOf(Data{Name: "Alice"})
		result := structToMap(rv)
		assert.Equal(t, "Alice", result["name"])
	})

	t.Run("embedded struct", func(t *testing.T) {
		type Base struct {
			ID int `form:"id"`
		}
		type Data struct {
			Base
			Name string `form:"name"`
		}
		rv := reflect.ValueOf(Data{Base: Base{ID: 1}, Name: "Alice"})
		result := structToMap(rv)
		assert.Equal(t, 1, result["id"])
		assert.Equal(t, "Alice", result["name"])
	})

	t.Run("nested struct normalized", func(t *testing.T) {
		type Address struct {
			City string `form:"city"`
		}
		type User struct {
			Name    string  `form:"name"`
			Address Address `form:"address"`
		}
		rv := reflect.ValueOf(User{Name: "Alice", Address: Address{City: "Beijing"}})
		result := structToMap(rv)
		assert.Equal(t, "Alice", result["name"])
		addr, ok := result["address"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "Beijing", addr["city"])
	})

	t.Run("unexported fields skipped", func(t *testing.T) {
		type Data struct {
			Public  string `form:"public"`
			private string //nolint:unused
		}
		rv := reflect.ValueOf(Data{Public: "yes"})
		result := structToMap(rv)
		assert.Equal(t, "yes", result["public"])
		assert.Len(t, result, 1)
	})
}

func TestNormalizeValue(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		rv := reflect.ValueOf("hello")
		assert.Equal(t, "hello", normalizeValue(rv))
	})

	t.Run("int", func(t *testing.T) {
		rv := reflect.ValueOf(42)
		assert.Equal(t, 42, normalizeValue(rv))
	})

	t.Run("nil pointer", func(t *testing.T) {
		var p *string
		rv := reflect.ValueOf(p)
		assert.Nil(t, normalizeValue(rv))
	})

	t.Run("slice", func(t *testing.T) {
		rv := reflect.ValueOf([]string{"a", "b"})
		result := normalizeValue(rv)
		assert.Equal(t, []any{"a", "b"}, result)
	})

	t.Run("map", func(t *testing.T) {
		rv := reflect.ValueOf(map[string]int{"a": 1})
		result := normalizeValue(rv).(map[string]any)
		assert.Equal(t, 1, result["a"])
	})

	t.Run("struct", func(t *testing.T) {
		type Data struct {
			Name string `form:"name"`
		}
		rv := reflect.ValueOf(Data{Name: "Alice"})
		result := normalizeValue(rv).(map[string]any)
		assert.Equal(t, "Alice", result["name"])
	})
}
