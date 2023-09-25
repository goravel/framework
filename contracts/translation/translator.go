package translation

//go:generate mockery --name=Translator
type Translator interface {
	Get(key string, options ...Option) (string, error)
	Choice(key string, number int, options ...Option) (string, error)
	Has(key string, options ...Option) bool
	GetLocale() string
	SetLocale(locale string)
	GetFallback() string
	SetFallback(locale string)
}

type Option struct {
	Fallback *bool
	Locale   string
	Replace  map[string]string
}

func Bool(value bool) *bool {
	return &value
}
