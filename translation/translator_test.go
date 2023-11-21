package translation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	translationcontract "github.com/goravel/framework/contracts/translation"
	"github.com/goravel/framework/http"
	mockloader "github.com/goravel/framework/mocks/translation"
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
	t.mockLoader.On("Load", "*", "fr").Once().Return(map[string]map[string]string{
		"en": {
			"foo": "{0} first|{1}Hello, :foo!",
		},
	}, nil)
	translation, err = translator.Choice("foo", 1, translationcontract.Option{
		Replace: map[string]string{
			"foo": "baz:bar",
			"bar": "abcdef",
		},
		Locale: "fr",
	})
	t.NoError(err)
	t.Equal("Hello, baz:bar!", translation)

	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "*", "en").Once().Return(nil, errors.New("some error"))
	translation, err = translator.Choice("foo", 1)
	t.EqualError(err, "some error")
	t.Equal("", translation)
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

	// Case: when file exists but there is some error
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "*", "en").Once().Return(nil, errors.New("some error"))
	translation, err = translator.Get("foo")
	t.EqualError(err, "some error")
	t.Equal("", translation)

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

	// Case: Fallback to a different locale
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "fr")
	t.mockLoader.On("Load", "*", "en").Once().Return(map[string]map[string]string{}, nil)
	t.mockLoader.On("Load", "*", "fr").Once().Return(map[string]map[string]string{
		"fr": {
			"nonexistentKey": "French translation",
		},
	}, nil)
	translation, err = translator.Get("nonexistentKey", translationcontract.Option{
		Fallback: translationcontract.Bool(true),
		Locale:   "en",
	})
	t.NoError(err)
	t.Equal("French translation", translation)
}

func (t *TranslatorTestSuite) TestGetLocale() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en")

	// Case: Get locale initially set
	locale := translator.GetLocale()
	t.Equal("en", locale)

	// Case: Set locale using SetLocale and then get it
	translator.SetLocale("fr")
	locale = translator.GetLocale()
	t.Equal("fr", locale)
}

func (t *TranslatorTestSuite) TestGetFallback() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en")

	// Case: No explicit fallback set
	fallback := translator.GetFallback()
	t.Equal("en", fallback)

	// Case: Set fallback using SetFallback
	newCtx := translator.SetFallback("fr")
	fallback = translator.GetFallback()
	t.Equal("fr", fallback)
	t.Equal("fr", newCtx.Value(string(fallbackLocaleKey)))
}

func (t *TranslatorTestSuite) TestHas() {
	// Case: Key exists in translations
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "*", "en").Once().Return(map[string]map[string]string{
		"en": {
			"hello": "world",
		},
	}, nil)
	hasKey := translator.Has("hello")
	t.True(hasKey)

	// Case: Key does not exist in translations
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "*", "en").Once().Return(map[string]map[string]string{
		"en": {
			"name": "Bowen",
		},
	}, nil)
	hasKey = translator.Has("email")
	t.False(hasKey)

	// Case: Key exists, but translation is the same as the key
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "*", "en").Once().Return(map[string]map[string]string{
		"en": {
			"sameKey": "sameKey",
		},
	}, nil)
	hasKey = translator.Has("sameKey")
	t.False(hasKey)
}

func (t *TranslatorTestSuite) TestSetFallback() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en")

	// Case: Set fallback using SetFallback
	newCtx := translator.SetFallback("fr")
	t.Equal("fr", translator.fallback)
	t.Equal("fr", newCtx.Value(string(fallbackLocaleKey)))
}

func (t *TranslatorTestSuite) TestSetLocale() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en")

	// Case: Set locale using SetLocale
	newCtx := translator.SetLocale("fr")
	t.Equal("fr", translator.locale)
	t.Equal("fr", newCtx.Value(string(localeKey)))

	// Case: use http.Context
	translator = NewTranslator(http.Background(), t.mockLoader, "en", "en")
	newCtx = translator.SetLocale("lv")
	t.Equal("lv", translator.locale)
	t.Equal("lv", newCtx.Value(string(localeKey)))
}

func (t *TranslatorTestSuite) TestLoad() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "test", "en").Once().Return(map[string]map[string]string{
		"en": {
			"foo": "one",
			"bar": "two",
		},
	}, nil)

	// Case: Not loaded, successful load
	err := translator.load("test", "en")
	t.NoError(err)
	t.Equal("one", translator.loaded["test"]["en"]["foo"])

	// Case: Already loaded
	err = translator.load("test", "en")
	t.NoError(err)
	t.Equal("two", translator.loaded["test"]["en"]["bar"])

	// Case: Not loaded, loader returns an error
	t.mockLoader.On("Load", "folder3", "es").Once().Return(nil, ErrFileNotExist)
	err = translator.load("folder3", "es")
	t.EqualError(err, "translation file does not exist")
	t.Nil(translator.loaded["folder3"])
}

func (t *TranslatorTestSuite) TestIsLoaded() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en")
	t.mockLoader.On("Load", "test", "en").Once().Return(map[string]map[string]string{
		"en": {
			"foo": "one",
		},
	}, nil)
	err := translator.load("test", "en")
	t.NoError(err)

	// Case: Folder and locale are not loaded
	t.False(translator.isLoaded("folder1", "fr"))

	// Case: Folder is loaded, but locale is not loaded
	t.False(translator.isLoaded("test", "fr"))

	// Case: Both folder and locale are loaded
	t.True(translator.isLoaded("test", "en"))
}

func TestMakeReplacements(t *testing.T) {
	tests := []struct {
		line     string
		replace  map[string]string
		expected string
	}{
		{
			line: "Hello, :name! Welcome to :location.",
			replace: map[string]string{
				"name":     "krishan",
				"location": "india",
			},
			expected: "Hello, krishan! Welcome to india.",
		},
		{
			line:     "Testing with no replacements.",
			replace:  map[string]string{},
			expected: "Testing with no replacements.",
		},
		{
			line: "Replace :mohan with :SOHAM.",
			replace: map[string]string{
				"mohan": "lower",
				"SOHAM": "UPPER",
			},
			expected: "Replace lower with UPPER.",
		},
	}

	for _, test := range tests {
		result := makeReplacements(test.line, test.replace)
		assert.Equal(t, test.expected, result)
	}
}

func TestParseKey(t *testing.T) {
	tests := []struct {
		key     string
		folder  string
		keyPart string
	}{
		{key: "foo", folder: "*", keyPart: "foo"},
		{key: "foo.bar", folder: "foo", keyPart: "bar"},
		{key: "foo.bar.baz", folder: "foo.bar", keyPart: "baz"},
	}

	for _, test := range tests {
		folder, keyPart := parseKey(test.key)
		assert.Equal(t, test.folder, folder)
		assert.Equal(t, test.keyPart, keyPart)
	}
}
