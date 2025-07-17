package pluralizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInflectorPlural(t *testing.T) {
	inflector := NewInflector(NewEnglishLanguage())

	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"book", "books"},
		{"person", "people"},
		{"child", "children"},
		{"mouse", "mice"},
		{"sheep", "sheep"},
		{"data", "data"},
		{"city", "cities"},
		{"half", "halves"},
		{"quiz", "quizzes"},
		{"ox", "oxen"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, inflector.Plural(test.input))
	}
}

func TestInflectorSingular(t *testing.T) {
	inflector := NewInflector(NewEnglishLanguage())

	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"books", "book"},
		{"people", "person"},
		{"children", "child"},
		{"mice", "mouse"},
		{"sheep", "sheep"},
		{"data", "data"},
		{"cities", "city"},
		{"halves", "half"},
		{"quizzes", "quiz"},
		{"oxen", "ox"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, inflector.Singular(test.input))
	}
}

func TestInflectorCasePreservation(t *testing.T) {
	inflector := NewInflector(NewEnglishLanguage())

	tests := []struct {
		input    string
		expected string
		method   string
	}{
		{"BOOK", "BOOKS", "plural"},
		{"Book", "Books", "plural"},
		{"book", "books", "plural"},
		{"BOOKS", "BOOK", "singular"},
		{"Books", "Book", "singular"},
		{"PERSON", "PEOPLE", "plural"},
		{"Person", "People", "plural"},
		{"PEOPLE", "PERSON", "singular"},
		{"People", "Person", "singular"},
	}

	for _, test := range tests {
		if test.method == "plural" {
			assert.Equal(t, test.expected, inflector.Plural(test.input))
		} else {
			assert.Equal(t, test.expected, inflector.Singular(test.input))
		}
	}
}

func TestInflectorUncountableWords(t *testing.T) {
	inflector := NewInflector(NewEnglishLanguage())

	uncountable := []string{
		"fish", "sheep", "deer", "moose", "swine",
		"information", "equipment", "money", "advice",
		"software", "news", "data",
	}

	for _, word := range uncountable {
		assert.Equal(t, word, inflector.Plural(word))
		assert.Equal(t, word, inflector.Singular(word))
	}
}

func TestInflectorIrregularWords(t *testing.T) {
	inflector := NewInflector(NewEnglishLanguage())

	irregulars := map[string]string{
		"person": "people",
		"child":  "children",
		"foot":   "feet",
		"tooth":  "teeth",
		"goose":  "geese",
		"man":    "men",
		"woman":  "women",
		"mouse":  "mice",
		"ox":     "oxen",
	}

	for singular, plural := range irregulars {
		assert.Equal(t, plural, inflector.Plural(singular))
		assert.Equal(t, singular, inflector.Singular(plural))
	}
}
