package validation

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"

	supportjson "github.com/goravel/framework/support/json"
)

// defaultMaxMultipartMemory is the default maximum memory (in bytes) used
// to buffer multipart form data in memory before spilling to disk.
const defaultMaxMultipartMemory = 32 << 20 // 32 MB

// DataBag provides a unified data abstraction for validation.
// It supports nested data access via dot notation (e.g., "user.name", "users.0.email").
type DataBag struct {
	data       map[string]any
	cachedKeys []string
}

// NewDataBag creates a DataBag from various input types:
// - map[string]any (supports nested maps/slices)
// - url.Values / map[string][]string (single values unwrapped from slices)
// - struct (reflected via "form" tag)
func NewDataBag(input any) (*DataBag, error) {
	if input == nil {
		return nil, fmt.Errorf("data cannot be nil")
	}

	switch v := input.(type) {
	case map[string]any:
		return &DataBag{data: v}, nil
	case url.Values:
		return &DataBag{data: urlValuesToMap(v)}, nil
	case map[string][]string:
		return &DataBag{data: urlValuesToMap(v)}, nil
	default:
		rv := reflect.Indirect(reflect.ValueOf(input))
		if rv.Kind() == reflect.Struct {
			m := structToMap(rv)
			return &DataBag{data: m}, nil
		}
		return nil, fmt.Errorf("unsupported data type: %T", input)
	}
}

// NewDataBagFromRequest parses an HTTP request into a DataBag.
// maxMemory controls the maximum memory (in bytes) used to buffer multipart
// form data before spilling to disk. If 0, defaults to 32 MB.
func NewDataBagFromRequest(r *http.Request, maxMemory int64) (*DataBag, error) {
	if r == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	bag := &DataBag{data: make(map[string]any)}

	// Parse query parameters first (body data takes priority)
	if r.URL != nil {
		for key, values := range r.URL.Query() {
			if len(values) == 1 {
				bag.data[key] = values[0]
			} else {
				bag.data[key] = values
			}
		}
	}

	contentType := r.Header.Get("Content-Type")

	switch {
	case strings.HasPrefix(contentType, "application/json"):
		if r.Body != nil {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				return bag, nil
			}
			// Restore body for potential re-reads
			r.Body = io.NopCloser(bytes.NewReader(body))
			var jsonData map[string]any
			if err = supportjson.Unmarshal(body, &jsonData); err == nil {
				// Body data overrides query data
				for k, v := range jsonData {
					bag.data[k] = v
				}
			}
		}
	case strings.HasPrefix(contentType, "multipart/form-data"):
		memory := int64(defaultMaxMultipartMemory)
		if maxMemory > 0 {
			memory = maxMemory
		}
		if err := r.ParseMultipartForm(memory); err == nil && r.MultipartForm != nil {
			for key, values := range r.MultipartForm.Value {
				if len(values) == 1 {
					bag.data[key] = values[0]
				} else {
					s := make([]any, len(values))
					for i, v := range values {
						s[i] = v
					}
					bag.data[key] = s
				}
			}
			for key, files := range r.MultipartForm.File {
				if len(files) == 1 {
					bag.data[key] = files[0]
				} else {
					bag.data[key] = files
				}
			}
		}
	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		if err := r.ParseForm(); err == nil {
			for key, values := range r.PostForm {
				if len(values) == 1 {
					bag.data[key] = values[0]
				} else {
					s := make([]any, len(values))
					for i, v := range values {
						s[i] = v
					}
					bag.data[key] = s
				}
			}
		}
	}

	return bag, nil
}

// Get retrieves a value using dot notation.
// Supports nested keys: "user.name", array indexes: "users.0.email", and wildcards.
func (d *DataBag) Get(key string) (any, bool) {
	if key == "" {
		return nil, false
	}

	// Fast path: no dot notation
	if !strings.Contains(key, ".") {
		val, ok := d.data[key]
		return val, ok
	}

	return dotGet(d.data, strings.Split(key, "."))
}

// Set sets a value using dot notation.
func (d *DataBag) Set(key string, val any) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	d.cachedKeys = nil // invalidate cache

	// Fast path: no dot notation
	if !strings.Contains(key, ".") {
		d.data[key] = val
		return nil
	}

	dotSet(d.data, strings.Split(key, "."), val)
	return nil
}

// Has checks if a key exists in the data (including files).
func (d *DataBag) Has(key string) bool {
	_, exists := d.Get(key)
	return exists
}

// All returns all the data as a flat map.
func (d *DataBag) All() map[string]any {
	return d.data
}

// Keys returns all dot-notation paths in the data, for wildcard expansion.
// Results are cached and invalidated when Set() is called.
func (d *DataBag) Keys() []string {
	if d.cachedKeys != nil {
		return d.cachedKeys
	}
	keys := make([]string, 0)
	collectKeys(d.data, "", &keys)
	sort.Strings(keys)
	d.cachedKeys = keys
	return d.cachedKeys
}
