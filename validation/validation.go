package validation

import (
	"context"
	"net/http"
	"net/url"
	"slices"

	validatecontract "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/errors"
)

type Validation struct {
	rules   []validatecontract.Rule
	filters []validatecontract.Filter
}

func NewValidation() *Validation {
	return &Validation{
		rules:   make([]validatecontract.Rule, 0),
		filters: make([]validatecontract.Filter, 0),
	}
}

func (r *Validation) Make(ctx context.Context, data any, rules map[string]any, options ...validatecontract.Option) (validatecontract.Validator, error) {
	if data == nil {
		return nil, errors.ValidationEmptyData
	}
	if len(rules) == 0 {
		return nil, errors.ValidationEmptyRules
	}

	// Process options
	opts := applyOptions(options)

	// Create DataBag from data
	var bag *DataBag
	var err error
	switch td := data.(type) {
	case *DataBag:
		bag = td
	case *http.Request:
		bag, err = NewDataBagFromRequest(td, opts.MaxMultipartMemory)
	case map[string]any:
		bag, err = NewDataBag(td)
	case url.Values:
		bag, err = NewDataBag(td)
	case map[string][]string:
		bag, err = NewDataBag(td)
	default:
		bag, err = NewDataBag(data)
	}
	if err != nil {
		return nil, errors.ValidationDataInvalidType
	}

	// Merge globally registered filters with per-call custom filters
	if len(r.filters) > 0 {
		opts.CustomFilters = append(opts.CustomFilters, r.filters...)
	}

	// Run PrepareForValidation if set
	if opts.PrepareForValidation != nil {
		if err := opts.PrepareForValidation(ctx, bag); err != nil {
			return nil, err
		}
	}

	// Apply filters before validation
	if len(opts.Filters) > 0 {
		if err := applyFilters(ctx, bag, opts.Filters, opts.CustomFilters); err != nil {
			return nil, err
		}
	}

	// Parse all rule strings
	parsedRules := make(map[string][]ParsedRule)
	for field, ruleVal := range rules {
		switch v := ruleVal.(type) {
		case string:
			parsedRules[field] = ParseRules(v)
		case []string:
			parsedRules[field] = ParseRuleSlice(v)
		default:
			return nil, errors.ValidationInvalidRuleType.Args(field)
		}
	}

	// Build custom rules map
	customRulesMap := make(map[string]validatecontract.Rule)
	for _, cr := range r.rules {
		customRulesMap[cr.Signature()] = cr
	}

	// Validate that all rule names are known (builtin, custom, or control)
	for field, fieldRules := range parsedRules {
		for _, pr := range fieldRules {
			// skip control rules
			if slices.Contains([]string{"bail", "sometimes", "nullable"}, pr.Name) {
				continue
			}
			if _, ok := builtinRules[pr.Name]; ok {
				continue
			}
			if _, ok := excludeRules[pr.Name]; ok {
				continue
			}
			if _, ok := customRulesMap[pr.Name]; ok {
				continue
			}
			return nil, errors.ValidationUnknownRule.Args(field + "." + pr.Name)
		}
	}

	// Get custom messages
	customMessages := opts.Messages
	if customMessages == nil {
		customMessages = make(map[string]string)
	}

	// Get custom attributes
	customAttributes := opts.Attributes
	if customAttributes == nil {
		customAttributes = make(map[string]string)
	}

	// Create and run engine
	engine := NewEngine(ctx, bag, parsedRules, engineOptions{
		customRules: customRulesMap,
		messages:    customMessages,
		attributes:  customAttributes,
	})
	errorBag := engine.Validate()
	validatedData := engine.ValidatedData()

	return NewValidator(bag, errorBag, validatedData), nil
}

func (r *Validation) AddRules(rules []validatecontract.Rule) error {
	existRuleNames := r.existRuleNames()
	for _, rule := range rules {
		if slices.Contains(existRuleNames, rule.Signature()) {
			return errors.ValidationDuplicateRule.Args(rule.Signature())
		}
		existRuleNames = append(existRuleNames, rule.Signature())
	}

	r.rules = append(r.rules, rules...)
	return nil
}

func (r *Validation) AddFilters(filters []validatecontract.Filter) error {
	existFilterNames := r.existFilterNames()
	for _, filter := range filters {
		if slices.Contains(existFilterNames, filter.Signature()) {
			return errors.ValidationDuplicateFilter.Args(filter.Signature())
		}
		existFilterNames = append(existFilterNames, filter.Signature())
	}

	r.filters = append(r.filters, filters...)
	return nil
}

func (r *Validation) Rules() []validatecontract.Rule {
	return r.rules
}

func (r *Validation) Filters() []validatecontract.Filter {
	return r.filters
}

func (r *Validation) existRuleNames() []string {
	names := make([]string, 0, len(builtinRules)+len(r.rules))
	for name := range builtinRules {
		names = append(names, name)
	}
	for _, rule := range r.rules {
		names = append(names, rule.Signature())
	}
	return names
}

func (r *Validation) existFilterNames() []string {
	names := make([]string, 0, len(builtinFilters)+len(r.filters))
	for name := range builtinFilters {
		names = append(names, name)
	}
	for _, filter := range r.filters {
		names = append(names, filter.Signature())
	}
	return names
}
