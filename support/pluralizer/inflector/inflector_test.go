package inflector

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support/pluralizer/english"
)

func TestInflectorPlural(t *testing.T) {
	inflector := New(english.New())

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
	inflector := New(english.New())

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
	inflector := New(english.New())

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
	inflector := New(english.New())

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
	inflector := New(english.New())

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

func TestEnglishIrregularPlurals(t *testing.T) {
	lang := english.New()
	infl := New(lang)

	irregulars := map[string]string{
		"person":     "people",
		"child":      "children",
		"foot":       "feet",
		"tooth":      "teeth",
		"man":        "men",
		"woman":      "women",
		"goose":      "geese",
		"mouse":      "mice",
		"ox":         "oxen",
		"axis":       "axes",
		"crisis":     "crises",
		"thesis":     "theses",
		"phenomenon": "phenomena",
		"criterion":  "criteria",
		"medium":     "media",
		"alumnus":    "alumni",
		"cactus":     "cacti",
		"fungus":     "fungi",
		"corpus":     "corpora",
		"genus":      "genera",
		"stimulus":   "stimuli",
		"syllabus":   "syllabi",
		"synopsis":   "synopses",
		"atlas":      "atlases",
		"lens":       "lenses",
		"octopus":    "octopuses",
		"plateau":    "plateaux",
		"chateau":    "chateaux",
	}

	for singular, plural := range irregulars {
		assert.Equal(t, plural, infl.Plural(singular))
		assert.Equal(t, singular, infl.Singular(plural))
	}
}

func TestEnglishRegularPatterns(t *testing.T) {
	lang := english.New()
	infl := New(lang)

	tests := []struct {
		rule     string
		singular string
		plural   string
	}{
		{"words ending in s/x/z/ch/sh", "bus", "buses"},
		{"words ending in s/x/z/ch/sh", "box", "boxes"},
		{"words ending in s/x/z/ch/sh", "buzz", "buzzes"},
		{"words ending in s/x/z/ch/sh", "church", "churches"},
		{"words ending in s/x/z/ch/sh", "dish", "dishes"},

		{"words ending in consonant+y", "city", "cities"},
		{"words ending in consonant+y", "baby", "babies"},
		{"words ending in consonant+y", "party", "parties"},
		{"words ending in consonant+y", "company", "companies"},
		{"words ending in vowel+y", "toy", "toys"},
		{"words ending in vowel+y", "boy", "boys"},

		{"words ending in f/fe", "life", "lives"},
		{"words ending in f/fe", "wife", "wives"},
		{"words ending in f/fe", "wolf", "wolves"},
		{"words ending in f/fe", "shelf", "shelves"},
		{"words ending in f/fe", "leaf", "leaves"},

		{"words ending in o", "hero", "heroes"},
		{"words ending in o", "potato", "potatoes"},
		{"words ending in o", "tomato", "tomatoes"},

		{"regular words", "book", "books"},
		{"regular words", "table", "tables"},
		{"regular words", "car", "cars"},
	}

	for _, test := range tests {
		assert.Equal(t, test.plural, infl.Plural(test.singular))
		assert.Equal(t, test.singular, infl.Singular(test.plural))
	}
}

func TestEnglishUncountableWords(t *testing.T) {
	lang := english.New()
	infl := New(lang)

	uncountable := []string{
		"sheep", "deer", "fish", "moose", "swine",
		"bison", "salmon", "trout", "species",
		"information", "equipment", "money", "advice",
		"news", "data",
		"furniture", "luggage", "baggage", "butter",
		"research", "traffic", "weather",
	}

	for _, word := range uncountable {
		assert.Equal(t, word, infl.Plural(word))
		assert.Equal(t, word, infl.Singular(word))
	}
}

func TestEnglishCompoundWords(t *testing.T) {
	lang := english.New()
	infl := New(lang)

	compounds := map[string]string{
		"son-in-law":      "sons-in-law",
		"daughter-in-law": "daughters-in-law",
		"runner-up":       "runners-up",
		"passer-by":       "passers-by",
		"mother-in-law":   "mothers-in-law",
	}

	for singular, plural := range compounds {
		assert.Equal(t, plural, infl.Plural(singular))
		assert.Equal(t, singular, infl.Singular(plural))
	}
}

func TestEnglishSpecialCases(t *testing.T) {
	lang := english.New()
	infl := New(lang)

	tests := []struct {
		singular string
		plural   string
	}{
		{"quiz", "quizzes"},
		{"analysis", "analyses"},
		{"basis", "bases"},
		{"diagnosis", "diagnoses"},
		{"hypothesis", "hypotheses"},
		{"oasis", "oases"},
		{"parenthesis", "parentheses"},
		{"synopsis", "synopses"},
		{"bacterium", "bacteria"},
		{"curriculum", "curricula"},
		{"erratum", "errata"},
		{"memorandum", "memoranda"},
		{"millennium", "millennia"},
		{"stratum", "strata"},
		{"symposium", "symposia"},
	}

	for _, test := range tests {
		assert.Equal(t, test.plural, infl.Plural(test.singular))
		assert.Equal(t, test.singular, infl.Singular(test.plural))
	}
}
