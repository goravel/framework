package translation

import (
	"context"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/goravel/framework/contracts/http"
	translationcontract "github.com/goravel/framework/contracts/translation"
)

type Translator struct {
	ctx      context.Context
	loader   translationcontract.Loader
	locale   string
	fallback string
	loaded   map[string]map[string]map[string]string
	selector *MessageSelector
	key      string
}

// contextKey is an unexported type for keys defined in this package.
type contextKey string

const fallbackLocaleKey = contextKey("fallback_locale")
const localeKey = contextKey("locale")

func NewTranslator(ctx context.Context, loader translationcontract.Loader, locale string, fallback string) *Translator {
	return &Translator{
		ctx:      ctx,
		loader:   loader,
		locale:   locale,
		fallback: fallback,
		loaded:   make(map[string]map[string]map[string]string),
		selector: NewMessageSelector(),
	}
}

func (t *Translator) Choice(key string, number int, options ...translationcontract.Option) (string, error) {
	line, err := t.Get(key, options...)
	if err != nil {
		return "", err
	}

	replace := map[string]string{
		"count": strconv.Itoa(number),
	}

	locale := t.GetLocale()
	if len(options) > 0 && options[0].Locale != "" {
		locale = options[0].Locale
	}

	return makeReplacements(t.selector.Choose(line, number, locale), replace), nil
}

func (t *Translator) Get(key string, options ...translationcontract.Option) (string, error) {
	if t.key == "" {
		t.key = key
	}

	locale := t.GetLocale()
	// Check if a custom locale is provided in options.
	if len(options) > 0 && options[0].Locale != "" {
		locale = options[0].Locale
	}

	fallback := true
	// Check if a custom fallback is provided in options.
	if len(options) > 0 && options[0].Fallback != nil {
		fallback = *options[0].Fallback
	}

	// Parse the key into folder and key parts.
	folder, keyPart := parseKey(key)

	// For JSON translations, there is only one file per locale, so we will
	// simply load the file and return the line if it exists.
	// If the file doesn't exist, we will return fallback if it is enabled.
	// Otherwise, we will return the key as the line.
	if err := t.load(folder, locale); err != nil && err != ErrFileNotExist {
		return "", err
	}

	line := t.loaded[folder][locale][keyPart]
	if line == "" {
		fallbackFolder, fallbackLocale := parseKey(t.GetFallback())
		// If the fallback locale is different from the current locale, we will
		// load in the lines for the fallback locale and try to retrieve the
		// translation for the given key.If it is translated, we will return it.
		// Otherwise, we can finally return the key as that will be the final
		// fallback.
		if (folder+locale != fallbackFolder+fallbackLocale) && fallback {
			var fallbackOptions translationcontract.Option
			if len(options) > 0 {
				fallbackOptions = options[0]
			}
			fallbackOptions.Fallback = translationcontract.Bool(false)
			fallbackOptions.Locale = fallbackLocale
			return t.Get(fallbackFolder+"."+keyPart, fallbackOptions)
		}
		return t.key, nil
	}

	// If the line doesn't contain any placeholders, we can return it right
	// away.Otherwise, we will make the replacements on the line and return
	// the result.
	if len(options) > 0 {
		return makeReplacements(line, options[0].Replace), nil
	}

	return line, nil
}

func (t *Translator) GetFallback() string {
	if fallback, ok := t.ctx.Value(string(fallbackLocaleKey)).(string); ok {
		return fallback
	}
	return t.fallback
}

func (t *Translator) GetLocale() string {
	if locale, ok := t.ctx.Value(string(localeKey)).(string); ok {
		return locale
	}
	return t.locale
}

func (t *Translator) Has(key string, options ...translationcontract.Option) bool {
	line, err := t.Get(key, options...)
	return err == nil && line != key
}

func (t *Translator) SetFallback(locale string) context.Context {
	t.fallback = locale
	//nolint:all
	t.ctx = context.WithValue(t.ctx, string(fallbackLocaleKey), locale)

	return t.ctx
}

func (t *Translator) SetLocale(locale string) context.Context {
	t.locale = locale
	if ctx, ok := t.ctx.(http.Context); ok {
		ctx.WithValue(string(localeKey), locale)
		t.ctx = ctx
	} else {
		//nolint:all
		t.ctx = context.WithValue(t.ctx, string(localeKey), locale)
	}
	return t.ctx
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
