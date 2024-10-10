package translation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	translationcontract "github.com/goravel/framework/contracts/translation"
	"github.com/goravel/framework/errors"
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
	loaded = make(map[string]map[string]map[string]any)
}

func (t *TranslatorTestSuite) TestChoice() {
	// load from `{locale}.json`
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{
		"test": map[string]any{
			"foo": "{0} first|{1}second",
		},
	}, nil)
	translation := translator.Choice("test.foo", 1)
	t.Equal("second", translation)

	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]any{
		"bar": "{0} first|{1}second",
	}, nil)
	translation = translator.Choice("test.bar", 1)
	t.Equal("second", translation)

	// test atomic replacements
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "fr", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "fr", "test").Once().Return(map[string]any{
		"baz": "{0} first|{1}Hello, :foo!",
	}, nil)
	translation = translator.Choice("test.baz", 1, translationcontract.Option{
		Replace: map[string]string{
			"foo": "baz:bar",
			"bar": "abcdef",
		},
		Locale: "fr",
	})
	t.Equal("Hello, baz:bar!", translation)

	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "auth").Once().Return(nil, errors.New("some error"))
	t.mockLog.On("Panic", errors.New("some error")).Once()
	translation = translator.Choice("auth.foo", 1)
	t.Equal("auth.foo", translation)

	// test nested folder and keys
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "foo/test").Once().Return(map[string]any{
		"bar": "{0} first|{1}second",
		"baz": map[string]any{
			"qux": "{0} first|{1}third",
		},
	}, nil)
	translation = translator.Choice("foo/test.baz.qux", 1)
	t.Equal("third", translation)
}

func (t *TranslatorTestSuite) TestGet() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]any{
		"bar": map[string]any{
			"baz": "two",
		},
	}, nil)
	translation := translator.Get("test.bar.baz")
	t.Equal("two", translation)

	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "en", "auth").Once().Return(map[string]any{
		"foo": "one",
	}, nil)
	translation = translator.Get("auth.foo")
	t.Equal("one", translation)

	// Case: when file exists but there is some error
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "en", "foo").Once().Return(nil, errors.New("some error"))
	t.mockLog.On("Panic", errors.New("some error")).Once()
	translation = translator.Get("foo.baz")
	t.Equal("foo.baz", translation)

	// Get json replacement
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "en", "greetings").Once().Return(map[string]any{
		"welcome_message": "Hello, :name! Welcome to :location.",
	}, nil)
	translation = translator.Get("greetings.welcome_message", translationcontract.Option{
		Replace: map[string]string{
			"location": "india",
			"name":     "krishan",
		},
	})
	t.Equal("Hello, krishan! Welcome to india.", translation)

	// test atomic replacements
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "en", "greet").Once().Return(map[string]any{
		"hi": "Hello, :who!",
	}, nil)
	translation = translator.Get("greet.hi", translationcontract.Option{
		Replace: map[string]string{
			"who": "baz:bar",
			"bar": "abcdef",
		},
	})
	t.Equal("Hello, baz:bar!", translation)

	// preserve order of replacements
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "en", "welcome").Once().Return(map[string]any{
		"message": ":greeting :name",
	}, nil)
	translation = translator.Get("welcome.message", translationcontract.Option{
		Replace: map[string]string{
			"name":     "krishan",
			"greeting": "Hello",
		},
	})
	t.Equal("Hello krishan", translation)

	// non-existing json key looks for regular keys
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "en", "foo/test").Once().Return(map[string]any{
		"bar": "one",
	}, nil)
	translation = translator.Get("foo/test.bar")
	t.Equal("one", translation)

	// empty fallback
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "en", "messages").Once().Return(map[string]any{}, errors.LangFileNotExist)
	translation = translator.Get("messages.foo")
	t.Equal("messages.foo", translation)

	// Case: Fallback to a different locale
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "fr", t.mockLog)
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "en", "test3").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "fr", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "fr", "test3").Once().Return(map[string]any{
		"nonexistentKey": "French translation",
	}, nil)
	translation = translator.Get("test3.nonexistentKey", translationcontract.Option{
		Fallback: translationcontract.Bool(true),
		Locale:   "en",
	})
	t.Equal("French translation", translation)

	// Case: Fallback to a different locale with fallback disabled
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "fr", t.mockLog)
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "en", "test4").Once().Return(map[string]any{}, errors.LangFileNotExist)
	translation = translator.Get("test4.nonexistentKey", translationcontract.Option{
		Fallback: translationcontract.Bool(false),
		Locale:   "en",
	})
	t.Equal("test4.nonexistentKey", translation)

	// load from `{locale}.json` file
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{
		"foo":            "bar",
		"nonexistentKey": "English translation",
	}, nil)
	translation = translator.Get("foo")
	t.Equal("bar", translation)

	// Case: use JSON file as fallback
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "fr", t.mockLog)
	t.mockLoader.On("Load", "en", "fallback").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "fr", "*").Once().Return(map[string]any{
		"fallback": map[string]any{
			"nonexistentKey": "French translation",
		},
	}, nil)
	translation = translator.Get("fallback.nonexistentKey", translationcontract.Option{
		Fallback: translationcontract.Bool(true),
		Locale:   "en",
	})
	t.Equal("French translation", translation)

	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "foo").Once().Return(map[string]any{}, nil)
	translation = translator.Get("foo.bar")
	t.Equal("foo.bar", translation)

	// Case: Nested folder and keys
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "foo/messages").Once().Return(map[string]any{
		"bar": "one",
		"baz": map[string]any{
			"qux": "two",
		},
	}, nil)
	translation = translator.Get("foo/messages.baz.qux")
	t.Equal("two", translation)
}

func (t *TranslatorTestSuite) TestGetLocale() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)

	// Case: Get locale initially set
	locale := translator.CurrentLocale()
	t.Equal("en", locale)

	// Case: Set locale using SetLocale and then get it
	ctx := translator.SetLocale("fr")

	translator = NewTranslator(ctx, t.mockLoader, "en", "en", t.mockLog)
	locale = translator.CurrentLocale()
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
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "en", "example").Once().Return(map[string]any{
		"hello": "world",
	}, nil)
	hasKey := translator.Has("example.hello")
	t.True(hasKey)

	// Case: Key does not exist in translations
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "en", "user").Once().Return(map[string]any{
		"name": "Bowen",
	}, nil)
	hasKey = translator.Has("user.email")
	t.False(hasKey)

	// Case: Nested folder and keys
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	t.mockLoader.On("Load", "en", "foo/test").Once().Return(map[string]any{
		"bar": "one",
		"baz": map[string]any{
			"qux": "two",
		},
	}, nil)
	t.True(translator.Has("foo/test.baz.qux"))

	// Case: Key exists in {locale}.json
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "fr", "*").Once().Return(map[string]any{
		"hello": "world",
	}, nil)
	hasKey = translator.Has("hello", translationcontract.Option{
		Locale: "fr",
	})
	t.True(hasKey)
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
	t.mockLoader.On("Load", "en", "test").Once().Return(map[string]any{
		"foo": "one",
		"bar": "two",
	}, nil)

	// Case: Not loaded, successful load
	err := translator.load("en", "test")
	t.NoError(err)
	t.Equal("one", loaded["en"]["test"]["foo"])

	// Case: Already loaded
	err = translator.load("en", "test")
	t.NoError(err)
	t.Equal("two", loaded["en"]["test"]["bar"])

	// Case: Not loaded, loader returns an error
	t.mockLoader.On("Load", "es", "folder3").Once().Return(nil, errors.LangFileNotExist)
	err = translator.load("es", "folder3")
	t.EqualError(err, "translation file does not exist")
	t.Nil(loaded["folder3"])

	// Case: Nested folder and keys
	translator = NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "foo/test").Once().Return(map[string]any{
		"bar": "one",
	}, nil)
	err = translator.load("en", "foo/test")
	t.NoError(err)
	t.Equal("one", loaded["en"]["foo/test"]["bar"])
}

func (t *TranslatorTestSuite) TestIsLoaded() {
	translator := NewTranslator(t.ctx, t.mockLoader, "en", "en", t.mockLog)
	t.mockLoader.On("Load", "en", "bar").Once().Return(map[string]any{
		"foo": "one",
	}, nil)
	err := translator.load("en", "bar")
	t.NoError(err)

	// Case: Folder and locale are not loaded
	t.False(translator.isLoaded("fr", "folder1"))

	// Case: Folder is loaded, but locale is not loaded
	t.False(translator.isLoaded("fr", "bar"))

	// Case: Both folder and locale are loaded
	t.True(translator.isLoaded("en", "bar"))
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
		key   string
		group string
		item  string
	}{
		{key: "foo", group: "foo", item: ""},
		{key: "foo.bar", group: "foo", item: "bar"},
		{key: "foo.bar.baz", group: "foo", item: "bar.baz"},
		{key: "foo/bar.baz", group: "foo/bar", item: "baz"},
	}

	for _, test := range tests {
		group, item := parseKey(test.key)
		assert.Equal(t, test.group, group)
		assert.Equal(t, test.item, item)
	}
}

func TestGetValue(t *testing.T) {
	obj := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "42",
			},
		},
	}

	result := getValue(obj, "a.b.c")
	assert.Equal(t, "42", result)

	result = getValue(obj, "x.y.z")
	assert.Equal(t, nil, result)
}

func Benchmark_Choice(b *testing.B) {
	s := new(TranslatorTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()
	b.StartTimer()
	b.ResetTimer()

	translator := NewTranslator(s.ctx, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Return(map[string]any{
		"test": map[string]any{
			"foo": "{0} first|{1}second",
		},
	}, nil)

	for i := 0; i < b.N; i++ {
		translator.Choice("test.foo", 1)
	}

	b.StopTimer()
}

func Benchmark_Get(b *testing.B) {
	s := new(TranslatorTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()
	b.StartTimer()
	b.ResetTimer()

	translator := NewTranslator(s.ctx, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Return(map[string]any{
		"test": map[string]any{
			"foo": "bar",
		},
	}, nil)

	for i := 0; i < b.N; i++ {
		translator.Get("test.foo")
	}

	b.StopTimer()
}

func Benchmark_Has(b *testing.B) {
	s := new(TranslatorTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()
	b.StartTimer()
	b.ResetTimer()

	translator := NewTranslator(s.ctx, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Return(map[string]any{
		"test": map[string]any{
			"foo": "bar",
		},
	}, nil)

	for i := 0; i < b.N; i++ {
		translator.Has("test.foo")
	}

	b.StopTimer()
}
