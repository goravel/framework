package translation

import (
	"context"
	"testing"

	translationcontract "github.com/goravel/framework/contracts/translation"
	mockloader "github.com/goravel/framework/mocks/translation"
	"github.com/stretchr/testify/suite"
)

type TranslatorTestSuite struct {
	suite.Suite
	mockLoader *mockloader.Loader
	ctx        context.Context
}

func TestTranslatorTestSuite(t *testing.T) {
	suite.Run(t, &TranslatorTestSuite{})
}

func (t *TranslatorTestSuite) SetupTest() {
	t.mockLoader = mockloader.NewLoader(t.T())
	t.ctx = context.Background()
}

func (t *TranslatorTestSuite) TestGet() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "*", "en").Once().Return(map[string]map[string]string{
		"en": {
			"foo": "one",
		},
	}, nil)
	translation, err := translator.Get("foo")
	t.NoError(err)
	t.Equal("one", translation)

	// Get json replacement
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "*", "en").Once().Return(map[string]map[string]string{
		"en": {
			"foo": "Hello, :name! Welcome to :location.",
		},
	}, nil)
	translation, err = translator.Get("foo", translationcontract.Option{
		Replace: map[string]string{
			"name":     "krishan",
			"location": "india",
		},
	})
	t.NoError(err)
	t.Equal("Hello, krishan! Welcome to india.", translation)

	// test atomic replacements
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "*", "en").Once().Return(map[string]map[string]string{
		"en": {
			"foo": "Hello, :foo!",
		},
	}, nil)
	translation, err = translator.Get("foo", translationcontract.Option{
		Replace: map[string]string{
			"foo": "baz:bar",
			"bar": "abcdef",
		},
	})
	t.NoError(err)
	t.Equal("Hello, baz:bar!", translation)

	// preserve order of replacements
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "*", "en").Once().Return(map[string]map[string]string{
		"en": {
			"foo": ":greeting :name",
		},
	}, nil)
	translation, err = translator.Get("foo", translationcontract.Option{
		Replace: map[string]string{
			"name":     "krishan",
			"greeting": "Hello",
		},
	})
	t.NoError(err)
	t.Equal("Hello krishan", translation)

	// non-existing json key looks for regular keys
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "foo", "en").Once().Return(map[string]map[string]string{
		"en": {
			"bar": "one",
		},
	}, nil)
	translation, err = translator.Get("foo.bar")
	t.NoError(err)
	t.Equal("one", translation)

	// empty fallback
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "*", "en").Once().Return(map[string]map[string]string{}, nil)
	translation, err = translator.Get("foo")
	t.NoError(err)
	t.Equal("foo", translation)
}

func (t *TranslatorTestSuite) TestChoice() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "*", "en").Once().Return(map[string]map[string]string{
		"en": {
			"foo": "{0} first|{1}second",
		},
	}, nil)
	translation, err := translator.Choice("foo", 1)
	t.NoError(err)
	t.Equal("second", translation)

	// test atomic replacements
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "*", "en").Once().Return(map[string]map[string]string{
		"en": {
			"foo": "{0} first|{1}Hello, :foo!",
		},
	}, nil)
	translation, err = translator.Choice("foo", 1, translationcontract.Option{
		Replace: map[string]string{
			"foo": "baz:bar",
			"bar": "abcdef",
		},
	})
	t.NoError(err)
	t.Equal("Hello, baz:bar!", translation)
}

func (t *TranslatorTestSuite) TestHas() {

}

func (t *TranslatorTestSuite) TestGetLocale() {

}

func (t *TranslatorTestSuite) TestSetLocale() {

}

func (t *TranslatorTestSuite) TestGetFallback() {

}

func (t *TranslatorTestSuite) TestSetFallback() {

}

func (t *TranslatorTestSuite) TestLoad() {

}

func (t *TranslatorTestSuite) TestIsLoaded() {

}

func (t *TranslatorTestSuite) TestMakeReplacements() {

}

func (t *TranslatorTestSuite) TestParseKey() {

}
