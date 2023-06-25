package translation

import (
	"encoding/json"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/translation"
)

var _ translation.Language = &Language{}

type Language struct {
	vip *viper.Viper
}

func NewLanguage() *Language {
	app := &Language{}
	app.vip = viper.New()

	return app
}

func (app *Language) Load(path string) error {
	files, err := filePathWalkDir(path)
	if err != nil {
		return err
	}
	for _, file := range files {
		jsonFile, err := os.Open(file)
		if err != nil {
			continue
		}

		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(byteValue), &data); err != nil {
			return err
		}

		key := app.extractKey(path, file)

		app.vip.Set(key, data)
	}
	return nil
}

func (app *Language) extractKey(path, file string) string {
	// first remove .json
	file = strings.TrimSuffix(file, filepath.Ext(file))

	// remove path
	file = strings.TrimPrefix(file, path+"/")

	// replace all / with dots
	file = strings.ReplaceAll(file, "/", ".")

	return file
}

// Add config to application.
func (app *Language) Add(name string, configuration any) {
	app.vip.Set(name, configuration)
}

// Get config from application.
func (app *Language) Get(word string, replace ...any) string {
	if !app.vip.IsSet(word) {
		return ""
	}

	if len(replace) == 0 {
		return cast.ToString(app.vip.Get(word))
	}

	return parseTranslation(cast.ToString(app.vip.Get(word)), replace[0].(http.Json))
}
