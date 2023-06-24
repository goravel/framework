package translation

//go:generate mockery --name=Translation
type Translation interface {
	// Load translations from path.
	Load(path string, locale ...string) error
	GetDefaultLocale() string
	//
	Language(locale string) Language
	// Add translations to application.
	Add(name, locale string, val any)
	// Get translation from application.
	Get(path string, locale string, options ...any) string
}

//go:generate mockery --name=Language
type Language interface {
	// Load translations from path.
	Load(path string) error
	// Add translations to application.
	Add(name string, val any)
	// Get translation from application.
	Get(path string, options ...any) string
}
