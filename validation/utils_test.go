package validation

import (
	"context"
	"mime/multipart"
	"net/url"
	"reflect"
	"testing"
	"time"

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

	t.Run("nil intermediate does not panic", func(t *testing.T) {
		data := map[string]any{"user": nil}
		_, ok := dotGet(data, []string{"user", "name"})
		assert.False(t, ok)
	})

	t.Run("nil slice element does not panic", func(t *testing.T) {
		data := map[string]any{"items": []any{nil, "b"}}
		_, ok := dotGet(data, []string{"items", "0", "key"})
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

func TestSetValidated(t *testing.T) {
	t.Run("empty segments does nothing", func(t *testing.T) {
		data := map[string]any{"name": "Alice"}
		setValidated(data, map[string]any{"name": "Alice"}, []string{}, "Bob")
		assert.Equal(t, map[string]any{"name": "Alice"}, data)
	})

	t.Run("creates slice containers based on source shape", func(t *testing.T) {
		data := map[string]any{}
		source := map[string]any{
			"tags": []any{"a", "b"},
		}

		setValidated(data, source, []string{"tags", "1"}, "B")
		normalized := normalizeValidatedShape(data, source).(map[string]any)

		tags, ok := normalized["tags"].([]any)
		assert.True(t, ok)
		assert.Equal(t, []any{nil, "B"}, tags)
	})

	t.Run("preserves map container for numeric map key", func(t *testing.T) {
		data := map[string]any{}
		source := map[string]any{
			"meta": map[string]any{"0": "x"},
		}

		setValidated(data, source, []string{"meta", "0"}, "value")

		meta, ok := data["meta"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "value", meta["0"])
	})

	t.Run("normalizes to typed slice when conversion is safe", func(t *testing.T) {
		data := map[string]any{}
		source := map[string]any{
			"scores": []int{1, 2},
		}

		setValidated(data, source, []string{"scores", "0"}, 1)
		setValidated(data, source, []string{"scores", "1"}, 2)

		normalized := normalizeValidatedShape(data, source).(map[string]any)
		scores, ok := normalized["scores"].([]int)
		assert.True(t, ok)
		assert.Equal(t, []int{1, 2}, scores)
	})

	t.Run("replaces incompatible existing container with map", func(t *testing.T) {
		data := map[string]any{"meta": []any{"wrong"}}
		source := map[string]any{
			"meta": map[string]any{},
		}

		setValidated(data, source, []string{"meta", "name"}, "goravel")

		meta, ok := data["meta"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "goravel", meta["name"])
	})
}

func TestSetValidatedOnSlice(t *testing.T) {
	t.Run("empty segments returns current", func(t *testing.T) {
		current := []any{"a"}
		result := setValidatedOnSlice(current, nil, []string{}, "b")
		assert.Equal(t, []any{"a"}, result)
	})

	t.Run("invalid index leaves current unchanged", func(t *testing.T) {
		current := []any{"a"}
		result := setValidatedOnSlice(current, nil, []string{"x"}, "b")
		assert.Equal(t, []any{"a"}, result)
	})

	t.Run("single segment sets value", func(t *testing.T) {
		current := []any{"a"}
		result := setValidatedOnSlice(current, nil, []string{"0"}, "b")
		assert.Equal(t, []any{"b"}, result)
	})

	t.Run("creates nested map when source is not slice", func(t *testing.T) {
		current := []any{}
		source := map[string]any{"name": "a"}

		result := setValidatedOnSlice(current, source, []string{"0", "name"}, "b")
		item, ok := result[0].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "b", item["name"])
	})

	t.Run("creates nested slice when source child is slice", func(t *testing.T) {
		current := []any{}
		source := []any{
			[]int{1, 2},
		}

		result := setValidatedOnSlice(current, source, []string{"0", "1"}, 99)
		nested, ok := result[0].([]any)
		assert.True(t, ok)
		assert.Equal(t, []any{nil, 99}, nested)
	})
}

func TestHelpersForValidatedShape(t *testing.T) {
	t.Run("isExpectedContainer", func(t *testing.T) {
		assert.True(t, isExpectedContainer([]any{"a"}, true))
		assert.True(t, isExpectedContainer(map[string]any{"a": 1}, false))
		assert.False(t, isExpectedContainer(map[string]any{"a": 1}, true))
		assert.False(t, isExpectedContainer([]any{"a"}, false))
	})

	t.Run("isIndexSegment", func(t *testing.T) {
		assert.True(t, isIndexSegment("0"))
		assert.False(t, isIndexSegment("-1"))
		assert.False(t, isIndexSegment("abc"))
	})

	t.Run("isSliceOrArray", func(t *testing.T) {
		assert.True(t, isSliceOrArray([]int{1}))
		assert.True(t, isSliceOrArray([1]int{1}))
		assert.False(t, isSliceOrArray(nil))
		assert.False(t, isSliceOrArray("x"))
	})

	t.Run("ensureAnySliceLen", func(t *testing.T) {
		assert.Equal(t, []any{"a", nil, nil}, ensureAnySliceLen([]any{"a"}, 3))
		assert.Equal(t, []any{"a"}, ensureAnySliceLen([]any{"a"}, 1))
	})

	t.Run("toAnySlice", func(t *testing.T) {
		v, ok := toAnySlice([]any{"a"})
		assert.True(t, ok)
		assert.Equal(t, []any{"a"}, v)

		v, ok = toAnySlice([]int{1, 2})
		assert.True(t, ok)
		assert.Equal(t, []any{1, 2}, v)

		v, ok = toAnySlice([2]int{3, 4})
		assert.True(t, ok)
		assert.Equal(t, []any{3, 4}, v)

		_, ok = toAnySlice(nil)
		assert.False(t, ok)

		_, ok = toAnySlice("not-slice")
		assert.False(t, ok)
	})

	t.Run("getValidatedChild", func(t *testing.T) {
		_, ok := getValidatedChild(nil, "name")
		assert.False(t, ok)

		child, ok := getValidatedChild(map[string]any{"name": "goravel"}, "name")
		assert.True(t, ok)
		assert.Equal(t, "goravel", child)

		_, ok = getValidatedChild(map[string]any{"name": "goravel"}, "missing")
		assert.False(t, ok)

		child, ok = getValidatedChild([]any{"a", "b"}, "1")
		assert.True(t, ok)
		assert.Equal(t, "b", child)

		_, ok = getValidatedChild([]any{"a", "b"}, "x")
		assert.False(t, ok)

		_, ok = getValidatedChild("plain", "0")
		assert.False(t, ok)
	})
}

func TestConvertAnySliceToSourceType(t *testing.T) {
	t.Run("returns original for nil source", func(t *testing.T) {
		data := []any{1}
		assert.Equal(t, data, convertAnySliceToSourceType(data, nil))
	})

	t.Run("returns original for non-slice source", func(t *testing.T) {
		data := []any{1}
		assert.Equal(t, data, convertAnySliceToSourceType(data, "x"))
	})

	t.Run("converts to typed slice when assignable", func(t *testing.T) {
		data := []any{1, 2}
		result := convertAnySliceToSourceType(data, []int{0})
		assert.Equal(t, []int{1, 2}, result)
	})

	t.Run("converts to typed slice when convertible", func(t *testing.T) {
		data := []any{int32(1), int32(2)}
		result := convertAnySliceToSourceType(data, []int{0})
		assert.Equal(t, []int{1, 2}, result)
	})

	t.Run("returns original when nil element cannot be represented", func(t *testing.T) {
		data := []any{nil}
		assert.Equal(t, data, convertAnySliceToSourceType(data, []int{0}))
	})

	t.Run("keeps nil for nil-able element type", func(t *testing.T) {
		data := []any{nil}
		result := convertAnySliceToSourceType(data, []*int{})
		typed, ok := result.([]*int)
		assert.True(t, ok)
		assert.Len(t, typed, 1)
		assert.Nil(t, typed[0])
	})

	t.Run("returns original when element type mismatches", func(t *testing.T) {
		data := []any{"x"}
		assert.Equal(t, data, convertAnySliceToSourceType(data, []int{0}))
	})

	t.Run("array source converts to slice type", func(t *testing.T) {
		data := []any{1, 2}
		result := convertAnySliceToSourceType(data, [2]int{})
		assert.Equal(t, []int{1, 2}, result)
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

	t.Run("nil map value does not panic", func(t *testing.T) {
		data := map[string]any{"name": nil}
		var keys []string
		assert.NotPanics(t, func() { collectKeys(data, "", &keys) })
		assert.Contains(t, keys, "name")
	})

	t.Run("nil slice element does not panic", func(t *testing.T) {
		data := []any{nil, "b"}
		var keys []string
		assert.NotPanics(t, func() { collectKeys(data, "items", &keys) })
		assert.Contains(t, keys, "items.0")
		assert.Contains(t, keys, "items.1")
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

func TestGetSize(t *testing.T) {
	t.Run("numeric", func(t *testing.T) {
		size, ok := getSize(42, "numeric")
		assert.True(t, ok)
		assert.Equal(t, float64(42), size)

		size, ok = getSize(3.14, "numeric")
		assert.True(t, ok)
		assert.Equal(t, 3.14, size)

		size, ok = getSize("100", "numeric")
		assert.True(t, ok)
		assert.Equal(t, float64(100), size)
	})

	t.Run("string", func(t *testing.T) {
		size, ok := getSize("hello", "string")
		assert.True(t, ok)
		assert.Equal(t, float64(5), size)

		size, ok = getSize("你好", "string")
		assert.True(t, ok)
		assert.Equal(t, float64(2), size)
	})

	t.Run("array", func(t *testing.T) {
		size, ok := getSize([]any{1, 2, 3}, "array")
		assert.True(t, ok)
		assert.Equal(t, float64(3), size)

		size, ok = getSize(map[string]any{"a": 1, "b": 2}, "array")
		assert.True(t, ok)
		assert.Equal(t, float64(2), size)

		_, ok = getSize(nil, "array")
		assert.False(t, ok)

		_, ok = getSize("not-array", "array")
		assert.False(t, ok)
	})

	t.Run("file", func(t *testing.T) {
		size, ok := getSize(&multipart.FileHeader{Size: 2048}, "file")
		assert.True(t, ok)
		assert.Equal(t, float64(2), size)

		size, ok = getSize([]*multipart.FileHeader{
			{Size: 1024},
			{Size: 3072},
		}, "file")
		assert.True(t, ok)
		assert.Equal(t, float64(4), size)

		_, ok = getSize("not-file", "file")
		assert.False(t, ok)
	})
}

func TestParseDateValue(t *testing.T) {
	bag, _ := NewDataBag(map[string]any{
		"start": "2024-01-15",
	})

	t.Run("field reference", func(t *testing.T) {
		dt, ok := parseDateValue("start", bag)
		assert.True(t, ok)
		assert.Equal(t, 2024, dt.Year())
		assert.Equal(t, time.January, dt.Month())
		assert.Equal(t, 15, dt.Day())
	})

	t.Run("literal date string", func(t *testing.T) {
		dt, ok := parseDateValue("2023-06-01", bag)
		assert.True(t, ok)
		assert.Equal(t, 2023, dt.Year())
	})

	t.Run("invalid date", func(t *testing.T) {
		_, ok := parseDateValue("not-a-date", bag)
		assert.False(t, ok)
	})

	t.Run("missing field falls back to literal", func(t *testing.T) {
		_, ok := parseDateValue("missing_field", bag)
		assert.False(t, ok)
	})
}

func TestParseDate(t *testing.T) {
	t.Run("time.Time value", func(t *testing.T) {
		now := time.Now()
		dt, ok := parseDate(now)
		assert.True(t, ok)
		assert.Equal(t, now, dt)
	})

	t.Run("RFC3339 string", func(t *testing.T) {
		dt, ok := parseDate("2024-01-15T10:30:00Z")
		assert.True(t, ok)
		assert.Equal(t, 2024, dt.Year())
	})

	t.Run("date only string", func(t *testing.T) {
		dt, ok := parseDate("2024-01-15")
		assert.True(t, ok)
		assert.Equal(t, 15, dt.Day())
	})

	t.Run("datetime string", func(t *testing.T) {
		dt, ok := parseDate("2024-01-15 10:30:00")
		assert.True(t, ok)
		assert.Equal(t, 10, dt.Hour())
	})

	t.Run("invalid string", func(t *testing.T) {
		_, ok := parseDate("not-a-date")
		assert.False(t, ok)
	})

	t.Run("non-string non-time", func(t *testing.T) {
		_, ok := parseDate(12345)
		assert.False(t, ok)
	})
}

func TestIsAcceptedValue(t *testing.T) {
	tests := []struct {
		name     string
		val      any
		expected bool
	}{
		{"nil", nil, false},
		{"string yes", "yes", true},
		{"string on", "on", true},
		{"string 1", "1", true},
		{"string true", "true", true},
		{"string YES", "YES", true},
		{"string True", "True", true},
		{"string  yes ", " yes ", true},
		{"string no", "no", false},
		{"string false", "false", false},
		{"string empty", "", false},
		{"string random", "hello", false},
		{"string 2", "2", false},
		{"bool true", true, true},
		{"bool false", false, false},
		{"int 1", 1, true},
		{"int 0", 0, false},
		{"int 2", 2, false},
		{"int -1", -1, false},
		{"int64 1", int64(1), true},
		{"int64 0", int64(0), false},
		{"float64 1", float64(1), true},
		{"float64 0", float64(0), false},
		{"float64 0.5", float64(0.5), false},
		{"slice", []any{1, 2}, false},
		{"map", map[string]any{"a": 1}, false},
		{"struct", struct{ Name string }{"test"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isAcceptedValue(tt.val))
		})
	}
}

func TestIsDeclinedValue(t *testing.T) {
	tests := []struct {
		name     string
		val      any
		expected bool
	}{
		{"nil", nil, false},
		{"string no", "no", true},
		{"string off", "off", true},
		{"string 0", "0", true},
		{"string false", "false", true},
		{"string NO", "NO", true},
		{"string False", "False", true},
		{"string  no ", " no ", true},
		{"string yes", "yes", false},
		{"string true", "true", false},
		{"string empty", "", false},
		{"string random", "hello", false},
		{"string 2", "2", false},
		{"bool false", false, true},
		{"bool true", true, false},
		{"int 0", 0, true},
		{"int 1", 1, false},
		{"int 2", 2, false},
		{"int -1", -1, false},
		{"int64 0", int64(0), true},
		{"int64 1", int64(1), false},
		{"float64 0", float64(0), true},
		{"float64 1", float64(1), false},
		{"float64 0.5", 0.5, false},
		{"slice", []any{1, 2}, false},
		{"map", map[string]any{"a": 1}, false},
		{"struct", struct{ Name string }{"test"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isDeclinedValue(tt.val))
		})
	}
}

func TestParseDependentValues(t *testing.T) {
	bag, _ := NewDataBag(map[string]any{"status": "active"})

	t.Run("no parameters", func(t *testing.T) {
		ctx := &RuleContext{Ctx: context.Background(), Data: bag, Parameters: []string{}}
		otherValue, comparisonValues, otherField := parseDependentValues(ctx)
		assert.Nil(t, otherValue)
		assert.Nil(t, comparisonValues)
		assert.Empty(t, otherField)
	})

	t.Run("with field and values", func(t *testing.T) {
		ctx := &RuleContext{Ctx: context.Background(), Data: bag, Parameters: []string{"status", "active", "pending"}}
		otherValue, comparisonValues, otherField := parseDependentValues(ctx)
		assert.Equal(t, "active", otherValue)
		assert.Equal(t, []string{"active", "pending"}, comparisonValues)
		assert.Equal(t, "status", otherField)
	})

	t.Run("field only no comparison values", func(t *testing.T) {
		ctx := &RuleContext{Ctx: context.Background(), Data: bag, Parameters: []string{"status"}}
		otherValue, comparisonValues, otherField := parseDependentValues(ctx)
		assert.Equal(t, "active", otherValue)
		assert.Empty(t, comparisonValues)
		assert.Equal(t, "status", otherField)
	})

	t.Run("missing field", func(t *testing.T) {
		ctx := &RuleContext{Ctx: context.Background(), Data: bag, Parameters: []string{"missing", "val"}}
		otherValue, comparisonValues, _ := parseDependentValues(ctx)
		assert.Nil(t, otherValue)
		assert.Equal(t, []string{"val"}, comparisonValues)
	})
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "helloWorld"},
		{"hello-world", "helloWorld"},
		{"hello world", "helloWorld"},
		{"HelloWorld", "helloWorld"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, toCamelCase(tt.input))
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"helloWorld", "hello_world"},
		{"hello-world", "hello_world"},
		{"hello world", "hello_world"},
		{"HelloWorld", "hello_world"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, toSnakeCase(tt.input))
		})
	}
}

func TestSplitWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"underscore", "hello_world", []string{"hello", "world"}},
		{"dash", "hello-world", []string{"hello", "world"}},
		{"space", "hello world", []string{"hello", "world"}},
		{"camelCase", "helloWorld", []string{"hello", "World"}},
		{"PascalCase", "HelloWorld", []string{"Hello", "World"}},
		{"empty", "", nil},
		{"single word", "hello", []string{"hello"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, splitWords(tt.input))
		})
	}
}

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<p>Hello</p>", "Hello"},
		{"<b>Bold</b> and <i>italic</i>", "Bold and italic"},
		{"No tags here", "No tags here"},
		{"<script>alert('xss')</script>", "alert('xss')"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, stripHTMLTags(tt.input))
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal", "photo.jpg", "jpg"},
		{"multiple dots", "archive.tar.gz", "gz"},
		{"no extension", "README", ""},
		{"dot only", ".", ""},
		{"hidden file", ".gitignore", "gitignore"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getFileExtension(tt.input))
		})
	}
}

func TestStrToInts(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3}, strToInts("1,2,3"))
	assert.Equal(t, []int{1, 0, 3}, strToInts("1, nope, 3"))
	assert.Equal(t, []int{}, strToInts(" , , "))
}

func TestStrToArray(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, strToArray("a,b,c"))
	assert.Equal(t, []string{"a", "b"}, strToArray(" a , , b "))
	assert.Equal(t, []string{}, strToArray(" , "))
}
