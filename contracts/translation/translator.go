package translation

import (
	"github.com/goravel/framework/contracts/http"
)

//go:generate mockery --name=Translator
type Translator interface {
	Get(ctx http.Context, key string, options ...Option) (string, error)
	Choice(ctx http.Context, key string, number int, options ...Option) (string, error)
	Has(ctx http.Context, key string, options ...Option) bool
	GetLocale(ctx http.Context) string
	SetLocale(ctx http.Context, locale string) error
	GetFallback(ctx http.Context) string
	SetFallback(ctx http.Context, locale string) error
}

type Option struct {
	Fallback *bool
	Locale   string
	Replace  map[string]string
}

func Bool(value bool) *bool {
	return &value
}
