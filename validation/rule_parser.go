package validation

import "strings"

// ParsedRule represents a single parsed validation rule with its name and parameters.
type ParsedRule struct {
	Name       string
	Parameters []string
}

// ParseRuleSlice parses a slice of individual rule strings into a slice of ParsedRule.
// Each element is parsed independently, avoiding pipe-splitting issues with regex patterns.
// Example: []string{"required", "regex:^(foo|bar)$", "string"} -> [{required []}, {regex [^(foo|bar)$]}, {string []}]
func ParseRuleSlice(rules []string) []ParsedRule {
	result := make([]ParsedRule, 0, len(rules))
	for _, raw := range rules {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		result = append(result, parseOneRule(raw))
	}
	return result
}

// ParseRules parses a pipe-separated rule string into a slice of ParsedRule.
// Example: "required|string|max:255|in:a,b,c" -> [{required []}, {string []}, {max [255]}, {in [a b c]}]
func ParseRules(ruleString string) []ParsedRule {
	if ruleString == "" {
		return nil
	}

	rawRules := splitRules(ruleString)
	rules := make([]ParsedRule, 0, len(rawRules))

	for _, raw := range rawRules {
		rules = append(rules, parseOneRule(raw))
	}

	return rules
}

// splitRules splits a rule string by '|', respecting escaped '\|'.
// Empty segments are skipped and each segment is trimmed.
// Special handling for regex/not_regex: once we encounter regex: or not_regex:,
// everything after the first ':' is a single parameter (may contain '|').
func splitRules(ruleString string) []string {
	var rules []string
	current := strings.Builder{}

	i := 0
	for i < len(ruleString) {
		// Check for escaped pipe
		if i < len(ruleString)-1 && ruleString[i] == '\\' && ruleString[i+1] == '|' {
			current.WriteByte('|')
			i += 2
			continue
		}

		if ruleString[i] == '|' {
			rule := strings.TrimSpace(current.String())
			// Check if the previous rule is regex or not_regex — if so,
			// the rest of the string is part of the regex parameter
			if rule != "" {
				ruleName := extractRuleName(rule)
				if ruleName == "regex" || ruleName == "not_regex" {
					// Everything remaining (including this '|') is part of the regex
					current.WriteByte('|')
					i++
					// Consume the rest
					current.WriteString(ruleString[i:])
					break
				}
				rules = append(rules, rule)
			}

			current.Reset()
			i++
			continue
		}

		current.WriteByte(ruleString[i])
		i++
	}

	if s := strings.TrimSpace(current.String()); s != "" {
		rules = append(rules, s)
	}

	return rules
}

// extractRuleName extracts the rule name from a rule string (before the first ':').
func extractRuleName(rule string) string {
	idx := strings.Index(rule, ":")
	if idx == -1 {
		return strings.TrimSpace(rule)
	}
	return strings.TrimSpace(rule[:idx])
}

// parseOneRule parses a single rule string "name:param1,param2" into a ParsedRule.
func parseOneRule(raw string) ParsedRule {
	idx := strings.Index(raw, ":")
	if idx == -1 {
		return ParsedRule{Name: raw}
	}

	name := raw[:idx]
	paramStr := raw[idx+1:]

	// For regex and not_regex, the entire parameter string is a single parameter
	if name == "regex" || name == "not_regex" {
		return ParsedRule{
			Name:       name,
			Parameters: []string{paramStr},
		}
	}

	params := splitParameters(paramStr)

	return ParsedRule{
		Name:       name,
		Parameters: params,
	}
}

// splitParameters splits parameter string by comma, respecting escaped '\,'.
func splitParameters(paramStr string) []string {
	if paramStr == "" {
		return nil
	}

	var params []string
	current := strings.Builder{}

	i := 0
	for i < len(paramStr) {
		if i < len(paramStr)-1 && paramStr[i] == '\\' && paramStr[i+1] == ',' {
			current.WriteByte(',')
			i += 2
			continue
		}

		if paramStr[i] == ',' {
			params = append(params, current.String())
			current.Reset()
			i++
			continue
		}

		current.WriteByte(paramStr[i])
		i++
	}

	params = append(params, current.String())

	return params
}
