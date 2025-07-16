package pluralizer

import "sync"

type Pluralizer interface {
	Plural(word string) string
	Singular(word string) string
}

var (
	instance Pluralizer
	once     sync.Once
)

func init() {
	once.Do(func() {
		instance = newEnglishInflector()
	})
}

func New() Pluralizer {
	return NewForLanguage("en")
}

func NewForLanguage(lang string) Pluralizer {
	switch lang {
	case "en":
		return newEnglishInflector()
	default:
		return newEnglishInflector()
	}
}

func Plural(word string) string {
	return instance.Plural(word)
}

func Singular(word string) string {
	return instance.Singular(word)
}
