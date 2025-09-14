package translation

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractstranslation "github.com/goravel/framework/contracts/translation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
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

func (s *TranslatorTestSuite) SetupTest() {
	s.mockLoader = mockloader.NewLoader(s.T())
	s.ctx = context.Background()
	s.mockLog = mocklog.NewLog(s.T())
	loaded = make(map[string]map[string]map[string]any)
}

func (s *TranslatorTestSuite) TestChoice() {
	// load from `{locale}.json`
	translator := NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{
		"test": map[string]any{
			"foo": "{0} first|{1}second",
		},
	}, nil)
	translation := translator.Choice("test.foo", 1)
	s.Equal("second", translation)

	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "test").Once().Return(map[string]any{
		"bar": "{0} first|{1}second",
	}, nil)
	translation = translator.Choice("test.bar", 1)
	s.Equal("second", translation)

	// test atomic replacements
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "fr", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "fr", "test").Once().Return(map[string]any{
		"baz": "{0} first|{1}Hello, :foo!",
	}, nil)
	translation = translator.Choice("test.baz", 1, contractstranslation.Option{
		Replace: map[string]string{
			"foo": "baz:bar",
			"bar": "abcdef",
		},
		Locale: "fr",
	})
	s.Equal("Hello, baz:bar!", translation)

	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "auth").Once().Return(nil, errors.New("some error"))
	s.mockLog.On("Panic", errors.New("some error")).Once()
	translation = translator.Choice("auth.foo", 1)
	s.Equal("auth.foo", translation)

	// test nested folder and keys
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "foo/test").Once().Return(map[string]any{
		"bar": "{0} first|{1}second",
		"baz": map[string]any{
			"qux": "{0} first|{1}third",
		},
	}, nil)
	translation = translator.Choice("foo/test.baz.qux", 1)
	s.Equal("third", translation)
}

func (s *TranslatorTestSuite) TestGet() {
	translator := NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "en", "test").Once().Return(map[string]any{
		"bar": map[string]any{
			"baz": "two",
		},
	}, nil)
	translation := translator.Get("test.bar.baz")
	s.Equal("two", translation)

	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "en", "auth").Once().Return(map[string]any{
		"foo": "one",
	}, nil)
	translation = translator.Get("auth.foo")
	s.Equal("one", translation)

	// Case: when file exists but there is some error
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "en", "foo").Once().Return(nil, errors.New("some error"))
	s.mockLog.On("Panic", errors.New("some error")).Once()
	translation = translator.Get("foo.baz")
	s.Equal("foo.baz", translation)

	// Get json replacement
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "en", "greetings").Once().Return(map[string]any{
		"welcome_message": "Hello, :name! Welcome to :location.",
	}, nil)
	translation = translator.Get("greetings.welcome_message", contractstranslation.Option{
		Replace: map[string]string{
			"location": "india",
			"name":     "krishan",
		},
	})
	s.Equal("Hello, krishan! Welcome to india.", translation)

	// test atomic replacements
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "en", "greet").Once().Return(map[string]any{
		"hi": "Hello, :who!",
	}, nil)
	translation = translator.Get("greet.hi", contractstranslation.Option{
		Replace: map[string]string{
			"who": "baz:bar",
			"bar": "abcdef",
		},
	})
	s.Equal("Hello, baz:bar!", translation)

	// preserve order of replacements
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "en", "welcome").Once().Return(map[string]any{
		"message": ":greeting :name",
	}, nil)
	translation = translator.Get("welcome.message", contractstranslation.Option{
		Replace: map[string]string{
			"name":     "krishan",
			"greeting": "Hello",
		},
	})
	s.Equal("Hello krishan", translation)

	// non-existing json key looks for regular keys
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "en", "foo/test").Once().Return(map[string]any{
		"bar": "one",
	}, nil)
	translation = translator.Get("foo/test.bar")
	s.Equal("one", translation)

	// empty fallback
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "en", "messages").Once().Return(map[string]any{}, errors.LangFileNotExist)
	translation = translator.Get("messages.foo")
	s.Equal("messages.foo", translation)

	// Case: Fallback to a different locale
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "fr", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "en", "test3").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "fr", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "fr", "test3").Once().Return(map[string]any{
		"nonexistentKey": "French translation",
	}, nil)
	translation = translator.Get("test3.nonexistentKey", contractstranslation.Option{
		Fallback: contractstranslation.Bool(true),
		Locale:   "en",
	})
	s.Equal("French translation", translation)

	// Case: Fallback to a different locale with fallback disabled
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "fr", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "en", "test4").Once().Return(map[string]any{}, errors.LangFileNotExist)
	translation = translator.Get("test4.nonexistentKey", contractstranslation.Option{
		Fallback: contractstranslation.Bool(false),
		Locale:   "en",
	})
	s.Equal("test4.nonexistentKey", translation)

	// load from `{locale}.json` file
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{
		"foo":            "bar",
		"nonexistentKey": "English translation",
	}, nil)
	translation = translator.Get("foo")
	s.Equal("bar", translation)

	// Case: use JSON file as fallback
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "fr", s.mockLog)
	s.mockLoader.On("Load", "en", "fallback").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "fr", "*").Once().Return(map[string]any{
		"fallback": map[string]any{
			"nonexistentKey": "French translation",
		},
	}, nil)
	translation = translator.Get("fallback.nonexistentKey", contractstranslation.Option{
		Fallback: contractstranslation.Bool(true),
		Locale:   "en",
	})
	s.Equal("French translation", translation)

	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "foo").Once().Return(map[string]any{}, nil)
	translation = translator.Get("foo.bar")
	s.Equal("foo.bar", translation)

	// Case: Nested folder and keys
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "foo/messages").Once().Return(map[string]any{
		"bar": "one",
		"baz": map[string]any{
			"qux": "two",
		},
	}, nil)
	translation = translator.Get("foo/messages.baz.qux")
	s.Equal("two", translation)

	// Case: No loaders available
	translator = NewTranslator(s.ctx, nil, nil, "en", "en", s.mockLog)
	translation = translator.Get("test.key")
	s.Equal("test.key", translation)
}

func (s *TranslatorTestSuite) TestGetLocale() {
	translator := NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)

	// Case: Get locale initially set
	locale := translator.CurrentLocale()
	s.Equal("en", locale)

	// Case: Set locale using SetLocale and then get it
	ctx := translator.SetLocale("fr")

	translator = NewTranslator(ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	locale = translator.CurrentLocale()
	s.Equal("fr", locale)
}

func (s *TranslatorTestSuite) TestGetFallback() {
	translator := NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)

	// Case: No explicit fallback set
	fallback := translator.GetFallback()
	s.Equal("en", fallback)

	// Case: Set fallback using SetFallback
	newCtx := translator.SetFallback("fr")
	fallback = translator.GetFallback()
	s.Equal("fr", fallback)
	s.Equal("fr", newCtx.Value(string(fallbackLocaleKey)))
}

func (s *TranslatorTestSuite) TestHas() {
	// Case: Key exists in translations
	translator := NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "en", "example").Once().Return(map[string]any{
		"hello": "world",
	}, nil)
	hasKey := translator.Has("example.hello")
	s.True(hasKey)

	// Case: Key does not exist in translations
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "en", "user").Once().Return(map[string]any{
		"name": "Bowen",
	}, nil)
	hasKey = translator.Has("user.email")
	s.False(hasKey)

	// Case: Nested folder and keys
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "*").Once().Return(map[string]any{}, errors.LangFileNotExist)
	s.mockLoader.On("Load", "en", "foo/test").Once().Return(map[string]any{
		"bar": "one",
		"baz": map[string]any{
			"qux": "two",
		},
	}, nil)
	s.True(translator.Has("foo/test.baz.qux"))

	// Case: Key exists in {locale}.json
	translator = NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "fr", "*").Once().Return(map[string]any{
		"hello": "world",
	}, nil)
	hasKey = translator.Has("hello", contractstranslation.Option{
		Locale: "fr",
	})
	s.True(hasKey)
}

func (s *TranslatorTestSuite) TestSetFallback() {
	translator := NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)

	// Case: Set fallback using SetFallback
	newCtx := translator.SetFallback("fr")
	s.Equal("fr", translator.fallback)
	s.Equal("fr", newCtx.Value(string(fallbackLocaleKey)))
}

func (s *TranslatorTestSuite) TestSetLocale() {
	translator := NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)

	// Case: Set locale using SetLocale
	newCtx := translator.SetLocale("fr")
	s.Equal("fr", translator.locale)
	s.Equal("fr", newCtx.Value(string(localeKey)))

	// Case: use http.Context
	translator = NewTranslator(http.Background(), nil, s.mockLoader, "en", "en", s.mockLog)
	newCtx = translator.SetLocale("lv")
	s.Equal("lv", translator.locale)
	s.Equal("lv", newCtx.Value(string(localeKey)))
}

func (s *TranslatorTestSuite) TestLoad() {
	fsLoader := NewFSLoader(fstest.MapFS{
		"en/test.json": &fstest.MapFile{Data: []byte(`{"foo": "bar", "baz": {"foo": "bar"}}`)},
	}, json.New())

	tests := []struct {
		name        string
		fsLoader    contractstranslation.Loader
		fileLoader  contractstranslation.Loader
		setup       func()
		locale      string
		group       string
		expected    map[string]any
		expectError error
	}{
		{
			name:       "already loaded",
			fsLoader:   fsLoader,
			fileLoader: s.mockLoader,
			setup: func() {
				loaded = map[string]map[string]map[string]any{
					"en": {
						"test": {
							"foo": "bar",
						},
					},
				}
			},
			locale: "en",
			group:  "test",
			expected: map[string]any{
				"foo": "bar",
			},
			expectError: nil,
		},
		{
			name:       "successful load with file loader",
			fsLoader:   fsLoader,
			fileLoader: s.mockLoader,
			setup: func() {
				s.mockLoader.EXPECT().Load("en", "test").Return(map[string]any{
					"foo": "bar",
					"baz": map[string]any{
						"qux": "quux",
					},
				}, nil).Once()
			},
			locale: "en",
			group:  "test",
			expected: map[string]any{
				"foo": "bar",
				"baz": map[string]any{
					"qux": "quux",
				},
			},
			expectError: nil,
		},
		{
			name:       "file loader error, fallback to fs loader",
			fsLoader:   fsLoader,
			fileLoader: s.mockLoader,
			setup: func() {
				s.mockLoader.EXPECT().Load("en", "test").Return(nil, assert.AnError).Once()
			},
			locale: "en",
			group:  "test",
			expected: map[string]any{
				"foo": "bar",
				"baz": map[string]any{
					"foo": "bar",
				},
			},
			expectError: nil,
		},
		{
			name:       "both loaders fail",
			fsLoader:   fsLoader,
			fileLoader: s.mockLoader,
			setup: func() {
				s.mockLoader.EXPECT().Load("fr", "nonexistent").Return(nil, assert.AnError).Once()
			},
			locale:      "fr",
			group:       "nonexistent",
			expected:    nil,
			expectError: errors.LangFileNotExist,
		},
		{
			name:       "nested folder structure",
			fsLoader:   fsLoader,
			fileLoader: s.mockLoader,
			setup: func() {
				s.mockLoader.EXPECT().Load("en", "foo/test").Return(map[string]any{
					"bar": "baz",
					"nested": map[string]any{
						"deep": "value",
					},
				}, nil)
			},
			locale: "en",
			group:  "foo/test",
			expected: map[string]any{
				"bar": "baz",
				"nested": map[string]any{
					"deep": "value",
				},
			},
			expectError: nil,
		},
		{
			name:       "empty translations from file loader and fs loader",
			fsLoader:   fsLoader,
			fileLoader: s.mockLoader,
			setup: func() {
				s.mockLoader.EXPECT().Load("en", "empty").Return(map[string]any{}, nil)
			},
			locale:      "en",
			group:       "empty",
			expected:    nil,
			expectError: errors.LangFileNotExist,
		},
		{
			name:       "file loader returns empty, fs loader succeeds",
			fsLoader:   fsLoader,
			fileLoader: s.mockLoader,
			setup: func() {
				s.mockLoader.EXPECT().Load("en", "test").Return(map[string]any{}, nil)
			},
			locale: "en",
			group:  "test",
			expected: map[string]any{
				"foo": "bar",
				"baz": map[string]any{
					"foo": "bar",
				},
			},
			expectError: nil,
		},
		{
			name:        "no loaders available",
			setup:       func() {},
			locale:      "en",
			group:       "test",
			expected:    nil,
			expectError: errors.LangNoLoaderAvailable,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			loaded = make(map[string]map[string]map[string]any)

			tt.setup()

			translator := NewTranslator(s.ctx, tt.fsLoader, tt.fileLoader, "en", "en", s.mockLog)

			err := translator.load(tt.locale, tt.group)

			if tt.expectError != nil {
				s.Equal(tt.expectError.Error(), err.Error())
			} else {
				s.NoError(err)
				if tt.expected != nil {
					s.Equal(tt.expected, loaded[tt.locale][tt.group])
				}
			}
		})
	}
}

func (s *TranslatorTestSuite) TestIsLoaded() {
	translator := NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
	s.mockLoader.On("Load", "en", "bar").Once().Return(map[string]any{
		"foo": "one",
	}, nil)
	err := translator.load("en", "bar")
	s.NoError(err)

	// Case: Folder and locale are not loaded
	s.False(translator.isLoaded("fr", "folder1"))

	// Case: Folder is loaded, but locale is not loaded
	s.False(translator.isLoaded("fr", "bar"))

	// Case: Both folder and locale are loaded
	s.True(translator.isLoaded("en", "bar"))
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

	translator := NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
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

	translator := NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
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

	translator := NewTranslator(s.ctx, nil, s.mockLoader, "en", "en", s.mockLog)
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
