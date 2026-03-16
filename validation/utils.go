package validation

import (
	"fmt"
	"mime/multipart"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// isValueEmpty checks if a value is considered "empty" for validation purposes.
func isValueEmpty(val any) bool {
	if val == nil {
		return true
	}
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			return rv.Len() == 0
		default:
			return false
		}
	}
}

// getAttributeType determines the type of an attribute for size rules.
func getAttributeType(attribute string, value any, rules map[string][]ParsedRule) string {
	if fieldRules, ok := rules[attribute]; ok {
		for _, r := range fieldRules {
			if numericRuleNames[r.Name] {
				return "numeric"
			}
			if r.Name == "array" || r.Name == "list" || r.Name == "slice" || r.Name == "map" {
				return "array"
			}
		}
	}

	// Check if the value is a file
	if _, ok := value.(*multipart.FileHeader); ok {
		return "file"
	}
	if _, ok := value.([]*multipart.FileHeader); ok {
		return "file"
	}

	// Fallback: determine type from runtime value
	if value != nil {
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return "numeric"
		case reflect.Slice, reflect.Array, reflect.Map:
			return "array"
		}
	}

	return "string"
}

// matchesOtherValue checks if otherValue matches any of the comparison values.
func matchesOtherValue(otherValue any, comparisonValues []string) bool {
	otherStr := fmt.Sprintf("%v", otherValue)
	for _, cv := range comparisonValues {
		if otherStr == cv {
			return true
		}
	}
	// Handle boolean matching
	if b, ok := otherValue.(bool); ok {
		for _, cv := range comparisonValues {
			if (b && (cv == "true" || cv == "1")) || (!b && (cv == "false" || cv == "0")) {
				return true
			}
		}
	}
	return false
}

// isControlRule returns true for rule names that are control directives (not actual validation rules).
func isControlRule(name string) bool {
	return name == "bail" || name == "nullable" || name == "sometimes"
}

// dotGet navigates nested maps/slices using path segments.
func dotGet(data any, segments []string) (any, bool) {
	if len(segments) == 0 {
		return data, true
	}

	segment := segments[0]
	remaining := segments[1:]

	switch v := data.(type) {
	case map[string]any:
		val, ok := v[segment]
		if !ok {
			return nil, false
		}
		return dotGet(val, remaining)
	case []any:
		idx, err := strconv.Atoi(segment)
		if err != nil || idx < 0 || idx >= len(v) {
			return nil, false
		}
		return dotGet(v[idx], remaining)
	case []map[string]any:
		idx, err := strconv.Atoi(segment)
		if err != nil || idx < 0 || idx >= len(v) {
			return nil, false
		}
		return dotGet(v[idx], remaining)
	default:
		return nil, false
	}
}

// dotSet sets a value in nested maps/slices using path segments.
func dotSet(data map[string]any, segments []string, val any) {
	if len(segments) == 0 {
		return
	}

	if len(segments) == 1 {
		data[segments[0]] = val
		return
	}

	segment := segments[0]
	remaining := segments[1:]

	next, ok := data[segment]
	if !ok {
		nextMap := make(map[string]any)
		data[segment] = nextMap
		dotSet(nextMap, remaining, val)
		return
	}

	switch v := next.(type) {
	case map[string]any:
		dotSet(v, remaining, val)
	case []any:
		idx, err := strconv.Atoi(remaining[0])
		if err != nil || idx < 0 || idx >= len(v) {
			return
		}
		if len(remaining) == 1 {
			v[idx] = val
			return
		}
		if m, ok := v[idx].(map[string]any); ok {
			dotSet(m, remaining[1:], val)
		}
	case []map[string]any:
		idx, err := strconv.Atoi(remaining[0])
		if err != nil || idx < 0 || idx >= len(v) {
			return
		}
		if len(remaining) == 1 {
			if m, ok := val.(map[string]any); ok {
				v[idx] = m
			}
			return
		}
		dotSet(v[idx], remaining[1:], val)
	default:
		nextMap := make(map[string]any)
		data[segment] = nextMap
		dotSet(nextMap, remaining, val)
	}
}

// collectKeys recursively collects all dot-notation keys.
func collectKeys(data any, prefix string, keys *[]string) {
	switch v := data.(type) {
	case map[string]any:
		for key, val := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}
			*keys = append(*keys, fullKey)
			collectKeys(val, fullKey, keys)
		}
	case []any:
		for i, val := range v {
			fullKey := strconv.Itoa(i)
			if prefix != "" {
				fullKey = prefix + "." + strconv.Itoa(i)
			}
			*keys = append(*keys, fullKey)
			collectKeys(val, fullKey, keys)
		}
	case []map[string]any:
		for i, val := range v {
			fullKey := strconv.Itoa(i)
			if prefix != "" {
				fullKey = prefix + "." + strconv.Itoa(i)
			}
			*keys = append(*keys, fullKey)
			collectKeys(val, fullKey, keys)
		}
	}
}

// expandWildcardFields expands wildcard (*) patterns in field keys to concrete data paths.
// If keepUnmatched is true, patterns with no matching data keys are kept as-is.
func expandWildcardFields[T any](fields map[string]T, dataKeys []string, keepUnmatched bool) map[string]T {
	expanded := make(map[string]T)

	for field, value := range fields {
		if !strings.Contains(field, "*") {
			expanded[field] = value
			continue
		}

		pattern := "^" + regexp.QuoteMeta(field) + "$"
		pattern = strings.ReplaceAll(pattern, `\*`, `[^.]+`)
		re, err := regexp.Compile(pattern)
		if err != nil {
			expanded[field] = value
			continue
		}

		matched := false
		for _, key := range dataKeys {
			if re.MatchString(key) {
				expanded[key] = value
				matched = true
			}
		}

		if !matched && keepUnmatched {
			expanded[field] = value
		}
	}

	return expanded
}

// urlValuesToMap converts url.Values to map[string]any, unwrapping single-element slices.
func urlValuesToMap(values url.Values) map[string]any {
	result := make(map[string]any, len(values))
	for key, vals := range values {
		if len(vals) == 1 {
			result[key] = vals[0]
		} else {
			s := make([]any, len(vals))
			for i, v := range vals {
				s[i] = v
			}
			result[key] = s
		}
	}
	return result
}

// structToMap converts a struct to map using "form" tags.
func structToMap(rv reflect.Value) map[string]any {
	result := make(map[string]any)
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("form")
		if tag == "" || tag == "-" {
			tag = field.Tag.Get("json")
		}
		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = field.Name
		}

		// Handle tag options like "name,omitempty"
		if idx := strings.Index(tag, ","); idx != -1 {
			tag = tag[:idx]
		}
		if tag == "" {
			continue
		}

		val := rv.Field(i)
		if field.Anonymous && val.Kind() == reflect.Struct {
			embedded := structToMap(val)
			for k, v := range embedded {
				result[k] = v
			}
			continue
		}

		result[tag] = normalizeValue(val)
	}

	return result
}

// normalizeValue recursively converts reflect.Value to map[string]any / []any
// so that dotGet and collectKeys can traverse nested data.
func normalizeValue(rv reflect.Value) any {
	if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return structToMap(rv)
	case reflect.Slice, reflect.Array:
		result := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = normalizeValue(rv.Index(i))
		}
		return result
	case reflect.Map:
		result := make(map[string]any, rv.Len())
		for _, key := range rv.MapKeys() {
			result[fmt.Sprintf("%v", key.Interface())] = normalizeValue(rv.MapIndex(key))
		}
		return result
	default:
		return rv.Interface()
	}
}
