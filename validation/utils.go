package validation

import (
	"fmt"
	"mime/multipart"
	"reflect"
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
	}

	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() == 0
	}
	return false
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

	// Fallback: infer numeric type from runtime value kind
	if value != nil {
		switch reflect.TypeOf(value).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return "numeric"
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
