package validation

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDataBag(t *testing.T) {
	t.Run("from map[string]any", func(t *testing.T) {
		bag, err := NewDataBag(map[string]any{"name": "Alice", "age": 30})
		require.NoError(t, err)
		val, ok := bag.Get("name")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)
	})

	t.Run("from url.Values", func(t *testing.T) {
		vals := url.Values{"name": {"Alice"}, "tags": {"a", "b"}}
		bag, err := NewDataBag(vals)
		require.NoError(t, err)

		val, ok := bag.Get("name")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)

		val, ok = bag.Get("tags")
		assert.True(t, ok)
		assert.Equal(t, []any{"a", "b"}, val)
	})

	t.Run("from map[string][]string", func(t *testing.T) {
		bag, err := NewDataBag(map[string][]string{"name": {"Bob"}})
		require.NoError(t, err)
		val, ok := bag.Get("name")
		assert.True(t, ok)
		assert.Equal(t, "Bob", val)
	})

	t.Run("from struct", func(t *testing.T) {
		type User struct {
			Name  string `form:"name"`
			Email string `json:"email"`
		}
		bag, err := NewDataBag(&User{Name: "Alice", Email: "alice@example.com"})
		require.NoError(t, err)

		val, ok := bag.Get("name")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)

		val, ok = bag.Get("email")
		assert.True(t, ok)
		assert.Equal(t, "alice@example.com", val)
	})

	t.Run("nil returns error", func(t *testing.T) {
		bag, err := NewDataBag(nil)
		assert.Nil(t, bag)
		assert.Error(t, err)
	})

	t.Run("unsupported type returns error", func(t *testing.T) {
		bag, err := NewDataBag(42)
		assert.Nil(t, bag)
		assert.Error(t, err)
	})
}

func TestNewDataBagFromRequest(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		bag, err := NewDataBagFromRequest(nil, 0)
		assert.Nil(t, bag)
		assert.Error(t, err)
	})

	t.Run("JSON body", func(t *testing.T) {
		body := `{"name":"Alice","age":30}`
		req, _ := http.NewRequest("POST", "http://example.com?q=search", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		bag, err := NewDataBagFromRequest(req, 0)
		require.NoError(t, err)

		// JSON body fields
		val, ok := bag.Get("name")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)

		// Query params
		val, ok = bag.Get("q")
		assert.True(t, ok)
		assert.Equal(t, "search", val)
	})

	t.Run("JSON body overrides query params", func(t *testing.T) {
		body := `{"name":"BodyName"}`
		req, _ := http.NewRequest("POST", "http://example.com?name=QueryName", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		bag, err := NewDataBagFromRequest(req, 0)
		require.NoError(t, err)

		val, ok := bag.Get("name")
		assert.True(t, ok)
		assert.Equal(t, "BodyName", val)
	})

	t.Run("form-urlencoded body", func(t *testing.T) {
		form := url.Values{"name": {"Alice"}, "tags": {"a", "b"}}
		req, _ := http.NewRequest("POST", "http://example.com", bytes.NewBufferString(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		bag, err := NewDataBagFromRequest(req, 0)
		require.NoError(t, err)

		val, ok := bag.Get("name")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)

		val, ok = bag.Get("tags")
		assert.True(t, ok)
		assert.Equal(t, []any{"a", "b"}, val)
	})

	t.Run("multipart form data", func(t *testing.T) {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		_ = writer.WriteField("name", "Alice")
		_ = writer.WriteField("city", "Beijing")
		fw, _ := writer.CreateFormFile("avatar", "test.png")
		_, _ = fw.Write([]byte("fake-image"))
		require.NoError(t, writer.Close())

		req, _ := http.NewRequest("POST", "http://example.com", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		bag, err := NewDataBagFromRequest(req, 0)
		require.NoError(t, err)

		val, ok := bag.Get("name")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)

		// File should be present
		val, ok = bag.Get("avatar")
		assert.True(t, ok)
		assert.IsType(t, &multipart.FileHeader{}, val)
	})

	t.Run("graceful degradation on invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://example.com?q=ok", bytes.NewBufferString("not-json"))
		req.Header.Set("Content-Type", "application/json")

		bag, err := NewDataBagFromRequest(req, 0)
		require.NoError(t, err)

		// Query params should still be available
		val, ok := bag.Get("q")
		assert.True(t, ok)
		assert.Equal(t, "ok", val)
	})

	t.Run("graceful degradation on nil body", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://example.com?q=ok", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Body = nil

		bag, err := NewDataBagFromRequest(req, 0)
		require.NoError(t, err)

		val, ok := bag.Get("q")
		assert.True(t, ok)
		assert.Equal(t, "ok", val)
	})

	t.Run("graceful degradation on read error", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "http://example.com?q=ok", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Body = io.NopCloser(&errorReader{})

		bag, err := NewDataBagFromRequest(req, 0)
		require.NoError(t, err)

		val, ok := bag.Get("q")
		assert.True(t, ok)
		assert.Equal(t, "ok", val)
	})

	t.Run("query only when no content type", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com?page=1&limit=10", nil)

		bag, err := NewDataBagFromRequest(req, 0)
		require.NoError(t, err)

		val, ok := bag.Get("page")
		assert.True(t, ok)
		assert.Equal(t, "1", val)
	})
}

type errorReader struct{}

func (e *errorReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestDataBag_Get(t *testing.T) {
	t.Run("simple key", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"name": "Alice"})
		val, ok := bag.Get("name")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)
	})

	t.Run("missing key", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"name": "Alice"})
		val, ok := bag.Get("missing")
		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("empty key", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"name": "Alice"})
		val, ok := bag.Get("")
		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("dot notation", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{
			"user": map[string]any{
				"name": "Alice",
				"address": map[string]any{
					"city": "Beijing",
				},
			},
		})

		val, ok := bag.Get("user.name")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)

		val, ok = bag.Get("user.address.city")
		assert.True(t, ok)
		assert.Equal(t, "Beijing", val)
	})

	t.Run("array index", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{
			"users": []any{
				map[string]any{"name": "Alice"},
				map[string]any{"name": "Bob"},
			},
		})

		val, ok := bag.Get("users.0.name")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)

		val, ok = bag.Get("users.1.name")
		assert.True(t, ok)
		assert.Equal(t, "Bob", val)

		// Out of range
		_, ok = bag.Get("users.5.name")
		assert.False(t, ok)
	})
}

func TestDataBag_Set(t *testing.T) {
	t.Run("simple key", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{})
		err := bag.Set("name", "Alice")
		require.NoError(t, err)

		val, ok := bag.Get("name")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)
	})

	t.Run("empty key returns error", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{})
		err := bag.Set("", "Alice")
		assert.Error(t, err)
	})

	t.Run("dot notation creates nested maps", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{})
		err := bag.Set("user.name", "Alice")
		require.NoError(t, err)

		val, ok := bag.Get("user.name")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)
	})

	t.Run("invalidates cached keys", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"a": 1})
		keys1 := bag.Keys()
		assert.Contains(t, keys1, "a")

		_ = bag.Set("b", 2)
		keys2 := bag.Keys()
		assert.Contains(t, keys2, "b")
	})
}

func TestDataBag_Has(t *testing.T) {
	bag, _ := NewDataBag(map[string]any{"name": "Alice"})
	assert.True(t, bag.Has("name"))
	assert.False(t, bag.Has("missing"))
}

func TestDataBag_All(t *testing.T) {
	data := map[string]any{"name": "Alice", "age": 30}
	bag, _ := NewDataBag(data)
	assert.Equal(t, data, bag.All())
}

func TestDataBag_Keys(t *testing.T) {
	t.Run("flat data", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"a": 1, "b": 2})
		keys := bag.Keys()
		assert.Contains(t, keys, "a")
		assert.Contains(t, keys, "b")
	})

	t.Run("nested data", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{
			"user": map[string]any{"name": "Alice"},
			"tags": []any{"a", "b"},
		})
		keys := bag.Keys()
		assert.Contains(t, keys, "user")
		assert.Contains(t, keys, "user.name")
		assert.Contains(t, keys, "tags")
		assert.Contains(t, keys, "tags.0")
		assert.Contains(t, keys, "tags.1")
	})

	t.Run("keys are sorted", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"c": 3, "a": 1, "b": 2})
		keys := bag.Keys()
		assert.True(t, len(keys) == 3)
		assert.Equal(t, "a", keys[0])
		assert.Equal(t, "b", keys[1])
		assert.Equal(t, "c", keys[2])
	})

	t.Run("results are cached", func(t *testing.T) {
		bag, _ := NewDataBag(map[string]any{"a": 1})
		keys1 := bag.Keys()
		keys2 := bag.Keys()
		assert.Equal(t, keys1, keys2)
	})
}
