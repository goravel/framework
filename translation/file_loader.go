package translation

import (
	"os"
	"path/filepath"

	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/json"
)

type FileLoader struct {
	paths []string
}

func NewFileLoader(paths []string) *FileLoader {
	return &FileLoader{
		paths: paths,
	}
}

func (f *FileLoader) Load(locale string, group string) (map[string]map[string]any, error) {
	translations := make(map[string]map[string]any)
	for _, path := range f.paths {
		var val map[string]any
		fullPath := filepath.Join(path, locale, group+".json")

		if file.Exists(fullPath) {
			data, err := os.ReadFile(fullPath)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(data, &val); err != nil {
				return nil, err
			}
			// Initialize the map if it's a nil
			if translations[group] == nil {
				translations[group] = make(map[string]any)
			}
			mergeMaps(translations[group], val)
		} else {
			return nil, ErrFileNotExist
		}
	}
	return translations, nil
}

func mergeMaps[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) {
	for k, v := range src {
		dst[k] = v
	}
}
