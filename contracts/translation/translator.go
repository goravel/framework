package translation

import (
	"context"

	"github.com/goravel/framework/contracts/http"
)

//go:generate mockery --name=Translator
type Translator interface {
	// Get the translation for the given key.
	Get(key string, options ...Option) (string, error)
	// Choice gets a translation according to an integer value.
	Choice(key string, number int, options ...Option) (string, error)
	// Has checks if a translation exists for a given key.
	Has(key string, options ...Option) bool
	// GetLocale get the current application/context locale.
	GetLocale() string
	// SetLocale set the current application/context locale.
	SetLocale(locale string) context.Context
	// SetLocaleByHttp set the current application/context locale by http request.
	SetLocaleByHttp(ctx http.Context, locale string)
	// GetFallback get the current application/context fallback locale.
	GetFallback() string
	// SetFallback set the current application/context fallback locale.
	SetFallback(locale string) context.Context
}

type Option struct {
	Fallback *bool
	Locale   string
	Replace  map[string]string
}

func Bool(value bool) *bool {
	return &value
}
