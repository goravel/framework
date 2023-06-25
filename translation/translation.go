package translation

import (
	"github.com/spf13/viper"
	"path/filepath"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/translation"
)

var _ translation.Translation = &Application{}

type Application struct {
	vip       *viper.Viper
	config    config.Config
	languages map[string]translation.Language
}

func NewTranslation(config config.Config) *Application {
	app := &Application{}
	app.config = config
	app.vip = viper.New()

	app.languages = make(map[string]translation.Language)
	return app
}

func (app *Application) GetDefaultLocale() string {
	return app.config.GetString("app.locale", "en")
}

func (app *Application) Language(locale string) translation.Language {
	if _, ok := app.languages[locale]; !ok {
		app.languages[locale] = NewLanguage()
	}

	return app.languages[locale]
}

func (app *Application) Load(path string, locale ...string) error {
	// load translations from folder
	files, err := getDirsInPath(path)
	if err != nil {
		return err
	}
	for _, file := range files {
		// get locale from last folder name
		l := filepath.Base(file)
		err := app.Language(l).Load(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *Application) Add(name string, locale string, data any) {
	app.Language(locale).Add(name, data)
}

func (app *Application) Get(word string, locale string, replace ...any) string {
	return app.Language(locale).Get(word, replace...)
}
