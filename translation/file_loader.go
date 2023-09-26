package translation

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bytedance/sonic"

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
				return nil, err
			}
			translations[locale] = val
		} else {
			return nil, ErrFileNotExist
		}
	}
	return translations, nil
}
