package translation

//go:generate mockery --name=Translator
type Translator interface {
	Get(key string, options ...Option) (string, error)
	Has(key string, options ...Option) bool
	GetLocale() string
	SetLocale(locale string)
	GetFallback() string
	SetFallback(locale string)
}

// Choice key, number, options => string [Get a translation for a given key]

type Option struct {
	Fallback bool
	Locale   string
	Replace  map[string]string
}
