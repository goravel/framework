package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

// RuleContext provides context for rule evaluation.
type RuleContext struct {
	Ctx        context.Context
	Attribute  string                  // Current field name
	Value      any                     // Field value
	Parameters []string                // Rule parameters
	Data       *DataBag                // Full data set
	Rules      map[string][]ParsedRule // All field rules (for type inference)
}

// builtinRules maps rule names to their implementations.
var builtinRules = map[string]func(ctx *RuleContext) bool{}

// implicitRules are rules that run even when the field is missing or empty.
var implicitRules = map[string]bool{}

// excludeRules are rules that may cause a field to be excluded from validated data.
var excludeRules = map[string]bool{}

// numericRuleNames are rules that indicate a field should be treated as numeric for size rules.
var numericRuleNames = map[string]bool{}

// ---- Helper functions ----

// isValueEmpty checks if a value is considered "empty" for validation purposes.
func isValueEmpty(val any) bool {
	if val == nil {
		return true
	}
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []any:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	}
	return false
}

// isValuePresent checks if a value is "present" (not nil/empty).
func isValuePresent(val any) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v) != ""
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			return rv.Len() > 0
		default:
			return true
		}
	}
}

// toFloat64 attempts to convert a value to float64.
func toFloat64(val any) (float64, bool) {
	switch v := val.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return 0, false
		}
		return f, true
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

// getSize returns the "size" of a value based on its attribute type.
func getSize(val any, attrType string) (float64, bool) {
	switch attrType {
	case "numeric":
		return toFloat64(val)
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
