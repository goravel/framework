package pluralizer

import "sync"

type Pluralizer interface {
	Plural(word string) string
	Singular(word string) string
}

type Language string

const (
	English Language = "en"
)

var (
	instance Pluralizer
	once     sync.Once

	factory = map[Language]func() Pluralizer{
		English: func() Pluralizer { return newEnglishInflector() },
	}
)

func init() {
	once.Do(func() {
		instance = newEnglishInflector()
	})
}

func New() Pluralizer {
	return NewForLanguage(English)
}

func NewForLanguage(lang Language) Pluralizer {
	if constructor, ok := factory[lang]; ok {
		return constructor()
	}
	return newEnglishInflector()
}

func Plural(word string) string {
	return instance.Plural(word)
}

func Singular(word string) string {
	return instance.Singular(word)
}
