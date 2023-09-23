package translation

import (
	"strings"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	translationcontract "github.com/goravel/framework/contracts/translation"
)

type Translator struct {
	loader   translationcontract.Loader
	locale   string
	fallback string
	loaded   map[string]map[string]any
}

func NewTranslator(loader translationcontract.Loader, locale string, fallback string) *Translator {
	return &Translator{
		loader:   loader,
		locale:   locale,
		fallback: fallback,
		loaded:   make(map[string]map[string]any),
	}
}

func (t *Translator) Get(key string, options ...translationcontract.Option) (string, error) { // Check if a custom locale is provided in options.
	var (
		locale = t.locale
	)
	if len(options) > 0 && options[0].Locale != "" {
		locale = options[0].Locale
	}

	folder, keyPart := parseKey(key)

	// Load translations for the specified folder and locale.
	if err := t.load(folder, locale); err != nil {
		return "", err
	}

	dataBytes, err := sonic.Marshal(t.loaded[folder][locale])
	if err != nil {
		return "", err
	}

	// Use Sonic to get the translation for the keyPart.
	root, err := sonic.Get(dataBytes, keyPart)
	if err != nil {
		// Handle errors when key not found or other Sonic-related errors.
		if err != ast.ErrNotExist {
			return "", err
		}
		// TODO: try to get using fallback if doesn't find any for now just return the key
		//if t.fallback != "" {
		//	return t.Get(key, translationcontract.Option{Locale: t.fallback})
		//}
		// If key not found, return the key itself for debugging.
		return key, nil
	}

	line, err := root.Raw()
	if err != nil {
		return "", err
	}

	return makeReplacements(line, options...), nil
}

func (t *Translator) Has(key string, options ...translationcontract.Option) bool {
	line, err := t.Get(key, options...)
	return err == nil && line != key
}

func (t *Translator) GetLocale() string {
	return t.locale
}

func (t *Translator) SetLocale(locale string) {
	t.locale = locale
}

func (t *Translator) GetFallback() string {
	return t.fallback
}

func (t *Translator) SetFallback(locale string) {
	t.fallback = locale
}

func (t *Translator) load(folder string, locale string) error {
	if t.isLoaded(folder, locale) {
		return nil
	}

	translations, err := t.loader.Load(folder, locale)
	if err != nil {
		return err
	}
	t.loaded[folder] = translations
	return nil
}

func (t *Translator) isLoaded(folder string, locale string) bool {
	if _, ok := t.loaded[folder]; !ok {
		return false
	}

	if _, ok := t.loaded[folder][locale]; !ok {
		return false
	}

	return true
}

func makeReplacements(line string, options ...translationcontract.Option) string {
	if len(options) == 0 {
		return line
	}

	replace := options[0].Replace

	if len(replace) == 0 {
		return line
	}

	var shouldReplace []string
	casesTitle := cases.Title(language.Und)
	for k, v := range replace {
		shouldReplace = append(shouldReplace, ":"+k, v)
		shouldReplace = append(shouldReplace, ":"+casesTitle.String(k), casesTitle.String(v))
		shouldReplace = append(shouldReplace, ":"+strings.ToUpper(k), strings.ToUpper(v))
	}

	return strings.NewReplacer(shouldReplace...).Replace(line)
}

func parseKey(key string) (folder, keyPart string) {
	parts := strings.Split(key, ".")
	folder = "*"
	keyPart = key
	if len(parts) > 1 {
		folder = strings.Join(parts[:len(parts)-1], ".")
		keyPart = parts[len(parts)-1]
	}
	return folder, keyPart
}
