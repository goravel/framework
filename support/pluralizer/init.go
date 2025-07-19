package pluralizer

import (
	"github.com/goravel/framework/contracts/support/pluralizer"
	"github.com/goravel/framework/support/pluralizer/english"
	"github.com/goravel/framework/support/pluralizer/inflector"
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

func UseLanguage(lang string) {
	if _, ok := inflectorFactory[lang]; ok {
		currentLanguage = lang
		instance = inflectorFactory[lang]
	}
}

func GetLanguage() string {
	return currentLanguage
}

func RegisterLanguage(language pluralizer.Language) {
	if language == nil {
		return
	}

	inflectorFactory[language.Name()] = inflector.New(language)
}

func Plural(word string) string {
	return instance.Plural(word)
}

func Singular(word string) string {
	return instance.Singular(word)
}
