package pluralizer

import (
	"strings"
	"unicode"

	"github.com/goravel/framework/contracts/support/pluralizer"
)

type Inflector struct {
	language pluralizer.Language
}

func NewInflector(language pluralizer.Language) pluralizer.Inflector {
	return &Inflector{
		language: language,
	}
}

func (r *Inflector) Plural(word string) string {
	return r.inflect(word, r.language.PluralRuleset())
}

func (r *Inflector) Singular(word string) string {
	return r.inflect(word, r.language.SingularRuleset())
}

func (r *Inflector) inflect(word string, ruleset pluralizer.Ruleset) string {
	if word == "" {
		return ""
	}

	if ruleset.IsUncountable(word) {
		return word
	}

	// Check if word is already in target form (To)
	for _, substitution := range ruleset.Irregular() {
		if strings.EqualFold(word, substitution.To()) {
			return word
		}
	}

	// Check if word is in source form (From) and convert to target form (To)
	for _, substitution := range ruleset.Irregular() {
		if strings.EqualFold(word, substitution.From()) {
			return matchCase(substitution.To(), word)
		}
	}

	for _, transformation := range ruleset.Regular() {
		if result := transformation.Apply(word); result != word {
			return matchCase(result, word)
		}
	}

	return word
}

func matchCase(value, comparison string) string {
	if len(comparison) == 0 {
		return value
	}

	isAllUpper := true
	for _, r := range comparison {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			isAllUpper = false
			break
		}
	}
	if isAllUpper {
		return strings.ToUpper(value)
	}

	isAllLower := true
	for _, r := range comparison {
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			isAllLower = false
			break
		}
	}
	if isAllLower {
		return strings.ToLower(value)
	}

	compRunes := []rune(comparison)
	if unicode.IsUpper(compRunes[0]) {
		vRunes := []rune(value)
		if len(vRunes) > 0 {
			return string(unicode.ToUpper(vRunes[0])) + strings.ToLower(string(vRunes[1:]))
		}
	}

	return value
}
