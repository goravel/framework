package validation

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	contractsvalidation "github.com/goravel/framework/contracts/validation"
)

// Engine is the core validation orchestrator.
type Engine struct {
	ctx            context.Context
	data           *DataBag
	rules          map[string][]ParsedRule
	customRules    map[string]contractsvalidation.Rule
	messages       map[string]string
	attributes     map[string]string
	errors         *Errors
	excludes       map[string]bool
	distinctValues map[string]map[string]bool // For tracking distinct values
	expandedRules  map[string][]ParsedRule    // Cached result of expandWildcardRules
	ruleCtx        RuleContext                // Reusable context for builtin rule execution
}

type engineOptions struct {
	customRules map[string]contractsvalidation.Rule
	messages    map[string]string
	attributes  map[string]string
}

// NewEngine creates a new validation engine.
func NewEngine(ctx context.Context, data *DataBag, rules map[string][]ParsedRule, opts engineOptions) *Engine {
	return &Engine{
		ctx:            ctx,
		data:           data,
		rules:          rules,
		customRules:    opts.customRules,
		messages:       opts.messages,
		attributes:     opts.attributes,
		errors:         NewErrors(),
		excludes:       make(map[string]bool),
		distinctValues: make(map[string]map[string]bool),
	}
}

// Validate runs all validation rules and returns the error bag.
func (e *Engine) Validate() *Errors {
	// Step 1: Expand wildcards
	expandedRules := e.expandWildcardRules()

	// Step 2: Sort keys for deterministic ordering
	sortedFields := make([]string, 0, len(expandedRules))
	for field := range expandedRules {
		sortedFields = append(sortedFields, field)
	}
	sort.Strings(sortedFields)

	// Step 3: Validate each field
	for _, field := range sortedFields {
		fieldRules := expandedRules[field]
		e.validateField(field, fieldRules, expandedRules)
	}

	return e.errors
}

// ValidatedData returns the data for all fields that have validation rules,
// excluding fields marked for exclusion. Callers should check Errors first
// to determine if validation passed before using the result.
func (e *Engine) ValidatedData() map[string]any {
	result := make(map[string]any)
	expandedRules := e.expandWildcardRules()

	for field := range expandedRules {
		if e.excludes[field] {
			continue
		}
		if val, ok := e.data.Get(field); ok {
			setValidated(result, e.data.All(), strings.Split(field, "."), val)
		}
	}

	if normalized, ok := normalizeValidatedShape(result, e.data.All()).(map[string]any); ok {
		return normalized
	}
	return result
}

// expandWildcardRules expands rules with wildcard (*) patterns based on actual data.
// Results are cached for the lifetime of the Engine.
func (e *Engine) expandWildcardRules() map[string][]ParsedRule {
	if e.expandedRules != nil {
		return e.expandedRules
	}
	e.expandedRules = expandWildcardFields(e.rules, e.data.Keys(), true)
	return e.expandedRules
}

// validateField validates a single field against all its rules.
func (e *Engine) validateField(field string, fieldRules []ParsedRule, allRules map[string][]ParsedRule) {
	if e.isExcluded(field) {
		return
	}

	hasBail := false
	hasNullable := false
	hasSometimes := false

	// Pre-scan for control rules
	for _, rule := range fieldRules {
		switch rule.Name {
		case "bail":
			hasBail = true
		case "nullable":
			hasNullable = true
		case "sometimes":
			hasSometimes = true
		}
	}

	// Sometimes: skip if field not present in data
	if hasSometimes && !e.data.Has(field) {
		return
	}

	// Pre-pass: run exclude rules before any validation to ensure order-independence
	for _, rule := range fieldRules {
		if !excludeRules[rule.Name] {
			continue
		}
		e.handleExcludeRule(field, rule)
		if e.isExcluded(field) {
			return
		}
	}

	value, exists := e.data.Get(field)

	// Nullable: if value is nil, skip non-implicit rules
	isNull := !exists || value == nil

	for _, rule := range fieldRules {
		// Skip control rules
		if rule.Name == "bail" || rule.Name == "nullable" || rule.Name == "sometimes" {
			continue
		}

		isImplicit := implicitRules[rule.Name]

		// Skip exclude rules (already handled in pre-pass)
		if excludeRules[rule.Name] {
			continue
		}

		// Non-implicit rule + value empty/missing → skip
		if !isImplicit {
			if isNull || isValueEmpty(value) {
				if hasNullable && isNull {
					continue
				}
				if !exists {
					continue
				}
				// Empty string for non-implicit rules: skip
				if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
					continue
				}
			}
		}

		// Execute the rule
		passed := e.executeRule(field, rule, value, allRules)

		// Handle distinct tracking: the builtin ruleDistinct always returns true,
		// actual duplicate detection happens here via cross-field value tracking.
		if rule.Name == "distinct" && passed {
			if e.trackDistinct(field, value) {
				passed = false
			}
		}

		if !passed {
			attrType := getAttributeType(field, value, allRules)
			msg := e.formatErrorMessage(field, rule, attrType)
			e.errors.Add(field, rule.Name, msg)
		}

		// Bail: stop on first error for this field
		if hasBail && e.errors.Has(field) {
			break
		}
	}
}

// executeRule runs a single rule and returns whether it passed.
func (e *Engine) executeRule(field string, rule ParsedRule, value any, allRules map[string][]ParsedRule) bool {
	// Check custom rules first
	if customRule, ok := e.customRules[rule.Name]; ok {
		return customRule.Passes(e.ctx, e.data, value, anySlice(rule.Parameters)...)
	}

	// Check built-in rules (reuse RuleContext to avoid heap allocation)
	if fn, ok := builtinRules[rule.Name]; ok {
		e.ruleCtx.Ctx = e.ctx
		e.ruleCtx.Attribute = field
		e.ruleCtx.Value = value
		e.ruleCtx.Parameters = rule.Parameters
		e.ruleCtx.Data = e.data
		e.ruleCtx.Rules = allRules
		return fn(&e.ruleCtx)
	}

	// Unknown rule (should not reach here — Make() validates all rule names)
	return false
}

// handleExcludeRule processes exclude rules and marks fields for exclusion.
func (e *Engine) handleExcludeRule(field string, rule ParsedRule) {
	switch rule.Name {
	case "exclude":
		e.excludes[field] = true

	case "exclude_if":
		if len(rule.Parameters) >= 2 {
			otherField := rule.Parameters[0]
			otherVal, _ := e.data.Get(otherField)
			comparisonValues := rule.Parameters[1:]
			if matchesOtherValue(otherVal, comparisonValues) {
				e.excludes[field] = true
			}
		}

	case "exclude_unless":
		if len(rule.Parameters) >= 2 {
			otherField := rule.Parameters[0]
			otherVal, _ := e.data.Get(otherField)
			comparisonValues := rule.Parameters[1:]
			if !matchesOtherValue(otherVal, comparisonValues) {
				e.excludes[field] = true
			}
		}

	case "exclude_with":
		for _, f := range rule.Parameters {
			if e.data.Has(f) {
				e.excludes[field] = true
				break
			}
		}

	case "exclude_without":
		for _, f := range rule.Parameters {
			if !e.data.Has(f) {
				e.excludes[field] = true
				break
			}
		}
	}
}

// isExcluded checks if a field has been marked for exclusion.
func (e *Engine) isExcluded(field string) bool {
	return e.excludes[field]
}

// trackDistinct tracks values for distinct validation across wildcard-expanded fields.
// Returns true if a duplicate was detected (i.e., validation should fail).
func (e *Engine) trackDistinct(field string, value any) bool {
	// Find the wildcard pattern this field belongs to
	for pattern := range e.rules {
		if strings.Contains(pattern, "*") {
			re := "^" + regexp.QuoteMeta(pattern) + "$"
			re = strings.ReplaceAll(re, `\*`, `[^.]+`)
			if matched, _ := regexp.MatchString(re, field); matched {
				if _, ok := e.distinctValues[pattern]; !ok {
					e.distinctValues[pattern] = make(map[string]bool)
				}
				valStr := fmt.Sprintf("%v", value)
				if e.distinctValues[pattern][valStr] {
					return true // duplicate found
				}
				e.distinctValues[pattern][valStr] = true
				return false
			}
		}
	}
	return false
}

// formatErrorMessage creates the error message for a rule failure.
func (e *Engine) formatErrorMessage(field string, rule ParsedRule, attrType string) string {
	msg := getMessage(field, rule.Name, e.messages, attrType)
	if _, hasFieldRuleMessage := e.messages[field+"."+rule.Name]; !hasFieldRuleMessage {
		if _, hasRuleMessage := e.messages[rule.Name]; !hasRuleMessage {
			if customRule, ok := e.customRules[rule.Name]; ok {
				msg = customRule.Message(e.ctx)
			}
		}
	}

	replacements := map[string]string{
		":attribute": getDisplayableAttribute(field, e.attributes),
	}

	// Add parameter-specific replacements
	switch rule.Name {
	case "min", "min_digits":
		if len(rule.Parameters) > 0 {
			replacements[":min"] = rule.Parameters[0]
		}
	case "max", "max_digits":
		if len(rule.Parameters) > 0 {
			replacements[":max"] = rule.Parameters[0]
		}
	case "between", "digits_between":
		if len(rule.Parameters) >= 2 {
			replacements[":min"] = rule.Parameters[0]
			replacements[":max"] = rule.Parameters[1]
		}
	case "size":
		if len(rule.Parameters) > 0 {
			replacements[":size"] = rule.Parameters[0]
		}
	case "gt", "gte", "lt", "lte":
		if len(rule.Parameters) > 0 {
			replacements[":value"] = getDisplayableAttribute(rule.Parameters[0], e.attributes)
		}
	case "required_if", "required_unless", "prohibited_if", "prohibited_unless",
		"accepted_if", "declined_if", "present_if", "present_unless",
		"missing_if", "missing_unless":
		if len(rule.Parameters) > 0 {
			replacements[":other"] = getDisplayableAttribute(rule.Parameters[0], e.attributes)
		}
		if len(rule.Parameters) > 1 {
			replacements[":value"] = strings.Join(rule.Parameters[1:], ", ")
			replacements[":values"] = strings.Join(rule.Parameters[1:], ", ")
		}
	case "required_if_accepted", "required_if_declined",
		"prohibited_if_accepted", "prohibited_if_declined":
		if len(rule.Parameters) > 0 {
			replacements[":other"] = getDisplayableAttribute(rule.Parameters[0], e.attributes)
		}
	case "required_with", "required_with_all", "required_without", "required_without_all",
		"present_with", "present_with_all", "missing_with", "missing_with_all",
		"exclude_with", "exclude_without":
		if len(rule.Parameters) > 0 {
			names := make([]string, len(rule.Parameters))
			for i, p := range rule.Parameters {
				names[i] = getDisplayableAttribute(p, e.attributes)
			}
			replacements[":values"] = strings.Join(names, ", ")
		}
	case "exclude_if", "exclude_unless":
		if len(rule.Parameters) > 0 {
			replacements[":other"] = getDisplayableAttribute(rule.Parameters[0], e.attributes)
		}
		if len(rule.Parameters) > 1 {
			replacements[":value"] = strings.Join(rule.Parameters[1:], ", ")
		}
	case "same", "different", "in_array", "confirmed", "prohibits":
		if len(rule.Parameters) > 0 {
			replacements[":other"] = getDisplayableAttribute(rule.Parameters[0], e.attributes)
		}
	case "eq", "ne":
		if len(rule.Parameters) > 0 {
			replacements[":value"] = rule.Parameters[0]
		}
	case "digits":
		if len(rule.Parameters) > 0 {
			replacements[":digits"] = rule.Parameters[0]
		}
	case "decimal":
		if len(rule.Parameters) > 0 {
			replacements[":decimal"] = rule.Parameters[0]
		}
	case "multiple_of":
		if len(rule.Parameters) > 0 {
			replacements[":value"] = rule.Parameters[0]
		}
	case "in", "not_in":
		replacements[":values"] = strings.Join(rule.Parameters, ", ")
	case "starts_with", "doesnt_start_with", "ends_with", "doesnt_end_with",
		"doesnt_contain", "mimes", "mimetypes", "extensions",
		"encoding", "required_array_keys", "in_array_keys":
		replacements[":values"] = strings.Join(rule.Parameters, ", ")
	case "before", "before_or_equal", "after", "after_or_equal", "date_equals":
		if len(rule.Parameters) > 0 {
			replacements[":date"] = rule.Parameters[0]
		}
	case "date_format":
		if len(rule.Parameters) > 0 {
			replacements[":format"] = rule.Parameters[0]
		}
	// Deprecated: will be removed in the next version.
	case "len":
		if len(rule.Parameters) > 0 {
			replacements[":size"] = rule.Parameters[0]
		}
	case "min_len":
		if len(rule.Parameters) > 0 {
			replacements[":min"] = rule.Parameters[0]
		}
	case "max_len":
		if len(rule.Parameters) > 0 {
			replacements[":max"] = rule.Parameters[0]
		}
	case "eq_field", "ne_field":
		if len(rule.Parameters) > 0 {
			replacements[":other"] = getDisplayableAttribute(rule.Parameters[0], e.attributes)
		}
	case "gt_field", "gte_field", "lt_field", "lte_field":
		if len(rule.Parameters) > 0 {
			replacements[":value"] = getDisplayableAttribute(rule.Parameters[0], e.attributes)
		}
	case "gt_date", "lt_date", "gte_date", "lte_date":
		if len(rule.Parameters) > 0 {
			replacements[":date"] = rule.Parameters[0]
		}
	}

	return formatMessage(msg, replacements)
}

// anySlice converts []string to []any.
func anySlice(s []string) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}
