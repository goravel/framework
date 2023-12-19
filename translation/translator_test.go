package translation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	translationcontract "github.com/goravel/framework/contracts/translation"
	"github.com/goravel/framework/http"
	mocklog "github.com/goravel/framework/mocks/log"
	mockloader "github.com/goravel/framework/mocks/translation"
)

type TranslatorTestSuite struct {
	suite.Suite
	mockLoader *mockloader.Loader
	ctx        context.Context
	mockLog    *mocklog.Log
}

func TestTranslatorTestSuite(t *testing.T) {
	suite.Run(t, &TranslatorTestSuite{})
}

func (t *TranslatorTestSuite) SetupTest() {
	t.mockLoader = mockloader.NewLoader(t.T())
	t.ctx = context.Background()
	t.mockLog = mocklog.NewLog(t.T())
}

func (t *TranslatorTestSuite) TestChoice() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]map[string]any{
		"test": {
			"foo": "{0} first|{1}second",
		},
	}, nil)
	translation := translator.Choice("test.foo", 1)
	t.Equal("second", translation)

	// test atomic replacements
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "fr", "test").Once().Return(map[string]map[string]any{
		"test": {
			"foo": "{0} first|{1}Hello, :foo!",
		},
	}, nil)
	translation = translator.Choice("test.foo", 1, translationcontract.Option{
		Replace: map[string]string{
			"foo": "baz:bar",
			"bar": "abcdef",
		},
		Locale: "fr",
	})
	t.Equal("Hello, baz:bar!", translation)

	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(nil, errors.New("some error"))
	translation = translator.Choice("test.foo", 1)
	t.Equal("test.foo", translation)

	// test nested folder and keys
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "foo/test").Once().Return(map[string]map[string]any{
		"foo/test": {
			"bar": "{0} first|{1}second",
			"baz": map[string]string{
				"qux": "{0} first|{1}third",
			},
		},
	}, nil)
	translation = translator.Choice("foo/test.baz.qux", 1)
	t.Equal("third", translation)
}

func (t *TranslatorTestSuite) TestGet() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]map[string]any{
		"test": {
			"foo": "one",
			"bar": map[string]string{
				"baz": "two",
			},
		},
	}, nil)
	translation := translator.Get("test.bar.baz")
	t.Equal("two", translation)

	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]map[string]any{
		"test": {
			"foo": "one",
		},
	}, nil)
	translation = translator.Get("test.foo")
	t.Equal("one", translation)

	// Case: when file exists but there is some error
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(nil, errors.New("some error"))
	translation = translator.Get("test.foo")
	t.Equal("test.foo", translation)

	// Get json replacement
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]map[string]any{
		"test": {
			"foo": "Hello, :name! Welcome to :location.",
		},
	}, nil)
	translation = translator.Get("test.foo", translationcontract.Option{
		Replace: map[string]string{
			"name":     "krishan",
			"location": "india",
		},
	})
	t.Equal("Hello, krishan! Welcome to india.", translation)

	// test atomic replacements
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]map[string]any{
		"test": {
			"foo": "Hello, :foo!",
		},
	}, nil)
	translation = translator.Get("test.foo", translationcontract.Option{
		Replace: map[string]string{
			"foo": "baz:bar",
			"bar": "abcdef",
		},
	})
	t.Equal("Hello, baz:bar!", translation)

	// preserve order of replacements
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]map[string]any{
		"test": {
			"foo": ":greeting :name",
		},
	}, nil)
	translation = translator.Get("test.foo", translationcontract.Option{
		Replace: map[string]string{
			"name":     "krishan",
			"greeting": "Hello",
		},
	})
	t.Equal("Hello krishan", translation)

	// non-existing json key looks for regular keys
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "foo/test").Once().Return(map[string]map[string]any{
		"foo/test": {
			"bar": "one",
		},
	}, nil)
	translation = translator.Get("foo/test.bar")
	t.Equal("one", translation)

	// empty fallback
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]map[string]any{
		"test": {},
	}, nil)
	translation = translator.Get("test.foo")
	t.Equal("test.foo", translation)

	// Case: Fallback to a different locale
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "fr", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]map[string]any{
		"test": {},
	}, nil)
	t.mockLoader.On("Load", "fr", "test").Once().Return(map[string]map[string]any{
		"test": {
			"nonexistentKey": "French translation",
		},
	}, nil)
	translation = translator.Get("test.nonexistentKey", translationcontract.Option{
		Fallback: translationcontract.Bool(true),
		Locale:   "en",
	})
	t.Equal("French translation", translation)

	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]map[string]any{}, nil)
	translation = translator.Get("test.foo")
	t.Equal("test.foo", translation)

	// Case: Nested folder and keys
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "foo/test").Once().Return(map[string]map[string]any{
		"foo/test": {
			"bar": "one",
			"baz": map[string]string{
				"qux": "two",
			},
		},
	}, nil)
	translation = translator.Get("foo/test.baz.qux")
	t.Equal("two", translation)
}

func (t *TranslatorTestSuite) TestGetLocale() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)

	// Case: Get locale initially set
	locale := translator.GetLocale()
	t.Equal("en", locale)

	// Case: Set locale using SetLocale and then get it
	translator.SetLocale("fr")
	locale = translator.GetLocale()
	t.Equal("fr", locale)
}

func (t *TranslatorTestSuite) TestGetFallback() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)

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
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]map[string]any{
		"test": {
			"hello": "world",
		},
	}, nil)
	hasKey := translator.Has("test.hello")
	t.True(hasKey)

	// Case: Key does not exist in translations
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]map[string]any{
		"test": {
			"name": "Bowen",
		},
	}, nil)
	hasKey = translator.Has("test.email")
	t.False(hasKey)

	// Case: Nested folder and keys
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "foo/test").Once().Return(map[string]map[string]any{
		"foo/test": {
			"bar": "one",
			"baz": map[string]string{
				"qux": "two",
			},
		},
	}, nil)
	t.True(translator.Has("foo/test.baz.qux"))
}

func (t *TranslatorTestSuite) TestSetFallback() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)

	// Case: Set fallback using SetFallback
	newCtx := translator.SetFallback("fr")
	t.Equal("fr", translator.fallback)
	t.Equal("fr", newCtx.Value(string(fallbackLocaleKey)))
}

func (t *TranslatorTestSuite) TestSetLocale() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)

	// Case: Set locale using SetLocale
	newCtx := translator.SetLocale("fr")
	t.Equal("fr", translator.locale)
	t.Equal("fr", newCtx.Value(string(localeKey)))

	// Case: use http.Context
	translator = NewTranslator(http.Background(), t.mockLoader, "en", "en", t.mockLog)
	newCtx = translator.SetLocale("lv")
	t.Equal("lv", translator.locale)
	t.Equal("lv", newCtx.Value(string(localeKey)))
}

func (t *TranslatorTestSuite) TestLoad() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]map[string]any{
		"test": {
			"foo": "one",
			"bar": "two",
		},
	}, nil)

	// Case: Not loaded, successful load
	err := translator.load("en", "test")
	t.NoError(err)
	t.Equal("one", translator.loaded["en"]["test"]["foo"])

	// Case: Already loaded
	err = translator.load("en", "test")
	t.NoError(err)
	t.Equal("two", translator.loaded["en"]["test"]["bar"])

	// Case: Not loaded, loader returns an error
	t.mockLoader.On("Load", "es", "folder3").Once().Return(nil, ErrFileNotExist)
	err = translator.load("es", "folder3")
	t.EqualError(err, "translation file does not exist")
	t.Nil(translator.loaded["folder3"])

	// Case: Nested folder and keys
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "foo/test").Once().Return(map[string]map[string]any{
		"foo/test": {
			"bar": "one",
		},
	}, nil)
	err = translator.load("en", "foo/test")
	t.NoError(err)
	t.Equal("one", translator.loaded["en"]["foo/test"]["bar"])
}

func (t *TranslatorTestSuite) TestIsLoaded() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]map[string]any{
		"test": {
			"foo": "one",
		},
	}, nil)
	err := translator.load("en", "test")
	t.NoError(err)

	// Case: Folder and locale are not loaded
	t.False(translator.isLoaded("fr", "folder1"))

	// Case: Folder is loaded, but locale is not loaded
	t.False(translator.isLoaded("fr", "test"))

	// Case: Both folder and locale are loaded
	t.True(translator.isLoaded("en", "test"))
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
		{key: "foo", folder: "foo", keyPart: ""},
		{key: "foo.bar", folder: "foo", keyPart: "bar"},
		{key: "foo.bar.baz", folder: "foo", keyPart: "bar.baz"},
		{key: "foo/bar.baz", folder: "foo/bar", keyPart: "baz"},
	}

	for _, test := range tests {
		folder, keyPart := parseKey(test.key)
		assert.Equal(t, test.folder, folder)
		assert.Equal(t, test.keyPart, keyPart)
	}
}
