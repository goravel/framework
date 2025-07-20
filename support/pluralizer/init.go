package pluralizer

import (
	"github.com/goravel/framework/contracts/support/pluralizer"
	"github.com/goravel/framework/support/pluralizer/english"
	"github.com/goravel/framework/support/pluralizer/inflector"
	"github.com/goravel/framework/support/pluralizer/rules"
)

var (
	instance         pluralizer.Inflector
	currentLanguage  string
	inflectorFactory = map[string]pluralizer.Inflector{
		"english": inflector.New(english.New()),
	}
)

func init() {
	currentLanguage = "english"
	instance = inflectorFactory[currentLanguage]
}

func UseLanguage(lang string) bool {
	if factory, exists := inflectorFactory[lang]; exists {
		currentLanguage = lang
		instance = factory
		return true
	}
	return false
}

func GetLanguage() string {
	return currentLanguage
}

func RegisterLanguage(language pluralizer.Language) bool {
	if language == nil || language.Name() == "" {
		return false
	}

	inflectorFactory[language.Name()] = inflector.New(language)
	return true
}

func getLanguageInstance(lang string) (pluralizer.Language, pluralizer.Inflector, bool) {
	factory, exists := inflectorFactory[lang]
	if !exists {
		return nil, nil, false
	}

	language := factory.Language()
	return language, factory, true
}

func RegisterIrregular(lang string, substitutions ...pluralizer.Substitution) bool {
	if len(substitutions) == 0 {
		return false
	}

	language, factory, exists := getLanguageInstance(lang)
	if !exists {
		return false
	}

	language.PluralRuleset().AddIrregular(substitutions...)

	flipped := rules.GetFlippedSubstitutions(substitutions...)
	language.SingularRuleset().AddIrregular(flipped...)

	factory.SetLanguage(language)
	return true
}

func RegisterUninflected(lang string, words ...string) bool {
	if len(words) == 0 {
		return false
	}

	language, factory, exists := getLanguageInstance(lang)
	if !exists {
		return false
	}

	language.PluralRuleset().AddUninflected(words...)
	language.SingularRuleset().AddUninflected(words...)

	factory.SetLanguage(language)
	return true
}

func RegisterPluralUninflected(lang string, words ...string) bool {
	if len(words) == 0 {
		return false
	}

	language, factory, exists := getLanguageInstance(lang)
	if !exists {
		return false
	}

	language.PluralRuleset().AddUninflected(words...)
	factory.SetLanguage(language)
	return true
}

func RegisterSingularUninflected(lang string, words ...string) bool {
	if len(words) == 0 {
		return false
	}

	language, factory, exists := getLanguageInstance(lang)
	if !exists {
		return false
	}

	language.SingularRuleset().AddUninflected(words...)
	factory.SetLanguage(language)
	return true
}

func Plural(word string) string {
	return instance.Plural(word)
}

func Singular(word string) string {
	return instance.Singular(word)
}
