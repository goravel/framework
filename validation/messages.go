package validation

import (
	"sort"
	"strings"
)

// defaultMessages contains all default validation error messages.
// Size-dependent rules have type-specific variants: "rule.string", "rule.numeric", "rule.array", "rule.file".
var defaultMessages = map[string]string{
	// Existence rules
	"required":             "The :attribute field is required.",
	"required_if":          "The :attribute field is required when :other is :value.",
	"required_if_accepted": "The :attribute field is required when :other is accepted.",
	"required_if_declined": "The :attribute field is required when :other is declined.",
	"required_unless":      "The :attribute field is required unless :other is in :values.",
	"required_with":        "The :attribute field is required when :values is present.",
	"required_with_all":    "The :attribute field is required when :values are present.",
	"required_without":     "The :attribute field is required when :values is not present.",
	"required_without_all": "The :attribute field is required when none of :values are present.",
	"filled":               "The :attribute field must have a value.",
	"present":              "The :attribute field must be present.",
	"present_if":           "The :attribute field must be present when :other is :value.",
	"present_unless":       "The :attribute field must be present unless :other is :value.",
	"present_with":         "The :attribute field must be present when :values is present.",
	"present_with_all":     "The :attribute field must be present when :values are present.",
	"missing":              "The :attribute field must be missing.",
	"missing_if":           "The :attribute field must be missing when :other is :value.",
	"missing_unless":       "The :attribute field must be missing unless :other is :value.",
	"missing_with":         "The :attribute field must be missing when :values is present.",
	"missing_with_all":     "The :attribute field must be missing when :values are present.",

	// Accept/Decline rules
	"accepted":    "The :attribute field must be accepted.",
	"accepted_if": "The :attribute field must be accepted when :other is :value.",
	"declined":    "The :attribute field must be declined.",
	"declined_if": "The :attribute field must be declined when :other is :value.",

	// Prohibition rules
	"prohibited":             "The :attribute field is prohibited.",
	"prohibited_if":          "The :attribute field is prohibited when :other is :value.",
	"prohibited_if_accepted": "The :attribute field is prohibited when :other is accepted.",
	"prohibited_if_declined": "The :attribute field is prohibited when :other is declined.",
	"prohibited_unless":      "The :attribute field is prohibited unless :other is in :values.",
	"prohibits":              "The :attribute field prohibits :other from being present.",

	// Type rules
	"string":  "The :attribute field must be a string.",
	"integer": "The :attribute field must be an integer.",
	"int":     "The :attribute field must be an integer.",
	"uint":    "The :attribute field must be a positive integer.",
	"numeric": "The :attribute field must be a number.",
	"boolean": "The :attribute field must be true or false.",
	"bool":    "The :attribute field must be true or false.",
	"float":   "The :attribute field must be a float.",
	"array":   "The :attribute field must be an array.",
	"list":    "The :attribute field must be a list.",
	"slice":   "The :attribute field must be a slice.",
	"map":     "The :attribute field must be a map.",

	// Size rules - string variants
	"size.string":     "The :attribute field must be :size characters.",
	"size.numeric":    "The :attribute field must be :size.",
	"size.array":      "The :attribute field must contain :size items.",
	"size.file":       "The :attribute field must be :size kilobytes.",
	"min.string":      "The :attribute field must be at least :min characters.",
	"min.numeric":     "The :attribute field must be at least :min.",
	"min.array":       "The :attribute field must have at least :min items.",
	"min.file":        "The :attribute field must be at least :min kilobytes.",
	"max.string":      "The :attribute field must not be greater than :max characters.",
	"max.numeric":     "The :attribute field must not be greater than :max.",
	"max.array":       "The :attribute field must not have more than :max items.",
	"max.file":        "The :attribute field must not be greater than :max kilobytes.",
	"between.string":  "The :attribute field must be between :min and :max characters.",
	"between.numeric": "The :attribute field must be between :min and :max.",
	"between.array":   "The :attribute field must have between :min and :max items.",
	"between.file":    "The :attribute field must be between :min and :max kilobytes.",
	"gt.string":       "The :attribute field must be greater than :value characters.",
	"gt.numeric":      "The :attribute field must be greater than :value.",
	"gt.array":        "The :attribute field must have more than :value items.",
	"gt.file":         "The :attribute field must be greater than :value kilobytes.",
	"gte.string":      "The :attribute field must be greater than or equal to :value characters.",
	"gte.numeric":     "The :attribute field must be greater than or equal to :value.",
	"gte.array":       "The :attribute field must have :value items or more.",
	"gte.file":        "The :attribute field must be greater than or equal to :value kilobytes.",
	"lt.string":       "The :attribute field must be less than :value characters.",
	"lt.numeric":      "The :attribute field must be less than :value.",
	"lt.array":        "The :attribute field must have less than :value items.",
	"lt.file":         "The :attribute field must be less than :value kilobytes.",
	"lte.string":      "The :attribute field must be less than or equal to :value characters.",
	"lte.numeric":     "The :attribute field must be less than or equal to :value.",
	"lte.array":       "The :attribute field must not have more than :value items.",
	"lte.file":        "The :attribute field must be less than or equal to :value kilobytes.",

	// Numeric rules
	"digits":         "The :attribute field must be :digits digits.",
	"digits_between": "The :attribute field must be between :min and :max digits.",
	"decimal":        "The :attribute field must have :decimal decimal places.",
	"multiple_of":    "The :attribute field must be a multiple of :value.",
	"min_digits":     "The :attribute field must have at least :min digits.",
	"max_digits":     "The :attribute field must not have more than :max digits.",

	// String format rules
	"alpha":       "The :attribute field must only contain letters.",
	"alpha_num":   "The :attribute field must only contain letters and numbers.",
	"alpha_dash":  "The :attribute field must only contain letters, numbers, dashes, and underscores.",
	"ascii":       "The :attribute field must only contain single-byte alphanumeric characters and symbols.",
	"email":       "The :attribute field must be a valid email address.",
	"url":         "The :attribute field must be a valid URL.",
	"active_url":  "The :attribute field must be a valid URL.",
	"ip":          "The :attribute field must be a valid IP address.",
	"ipv4":        "The :attribute field must be a valid IPv4 address.",
	"ipv6":        "The :attribute field must be a valid IPv6 address.",
	"mac_address": "The :attribute field must be a valid MAC address.",
	"mac":         "The :attribute field must be a valid MAC address.",
	"json":        "The :attribute field must be a valid JSON string.",
	"uuid":        "The :attribute field must be a valid UUID.",
	"uuid3":       "The :attribute field must be a valid UUID v3.",
	"uuid4":       "The :attribute field must be a valid UUID v4.",
	"uuid5":       "The :attribute field must be a valid UUID v5.",
	"ulid":        "The :attribute field must be a valid ULID.",
	"hex_color":   "The :attribute field must be a valid hexadecimal color.",
	"regex":       "The :attribute field format is invalid.",
	"not_regex":   "The :attribute field format is invalid.",
	"lowercase":   "The :attribute field must be lowercase.",
	"uppercase":   "The :attribute field must be uppercase.",

	// String content rules
	"starts_with":       "The :attribute field must start with one of the following: :values.",
	"doesnt_start_with": "The :attribute field must not start with one of the following: :values.",
	"ends_with":         "The :attribute field must end with one of the following: :values.",
	"doesnt_end_with":   "The :attribute field must not end with one of the following: :values.",
	"contains":          "The :attribute field is missing a required value.",
	"doesnt_contain":    "The :attribute field must not contain any of the following: :values.",
	"confirmed":         "The :attribute field confirmation does not match.",

	// Comparison rules
	"same":          "The :attribute field must match :other.",
	"different":     "The :attribute field and :other must be different.",
	"eq":            "The :attribute field must be equal to :value.",
	"ne":            "The :attribute field must not be equal to :value.",
	"in":            "The selected :attribute is invalid.",
	"not_in":        "The selected :attribute is invalid.",
	"in_array":      "The :attribute field must exist in :other.",
	"in_array_keys": "The :attribute field must contain at least one of the specified keys.",

	// Date rules
	"date":            "The :attribute field must be a valid date.",
	"date_format":     "The :attribute field must match the format :format.",
	"date_equals":     "The :attribute field must be a date equal to :date.",
	"before":          "The :attribute field must be a date before :date.",
	"before_or_equal": "The :attribute field must be a date before or equal to :date.",
	"after":           "The :attribute field must be a date after :date.",
	"after_or_equal":  "The :attribute field must be a date after or equal to :date.",
	"timezone":        "The :attribute field must be a valid timezone.",

	// Exclude rules
	"exclude":         "The :attribute field is excluded.",
	"exclude_if":      "The :attribute field is excluded when :other is :value.",
	"exclude_unless":  "The :attribute field is excluded unless :other is :value.",
	"exclude_with":    "The :attribute field is excluded when :values is present.",
	"exclude_without": "The :attribute field is excluded when :values is not present.",

	// File rules
	"file":       "The :attribute field must be a file.",
	"image":      "The :attribute field must be an image.",
	"mimes":      "The :attribute field must be a file of type: :values.",
	"mimetypes":  "The :attribute field must be a file of type: :values.",
	"extensions": "The :attribute field must have one of the following extensions: :values.",
	"dimensions": "The :attribute field has invalid image dimensions.",
	"encoding":   "The :attribute field must be encoded as :values.",

	// Other rules
	"distinct":            "The :attribute field has a duplicate value.",
	"required_array_keys": "The :attribute field must contain entries for: :values.",

	// Database rules
	"exists": "The :attribute does not exist.",
	"unique": "The :attribute has already been taken.",

	// Deprecated: use the new names instead, will be removed in the next version.
	"len":       "The :attribute field must be :size characters.",
	"min_len":   "The :attribute field must be at least :min characters.",
	"max_len":   "The :attribute field must not be greater than :max characters.",
	"eq_field":  "The :attribute field must match :other.",
	"ne_field":  "The :attribute field and :other must be different.",
	"gt_field":  "The :attribute field must be greater than :value.",
	"gte_field": "The :attribute field must be greater than or equal to :value.",
	"lt_field":  "The :attribute field must be less than :value.",
	"lte_field": "The :attribute field must be less than or equal to :value.",
	"gt_date":   "The :attribute field must be a date after :date.",
	"lt_date":   "The :attribute field must be a date before :date.",
	"gte_date":  "The :attribute field must be a date after or equal to :date.",
	"lte_date":  "The :attribute field must be a date before or equal to :date.",
	"number":    "The :attribute field must be a number.",
	"full_url":  "The :attribute field must be a valid URL.",
}

// sizeRules are rules that have type-specific messages.
var sizeRules = map[string]bool{
	"size": true, "min": true, "max": true, "between": true,
	"gt": true, "gte": true, "lt": true, "lte": true,
}

// getMessage resolves the error message for a rule failure.
// Priority:
// 1. Custom field+rule message: customMessages["field.rule"]
// 2. Custom rule message: customMessages["rule"]
// 3. Type-specific default: defaultMessages["rule.type"] (for size rules)
// 4. Generic default: defaultMessages["rule"]
func getMessage(field, rule string, customMessages map[string]string, attrType string) string {
	// 1. Custom field+rule message
	if msg, ok := customMessages[field+"."+rule]; ok {
		return msg
	}

	// 2. Custom rule message
	if msg, ok := customMessages[rule]; ok {
		return msg
	}

	// 3. Type-specific default for size rules
	if sizeRules[rule] {
		if msg, ok := defaultMessages[rule+"."+attrType]; ok {
			return msg
		}
	}

	// 4. Generic default
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
