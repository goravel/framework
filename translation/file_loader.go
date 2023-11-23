package translation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/decoder"

	"github.com/goravel/framework/support/file"
)

type FileLoader struct {
	paths []string
}

func NewFileLoader(paths []string) *FileLoader {
	return &FileLoader{
		paths: paths,
	}
}

func (f *FileLoader) Load(folder string, locale string) (map[string]map[string]string, error) {
	translations := make(map[string]map[string]string)
	for _, path := range f.paths {
		var val map[string]string
		fullPath := path
		// Check if the folder is not "*", and if so, split it into subFolders
		if folder != "*" {
			subFolders := strings.Split(folder, ".")
			for _, subFolder := range subFolders {
				fullPath = filepath.Join(fullPath, subFolder)
			}
		}
		fullPath = filepath.Join(fullPath, locale+".json")

		if file.Exists(fullPath) {
			data, err := os.ReadFile(fullPath)
			if err != nil {
				return nil, err
			}
			if err := sonic.Unmarshal(data, &val); err != nil {
				if _, ok := err.(decoder.SyntaxError); ok {
					return nil, fmt.Errorf("translation file [%s] contains an invalid JSON structure", fullPath)
				} else if _, ok := err.(*decoder.MismatchTypeError); ok {
					return nil, fmt.Errorf("translation file [%s] contains mismatched types", fullPath)
				}
				return nil, err
			}
			// Initialize the map if it's a nil
			if translations[locale] == nil {
				translations[locale] = make(map[string]string)
			}
			mergeMaps(translations[locale], val)
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
