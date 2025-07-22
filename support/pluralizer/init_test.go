package pluralizer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/support/pluralizer"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/pluralizer/rules"
)

func TestDefaultLanguage(t *testing.T) {
	assert.Equal(t, LanguageEnglish, GetLanguage().Name())
}

func TestUseLanguage(t *testing.T) {
	originalLang := GetLanguage().Name()
	defer func() {
		assert.Nil(t, UseLanguage(originalLang))
	}()

	err := UseLanguage(LanguageEnglish)
	assert.Nil(t, err)
	assert.Equal(t, LanguageEnglish, GetLanguage().Name())

	err = UseLanguage("nonexistent")
	assert.NotNil(t, err)
	assert.Equal(t, LanguageEnglish, GetLanguage().Name())
	assert.ErrorIs(t, err, errors.PluralizerLanguageNotFound)
}

func TestRegisterLanguage(t *testing.T) {
	originalLang := GetLanguage().Name()
	defer func() {
		assert.Nil(t, UseLanguage(originalLang))
	}()

	mockLang := newMockLanguage("test")
	err := RegisterLanguage(mockLang)
	assert.Nil(t, err)

	err = UseLanguage("test")
	assert.Nil(t, err)
	assert.Equal(t, "test", GetLanguage().Name())

	err = RegisterLanguage(nil)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, errors.PluralizerEmptyLanguageName)

	emptyLang := newMockLanguage("")
	err = RegisterLanguage(emptyLang)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, errors.PluralizerEmptyLanguageName)
}

func TestRegisterIrregular(t *testing.T) {
	sub1 := rules.NewSubstitution("test", "tests")
	sub2 := rules.NewSubstitution("exam", "exams")

	err := RegisterIrregular(LanguageEnglish, sub1, sub2)
	assert.Nil(t, err)

	assert.Equal(t, "tests", Plural("test"))
	assert.Equal(t, "test", Singular("tests"))
	assert.Equal(t, "exams", Plural("exam"))
	assert.Equal(t, "exam", Singular("exams"))

	err = RegisterIrregular("nonexistent", sub1)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, errors.PluralizerLanguageNotFound)

	err = RegisterIrregular(LanguageEnglish)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, errors.PluralizerNoSubstitutionsGiven)
}

func TestRegisterUninflected(t *testing.T) {
	err := RegisterUninflected(LanguageEnglish, "testdata", "metadata")
	assert.Nil(t, err)

	assert.Equal(t, "testdata", Plural("testdata"))
	assert.Equal(t, "testdata", Singular("testdata"))
	assert.Equal(t, "metadata", Plural("metadata"))
	assert.Equal(t, "metadata", Singular("metadata"))

	err = RegisterUninflected("nonexistent", "data")
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, errors.PluralizerLanguageNotFound)

	err = RegisterUninflected(LanguageEnglish)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, errors.PluralizerNoWordsGiven)
}

func TestRegisterPluralUninflected(t *testing.T) {
	err := RegisterPluralUninflected(LanguageEnglish, "pluraldata")
	assert.Nil(t, err)

	assert.Equal(t, "pluraldata", Plural("pluraldata"))

	err = RegisterPluralUninflected("nonexistent", "data")
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, errors.PluralizerLanguageNotFound)

	err = RegisterPluralUninflected(LanguageEnglish)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, errors.PluralizerNoWordsGiven)
}

func TestRegisterSingularUninflected(t *testing.T) {
	err := RegisterSingularUninflected(LanguageEnglish, "singulardata")
	assert.Nil(t, err)
	assert.Equal(t, "singulardata", Singular("singulardata"))

	err = RegisterSingularUninflected("nonexistent", "data")
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, errors.PluralizerLanguageNotFound)

	err = RegisterSingularUninflected(LanguageEnglish)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, errors.PluralizerNoWordsGiven)
}

func TestGlobalPluralFunction(t *testing.T) {
	result := Plural("book")
	assert.Equal(t, "books", result)
}

func TestGlobalSingularFunction(t *testing.T) {
	result := Singular("books")
	assert.Equal(t, "book", result)
}

func TestComplexWorkflow(t *testing.T) {
	originalLang := GetLanguage().Name()
	defer func() {
		assert.Nil(t, UseLanguage(originalLang))
	}()

	testLang := newMockLanguage("testlang")
	err := RegisterLanguage(testLang)
	assert.Nil(t, err)

	err = UseLanguage("testlang")
	assert.Nil(t, err)
	assert.Equal(t, "testlang", GetLanguage().Name())

	err = RegisterIrregular("testlang", rules.NewSubstitution("testword", "testwords"))
	assert.Nil(t, err)

	err = RegisterUninflected("testlang", "staticword")
	assert.Nil(t, err)

	assert.Equal(t, "testwords", Plural("testword"))
	assert.Equal(t, "testword", Singular("testwords"))
	assert.Equal(t, "staticword", Plural("staticword"))
	assert.Equal(t, "staticword", Singular("staticword"))

	err = UseLanguage(LanguageEnglish)
	assert.Nil(t, err)

	err = RegisterIrregular(LanguageEnglish, rules.NewSubstitution("workflowtest", "workflowtests"))
	assert.Nil(t, err)

	assert.Equal(t, "workflowtests", Plural("workflowtest"))
	assert.Equal(t, "workflowtest", Singular("workflowtests"))
	assert.Equal(t, "books", Plural("book"))
}

func TestEdgeCases(t *testing.T) {
	result := Plural("")
	assert.Equal(t, "", result)

	result = Singular("")
	assert.Equal(t, "", result)
	result = Plural("Book")
	assert.Equal(t, "Books", result)

	result = Plural("test-case")
	assert.NotEqual(t, "", result)
}

func TestErrorReturns(t *testing.T) {
	assert.Nil(t, UseLanguage(LanguageEnglish))
	assert.Nil(t, RegisterLanguage(newMockLanguage("testreturn")))
	assert.Nil(t, RegisterIrregular(LanguageEnglish, rules.NewSubstitution("a", "as")))
	assert.Nil(t, RegisterUninflected(LanguageEnglish, "testword"))
	assert.Nil(t, RegisterPluralUninflected(LanguageEnglish, "testword2"))
	assert.Nil(t, RegisterSingularUninflected(LanguageEnglish, "testword3"))

	assert.NotNil(t, UseLanguage("nonexistent"))
	assert.NotNil(t, RegisterLanguage(nil))
	assert.NotNil(t, RegisterIrregular("nonexistent", rules.NewSubstitution("a", "as")))
	assert.NotNil(t, RegisterUninflected("nonexistent", "testword"))
	assert.NotNil(t, RegisterPluralUninflected(LanguageEnglish))
}

type mockLanguage struct {
	name            string
	pluralRuleset   pluralizer.Ruleset
	singularRuleset pluralizer.Ruleset
}

func newMockLanguage(name string) *mockLanguage {
	return &mockLanguage{
		name:            name,
		pluralRuleset:   rules.NewRuleset(nil, nil, nil),
		singularRuleset: rules.NewRuleset(nil, nil, nil),
	}
}

func (m *mockLanguage) Name() string {
	return m.name
}

func (m *mockLanguage) PluralRuleset() pluralizer.Ruleset {
	return m.pluralRuleset
}

func (m *mockLanguage) SingularRuleset() pluralizer.Ruleset {
	return m.singularRuleset
}
