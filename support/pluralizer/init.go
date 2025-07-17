package pluralizer

import "github.com/goravel/framework/contracts/support/pluralizer"

var (
	instance         pluralizer.Inflector
	currentLanguage  string
	inflectorFactory = map[string]pluralizer.Inflector{
		"english": NewInflector(NewEnglishLanguage()),
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

	inflectorFactory[language.Name()] = NewInflector(language)
}

func Plural(word string) string {
	return instance.Plural(word)
}

func Singular(word string) string {
	return instance.Singular(word)
}
