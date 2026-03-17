package validation

import (
	"fmt"
	"mime/multipart"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gabriel-vasile/mimetype"
	"github.com/spf13/cast"
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

// getSize returns the "size" of a value based on its attribute type.
func getSize(val any, attrType string) (float64, bool) {
	switch attrType {
	case "numeric":
		num, err := cast.ToFloat64E(val)
		return num, err == nil
	case "array":
		if val == nil {
			return 0, false
		}
		rv := reflect.ValueOf(val)
		kind := rv.Kind()
		if kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map {
			return float64(rv.Len()), true
		}
		return 0, false
	case "file":
		if fh, ok := val.(*multipart.FileHeader); ok {
			return float64(fh.Size) / 1024, true // kilobytes
		}
		if fhs, ok := val.([]*multipart.FileHeader); ok {
			var total int64
			for _, fh := range fhs {
				total += fh.Size
			}
			return float64(total) / 1024, true
		}
		return 0, false
	default: // string
		s := fmt.Sprintf("%v", val)
		return float64(utf8.RuneCountInString(s)), true
	}
}

// parseDateValue attempts to parse a date from a value or field reference.
func parseDateValue(val string, data *DataBag) (time.Time, bool) {
	// Try as a field reference first
	if fieldVal, ok := data.Get(val); ok {
		return parseDate(fieldVal)
	}
	return parseDate(val)
}

// parseDate attempts to parse a value as a date.
func parseDate(val any) (time.Time, bool) {
	switch v := val.(type) {
	case time.Time:
		return v, true
	case string:
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
			time.RFC1123,
			time.RFC822,
		}
		for _, f := range formats {
			if t, err := time.Parse(f, v); err == nil {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

// isAcceptedValue checks if a value is one of the "accepted" values.
func isAcceptedValue(val any) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case string:
		v = strings.ToLower(strings.TrimSpace(v))
		return v == "yes" || v == "on" || v == "1" || v == "true"
	}
	v := cast.ToInt(val)
	return v == 1
}

// isDeclinedValue checks if a value is one of the "declined" values.
func isDeclinedValue(val any) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case string:
		v = strings.ToLower(strings.TrimSpace(v))
		return v == "no" || v == "off" || v == "0" || v == "false"
	}
	v := cast.ToInt(val)
	return v == 0
}

// parseDependentValues extracts the other field's value and comparison values from parameters.
// params[0] is the other field name, params[1:] are comparison values.
func parseDependentValues(ctx *RuleContext) (otherValue any, comparisonValues []string, otherField string) {
	if len(ctx.Parameters) == 0 {
		return nil, nil, ""
	}
	otherField = ctx.Parameters[0]
	otherValue, _ = ctx.Data.Get(otherField)
	comparisonValues = ctx.Parameters[1:]
	return
}

// toCamelCase converts a string to camelCase.
func toCamelCase(s string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return ""
	}

	result := strings.ToLower(words[0])
	for _, w := range words[1:] {
		if len(w) > 0 {
			runes := []rune(strings.ToLower(w))
			runes[0] = unicode.ToUpper(runes[0])
			result += string(runes)
		}
	}

	return result
}

// toSnakeCase converts a string to snake_case.
func toSnakeCase(s string) string {
	words := splitWords(s)
	for i, w := range words {
		words[i] = strings.ToLower(w)
	}
	return strings.Join(words, "_")
}

// splitWords splits a string into words based on separators and case changes.
func splitWords(s string) []string {
	// Replace common separators with spaces
	s = strings.NewReplacer("-", " ", "_", " ").Replace(s)

	// Split on camelCase boundaries
	var words []string
	current := strings.Builder{}

	runes := []rune(s)
	for i, r := range runes {
		if r == ' ' {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			continue
		}

		if i > 0 && unicode.IsUpper(r) && !unicode.IsUpper(runes[i-1]) {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// stripHTMLTags removes HTML tags from a string.
func stripHTMLTags(s string) string {
	return htmlTagRegex.ReplaceAllString(s, "")
}

// escapeJS escapes a string for safe embedding in JavaScript.
func escapeJS(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		`'`, `\'`,
		"\n", `\n`,
		"\r", `\r`,
		"<", `\x3c`,
		">", `\x3e`,
		"/", `\/`,
	)
	return replacer.Replace(s)
}

// strToInts splits a comma-separated string into []int.
func strToInts(s string) []int {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		result = append(result, cast.ToInt(p))
	}
	return result
}

// strToArray splits a comma-separated string into []string.
func strToArray(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		result = append(result, p)
	}
	return result
}

// detectMIME detects the real MIME type of a multipart file by reading its content.
func detectMIME(fh *multipart.FileHeader) (*mimetype.MIME, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer func(f multipart.File) { _ = f.Close() }(f)

	return mimetype.DetectReader(f)
}

func getFileExtension(filename string) string {
	idx := strings.LastIndex(filename, ".")
	if idx == -1 {
		return ""
	}
	return filename[idx+1:]
}
