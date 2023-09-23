package translation

import (
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/goravel/framework/contracts/http"
	translationcontract "github.com/goravel/framework/contracts/translation"
)

type Translator struct {
	loader   translationcontract.Loader
	locale   string
	fallback string
	loaded   map[string]map[string]any
	selector *MessageSelector
}

func NewTranslator(loader translationcontract.Loader, locale string, fallback string) *Translator {
	return &Translator{
		loader:   loader,
		locale:   locale,
		fallback: fallback,
		loaded:   make(map[string]map[string]any),
	}
}

func (t *Translator) Get(ctx http.Context, key string, options ...translationcontract.Option) (string, error) { // Check if a custom locale is provided in options.
	locale := t.GetLocale(ctx)
	if len(options) > 0 && options[0].Locale != "" {
		locale = options[0].Locale
	}

	// Check if a custom fallback locale is provided in options.
	fallback := true
	if len(options) > 0 && options[0].Fallback != nil {
		fallback = *options[0].Fallback
	}

	// Parse the key into folder and key parts.
	folder, keyPart := parseKey(key)

	// Load the translations for the given locale.
	if err := t.load(folder, locale); err != nil {
		return "", err
	}

	// Check if the key exists in the loaded translations.
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

		if fallback {
			return t.Get(ctx, key, translationcontract.Option{
				Locale:   t.fallback,
				Fallback: options[0].Fallback,
				Replace:  options[0].Replace,
			})
		}

		// If key not found, return the key itself for debugging.
		return key, nil
	}

	line, err := root.Raw()
	if err != nil {
		return "", err
	}
	if len(options) > 0 {
		return makeReplacements(line, options[0].Replace), nil
	}

	return line, nil
}

func (t *Translator) Choice(ctx http.Context, key string, number int, options ...translationcontract.Option) (string, error) {
	line, err := t.Get(ctx, key, options...)
	if err != nil {
		return "", err
	}
	replace := map[string]string{
		"count": strconv.Itoa(number),
	}
	locale := t.GetLocale(ctx)
	if len(options) > 0 && options[0].Locale != "" {
		locale = options[0].Locale
	}
	return makeReplacements(t.getSelector().Choose(line, number, locale), replace), nil
}

func (t *Translator) Has(ctx http.Context, key string, options ...translationcontract.Option) bool {
	line, err := t.Get(ctx, key, options...)
	return err == nil && line != key
}

func (t *Translator) GetLocale(ctx http.Context) string {
	if locale, ok := ctx.Value("locale").(string); ok {
		return locale
	}
	return t.locale
}

func (t *Translator) SetLocale(ctx http.Context, locale string) error {
	t.locale = locale
	ctx.WithValue("locale", locale)
	return nil
}

func (t *Translator) GetFallback(ctx http.Context) string {
	if fallback, ok := ctx.Value("fallback_locale").(string); ok {
		return fallback
	}
	return t.fallback
}

func (t *Translator) SetFallback(ctx http.Context, locale string) error {
	t.fallback = locale
	ctx.WithValue("fallback_locale", locale)
	return nil
}

func (t *Translator) getSelector() *MessageSelector {
	if t.selector == nil {
		t.selector = NewMessageSelector()
	}
	return t.selector
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

func (t *Translator) localeArray(locale string) []string {
	if locale != "" {
		return []string{locale, t.fallback}
	}
	return []string{t.locale, t.fallback}
}

func makeReplacements(line string, replace map[string]string) string {
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
