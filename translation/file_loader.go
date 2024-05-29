package translation

import (
	"os"
	"path/filepath"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/file"
)

type FileLoader struct {
	paths []string
	json  foundation.Json
}

func NewFileLoader(paths []string, json foundation.Json) *FileLoader {
	return &FileLoader{
		paths: paths,
		json:  json,
	}
}

func (f *FileLoader) Load(locale string, group string) (map[string]any, error) {
	for _, path := range f.paths {
		var val map[string]any
		fullPath := filepath.Join(path, locale, group+".json")
		if group == "*" {
			fullPath = filepath.Join(path, locale+".json")
		}

		if file.Exists(fullPath) {
			data, err := os.ReadFile(fullPath)
			if err != nil {
				return nil, err
			}
			if err := f.json.Unmarshal(data, &val); err != nil {
				return nil, err
			}
			return val, nil
		}
	}
	return nil, ErrFileNotExist
}
