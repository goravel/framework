package validation

import (
	"context"
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
