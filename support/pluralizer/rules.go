package pluralizer

import (
	"regexp"
	"strings"
)

type Rule struct {
	Pattern     *regexp.Regexp
	Replacement string
}

type RuleSet struct {
	uncountable       map[string]bool
	uncountableRegex  []*regexp.Regexp
	irregularPlural   map[string]string
	irregularSingular map[string]string
	plural            []Rule
	singular          []Rule
}

func (r *RuleSet) isUncountable(word string) bool {
	lowerWord := strings.ToLower(word)
	if _, exists := r.uncountable[lowerWord]; exists {
		return true
	}

	for _, pattern := range r.uncountableRegex {
		if pattern.MatchString(word) {
			return true
		}
	}
	return false
}
