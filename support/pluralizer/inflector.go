package pluralizer

import (
	"strings"
	"unicode"
)

type inflector struct {
	rules *RuleSet
}

func (r *inflector) Plural(word string) string {
	if word == "" || r.rules.isUncountable(word) {
		return word
	}

	for _, plural := range r.rules.irregularPlural {
		if strings.EqualFold(word, plural) {
			return word
		}
	}

	for singular, plural := range r.rules.irregularPlural {
		if strings.EqualFold(word, singular) {
			return r.matchCase(plural, word)
		}
	}

	for _, rule := range r.rules.plural {
		if rule.Pattern.MatchString(word) {
			return r.matchCase(rule.Pattern.ReplaceAllString(word, rule.Replacement), word)
		}
	}

	return word
}

func (r *inflector) Singular(word string) string {
	if word == "" || r.rules.isUncountable(word) {
		return word
	}

	for plural, singular := range r.rules.irregularSingular {
		if strings.EqualFold(word, plural) {
			return r.matchCase(singular, word)
		}
	}

	for _, rule := range r.rules.singular {
		if rule.Pattern.MatchString(word) {
			return r.matchCase(rule.Pattern.ReplaceAllString(word, rule.Replacement), word)
		}
	}

	return word
}

func (r *inflector) matchCase(value, comparison string) string {
	if len(comparison) == 0 {
		return value
	}

	isAllUpper := true
	for _, r := range comparison {
		if !unicode.IsUpper(r) {
			isAllUpper = false
			break
		}
	}
	if isAllUpper {
		return strings.ToUpper(value)
	}

	firstChar := rune(comparison[0])
	if unicode.IsUpper(firstChar) {
		if len(value) > 0 {
			vRunes := []rune(value)
			return string(unicode.ToUpper(vRunes[0])) + string(vRunes[1:])
		}
	}

	return value
}
