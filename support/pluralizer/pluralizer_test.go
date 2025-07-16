package pluralizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlural(t *testing.T) {
	tests := []struct {
		singular string
		plural   string
	}{
		// Regular plurals
		{"book", "books"},
		{"table", "tables"},
		{"chair", "chairs"},
		{"car", "cars"},
		{"house", "houses"},

		// Irregular plurals
		{"person", "people"},
		{"child", "children"},
		{"goose", "geese"},
		{"mouse", "mice"},
		{"ox", "oxen"},
		{"leaf", "leaves"},
		{"foot", "feet"},
		{"tooth", "teeth"},
		{"woman", "women"},
		{"man", "men"},

		// Words ending in 'y'
		{"city", "cities"},
		{"baby", "babies"},
		{"toy", "toys"},
		{"boy", "boys"},

		// Words ending in 's', 'x', 'z', 'ch', 'sh'
		{"bus", "buses"},
		{"box", "boxes"},
		{"buzz", "buzzes"},
		{"church", "churches"},
		{"dish", "dishes"},

		// Words ending in 'f' or 'fe'
		{"life", "lives"},
		{"wife", "wives"},
		{"wolf", "wolves"},
		{"shelf", "shelves"},

		// Uncountable words
		{"fish", "fish"},
		{"sheep", "sheep"},
		{"deer", "deer"},
		{"information", "information"},
		{"rice", "rice"},
		{"equipment", "equipment"},

		// Special cases
		{"analysis", "analyses"},
		{"criterion", "criteria"},
		{"datum", "data"},
		{"medium", "media"},
		{"phenomenon", "phenomena"},
	}

	for _, test := range tests {
		result := Plural(test.singular)
		assert.Equal(t, test.plural, result, "Plural(%s) should return %s, got %s", test.singular, test.plural, result)
	}

	inflector := New()
	for _, test := range tests {
		result := inflector.Plural(test.singular)
		assert.Equal(t, test.plural, result, "inflector.Plural(%s) should return %s, got %s", test.singular, test.plural, result)
	}

	assert.Equal(t, "Books", Plural("Book"))
	assert.Equal(t, "BOOKS", Plural("BOOK"))
}

func TestSingular(t *testing.T) {
	tests := []struct {
		plural   string
		singular string
	}{
		// Regular singulars
		{"books", "book"},
		{"tables", "table"},
		{"chairs", "chair"},
		{"cars", "car"},
		{"houses", "house"},

		// Irregular singulars
		{"people", "person"},
		{"children", "child"},
		{"geese", "goose"},
		{"mice", "mouse"},
		{"oxen", "ox"},
		{"leaves", "leaf"},
		{"feet", "foot"},
		{"teeth", "tooth"},
		{"women", "woman"},
		{"men", "man"},

		// Words ending in 'ies'
		{"cities", "city"},
		{"babies", "baby"},
		{"toys", "toy"},
		{"boys", "boy"},

		// Words ending in 'es'
		{"buses", "bus"},
		{"boxes", "box"},
		{"buzzes", "buzz"},
		{"churches", "church"},
		{"dishes", "dish"},

		// Words ending in 'ves'
		{"lives", "life"},
		{"wives", "wife"},
		{"wolves", "wolf"},
		{"shelves", "shelf"},

		// Uncountable words
		{"fish", "fish"},
		{"sheep", "sheep"},
		{"deer", "deer"},
		{"information", "information"},
		{"rice", "rice"},
		{"equipment", "equipment"},

		// Special cases
		{"analyses", "analysis"},
		{"criteria", "criterion"},
		{"data", "datum"},
		{"media", "medium"},
		{"phenomena", "phenomenon"},
	}

	for _, test := range tests {
		result := Singular(test.plural)
		assert.Equal(t, test.singular, result, "Singular(%s) should return %s, got %s", test.plural, test.singular, result)
	}

	inflector := New()
	for _, test := range tests {
		result := inflector.Singular(test.plural)
		assert.Equal(t, test.singular, result, "inflector.Singular(%s) should return %s, got %s", test.plural, test.singular, result)
	}

	assert.Equal(t, "Book", Singular("Books"))
	assert.Equal(t, "BOOK", Singular("BOOKS"))
}

func TestNewForLanguage(t *testing.T) {
	inflectorEn := NewForLanguage("en")
	inflectorDefault := NewForLanguage("unsupported")

	tests := []struct {
		singular string
		plural   string
	}{
		{"book", "books"},
		{"person", "people"},
	}

	for _, test := range tests {
		assert.Equal(t, test.plural, inflectorEn.Plural(test.singular))
		assert.Equal(t, test.plural, inflectorDefault.Plural(test.singular))
	}
}
