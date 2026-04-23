package validation

import (
	"sort"
	"strings"

	contractstranslation "github.com/goravel/framework/contracts/translation"
)

// defaultMessages contains all default validation error messages.
// It is populated at init time from the embedded lang/en/validation.json file.
// Size-dependent rules have type-specific variants: "rule.string", "rule.numeric", "rule.array", "rule.file".
var defaultMessages map[string]string

// sizeRules are rules that have type-specific messages.
var sizeRules = map[string]bool{
	"size": true, "min": true, "max": true, "between": true,
	"gt": true, "gte": true, "lt": true, "lte": true,
}

// getMessage resolves the error message for a rule failure.
// Priority:
// 1. Custom field+rule message: customMessages["field.rule"]
// 2. Custom rule message: customMessages["rule"]
// 3. Translated message: translator.Get("validation.rule") (if translator available)
// 4. Type-specific default: defaultMessages["rule.type"] (for size rules)
// 5. Generic default: defaultMessages["rule"]
func getMessage(field, rule string, customMessages map[string]string, attrType string, translator contractstranslation.Translator) string {
	// 1. Custom field+rule message
	if msg, ok := customMessages[field+"."+rule]; ok {
		return msg
	}

	// 2. Custom rule message
	if msg, ok := customMessages[rule]; ok {
		return msg
	}

	// 3. Translated message
	if translator != nil {
		if sizeRules[rule] {
			key := "validation." + rule + "." + attrType
			if translator.Has(key) {
				return translator.Get(key)
			}
		}
		key := "validation." + rule
		if translator.Has(key) {
			return translator.Get(key)
		}
	}

	// 4. Type-specific default for size rules
	if sizeRules[rule] {
		if msg, ok := defaultMessages[rule+"."+attrType]; ok {
			return msg
		}
	}

	// 5. Generic default
	if msg, ok := defaultMessages[rule]; ok {
		return msg
	}

	return "The :attribute field is invalid."
}

// formatMessage replaces placeholders in a message template.
// Keys are sorted by length descending so that longer placeholders (e.g. ":values")
// are replaced before shorter ones (e.g. ":value") to avoid partial replacements.
func formatMessage(message string, replacements map[string]string) string {
	if len(replacements) == 0 {
		return message
	}

	// Fast path: most error messages have 1-4 replacements.
	// Use a fixed-size array with insertion sort to avoid slice allocation.
	if len(replacements) <= 4 {
		type kv struct{ k, v string }
		var sorted [4]kv
		n := 0
		for k, v := range replacements {
			// Insertion sort by key length descending
			pos := n
			for pos > 0 && len(sorted[pos-1].k) < len(k) {
				sorted[pos] = sorted[pos-1]
				pos--
			}
			sorted[pos] = kv{k, v}
			n++
		}
		for i := range n {
			message = strings.ReplaceAll(message, sorted[i].k, sorted[i].v)
		}
		return message
	}

	keys := make([]string, 0, len(replacements))
	for k := range replacements {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})
	for _, k := range keys {
		message = strings.ReplaceAll(message, k, replacements[k])
	}
	return message
}

// getDisplayableAttribute returns a human-readable field name.
func getDisplayableAttribute(field string, customAttributes map[string]string) string {
	if name, ok := customAttributes[field]; ok {
		return name
	}
	return strings.ReplaceAll(field, "_", " ")
}
