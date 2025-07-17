package pluralizer

import (
	"testing"

	"github.com/goravel/framework/contracts/support/pluralizer"
	"github.com/stretchr/testify/assert"
)

func TestDefaultLanguage(t *testing.T) {
	assert.Equal(t, "english", GetLanguage())
}

func TestUseLanguage(t *testing.T) {
	originalLang := GetLanguage()
	defer func() {
		UseLanguage(originalLang)
	}()

	UseLanguage("english")
	assert.Equal(t, "english", GetLanguage())

	UseLanguage("nonexistent")
	assert.Equal(t, "english", GetLanguage())
}

func TestRegisterLanguage(t *testing.T) {
	originalLang := GetLanguage()
	defer func() {
		UseLanguage(originalLang)
	}()

	mockLang := &mockLanguage{name: "test"}
	RegisterLanguage(mockLang)

	UseLanguage("test")
	assert.Equal(t, "test", GetLanguage())
}

func TestRegisterNilLanguage(t *testing.T) {
	RegisterLanguage(nil)
	assert.Equal(t, "english", GetLanguage())
}

func TestGlobalPluralFunction(t *testing.T) {
	result := Plural("book")
	assert.Equal(t, "books", result)
}

func TestGlobalSingularFunction(t *testing.T) {
	result := Singular("books")
	assert.Equal(t, "book", result)
}

type mockLanguage struct {
	name string
}

func (m *mockLanguage) Name() string {
	return m.name
}

func (m *mockLanguage) PluralRuleset() pluralizer.Ruleset {
	return NewRuleset(nil, nil, nil)
}

func (m *mockLanguage) SingularRuleset() pluralizer.Ruleset {
	return NewRuleset(nil, nil, nil)
}
