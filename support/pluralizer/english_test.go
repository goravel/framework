package pluralizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnglishIrregularPlurals(t *testing.T) {
	lang := NewEnglishLanguage()
	inflector := NewInflector(lang)

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
		"gas":        "gases",
		"lens":       "lenses",
		"octopus":    "octopuses",
		"plateau":    "plateaux",
		"chateau":    "chateaux",
	}

	for singular, plural := range irregulars {
		assert.Equal(t, plural, inflector.Plural(singular))
		assert.Equal(t, singular, inflector.Singular(plural))
	}
}

func TestEnglishRegularPatterns(t *testing.T) {
	lang := NewEnglishLanguage()
	inflector := NewInflector(lang)

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
		assert.Equal(t, test.plural, inflector.Plural(test.singular))
		assert.Equal(t, test.singular, inflector.Singular(test.plural))
	}
}

func TestEnglishUncountableWords(t *testing.T) {
	lang := NewEnglishLanguage()
	inflector := NewInflector(lang)

	uncountable := []string{
		"sheep", "deer", "fish", "moose", "swine",
		"bison", "salmon", "trout", "species",
		"information", "equipment", "money", "advice",
		"news", "data",
		"furniture", "luggage", "baggage",
		"bread", "butter", "cheese", "milk",
		"research", "traffic", "weather",
	}

	for _, word := range uncountable {
		assert.Equal(t, word, inflector.Plural(word))
		assert.Equal(t, word, inflector.Singular(word))
	}
}

func TestEnglishCompoundWords(t *testing.T) {
	lang := NewEnglishLanguage()
	inflector := NewInflector(lang)

	compounds := map[string]string{
		"son-in-law":      "sons-in-law",
		"daughter-in-law": "daughters-in-law",
		"runner-up":       "runners-up",
		"passer-by":       "passers-by",
		"mother-in-law":   "mothers-in-law",
	}

	for singular, plural := range compounds {
		assert.Equal(t, plural, inflector.Plural(singular))
		assert.Equal(t, singular, inflector.Singular(plural))
	}
}

func TestEnglishSpecialCases(t *testing.T) {
	lang := NewEnglishLanguage()
	inflector := NewInflector(lang)

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
		assert.Equal(t, test.plural, inflector.Plural(test.singular))
		assert.Equal(t, test.singular, inflector.Singular(test.plural))
	}
}

func TestEnglishRulesetStructure(t *testing.T) {
	lang := NewEnglishLanguage()

	assert.Equal(t, "english", lang.Name())
	assert.NotNil(t, lang.PluralRuleset())
	assert.NotNil(t, lang.SingularRuleset())

	pluralRules := lang.PluralRuleset()
	singularRules := lang.SingularRuleset()

	assert.True(t, len(pluralRules.Irregular()) > 0)
	assert.True(t, len(singularRules.Irregular()) > 0)
	assert.True(t, len(pluralRules.Regular()) > 0)
	assert.True(t, len(singularRules.Regular()) > 0)
	assert.True(t, len(pluralRules.Uninflected()) > 0)
	assert.True(t, len(singularRules.Uninflected()) > 0)
}
