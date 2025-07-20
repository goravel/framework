package pluralizer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/support/pluralizer"
	"github.com/goravel/framework/support/pluralizer/rules"
)

func TestDefaultLanguage(t *testing.T) {
	assert.Equal(t, "english", GetLanguage())
}

func TestUseLanguage(t *testing.T) {
	originalLang := GetLanguage()
	defer func() {
		UseLanguage(originalLang)
	}()

	success := UseLanguage("english")
	assert.True(t, success)
	assert.Equal(t, "english", GetLanguage())

	success = UseLanguage("nonexistent")
	assert.False(t, success)
	assert.Equal(t, "english", GetLanguage())
}

func TestRegisterLanguage(t *testing.T) {
	originalLang := GetLanguage()
	defer func() {
		UseLanguage(originalLang)
	}()

	mockLang := newMockLanguage("test")
	success := RegisterLanguage(mockLang)
	assert.True(t, success)

	success = UseLanguage("test")
	assert.True(t, success)
	assert.Equal(t, "test", GetLanguage())

	success = RegisterLanguage(nil)
	assert.False(t, success)

	emptyLang := newMockLanguage("")
	success = RegisterLanguage(emptyLang)
	assert.False(t, success)
}

func TestRegisterIrregular(t *testing.T) {
	sub1 := rules.NewSubstitution("test", "tests")
	sub2 := rules.NewSubstitution("exam", "exams")

	success := RegisterIrregular("english", sub1, sub2)
	assert.True(t, success)

	assert.Equal(t, "tests", Plural("test"))
	assert.Equal(t, "test", Singular("tests"))
	assert.Equal(t, "exams", Plural("exam"))
	assert.Equal(t, "exam", Singular("exams"))

	success = RegisterIrregular("nonexistent", sub1)
	assert.False(t, success)

	success = RegisterIrregular("english")
	assert.False(t, success)
}

func TestRegisterUninflected(t *testing.T) {
	success := RegisterUninflected("english", "testdata", "metadata")
	assert.True(t, success)

	assert.Equal(t, "testdata", Plural("testdata"))
	assert.Equal(t, "testdata", Singular("testdata"))
	assert.Equal(t, "metadata", Plural("metadata"))
	assert.Equal(t, "metadata", Singular("metadata"))

	success = RegisterUninflected("nonexistent", "data")
	assert.False(t, success)

	success = RegisterUninflected("english")
	assert.False(t, success)
}

func TestRegisterPluralUninflected(t *testing.T) {
	success := RegisterPluralUninflected("english", "pluraldata")
	assert.True(t, success)

	assert.Equal(t, "pluraldata", Plural("pluraldata"))

	success = RegisterPluralUninflected("nonexistent", "data")
	assert.False(t, success)

	success = RegisterPluralUninflected("english")
	assert.False(t, success)
}

func TestRegisterSingularUninflected(t *testing.T) {
	success := RegisterSingularUninflected("english", "singulardata")
	assert.True(t, success)
	assert.Equal(t, "singulardata", Singular("singulardata"))

	success = RegisterSingularUninflected("nonexistent", "data")
	assert.False(t, success)

	success = RegisterSingularUninflected("english")
	assert.False(t, success)
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
	originalLang := GetLanguage()
	defer func() {
		UseLanguage(originalLang)
	}()

	testLang := newMockLanguage("testlang")
	success := RegisterLanguage(testLang)
	assert.True(t, success)

	success = UseLanguage("testlang")
	assert.True(t, success)
	assert.Equal(t, "testlang", GetLanguage())

	success = RegisterIrregular("testlang", rules.NewSubstitution("testword", "testwords"))
	assert.True(t, success)

	success = RegisterUninflected("testlang", "staticword")
	assert.True(t, success)

	assert.Equal(t, "testwords", Plural("testword"))
	assert.Equal(t, "testword", Singular("testwords"))
	assert.Equal(t, "staticword", Plural("staticword"))
	assert.Equal(t, "staticword", Singular("staticword"))

	success = UseLanguage("english")
	assert.True(t, success)

	success = RegisterIrregular("english", rules.NewSubstitution("workflowtest", "workflowtests"))
	assert.True(t, success)

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

func TestReturnValues(t *testing.T) {
	assert.True(t, UseLanguage("english"))
	assert.True(t, RegisterLanguage(newMockLanguage("testreturn")))
	assert.True(t, RegisterIrregular("english", rules.NewSubstitution("a", "as")))
	assert.True(t, RegisterUninflected("english", "testword"))
	assert.True(t, RegisterPluralUninflected("english", "testword2"))
	assert.True(t, RegisterSingularUninflected("english", "testword3"))

	assert.False(t, UseLanguage("nonexistent"))
	assert.False(t, RegisterLanguage(nil))
	assert.False(t, RegisterIrregular("nonexistent", rules.NewSubstitution("a", "as")))
	assert.False(t, RegisterUninflected("nonexistent", "testword"))
	assert.False(t, RegisterPluralUninflected("english"))
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
