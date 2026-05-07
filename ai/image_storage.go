package ai

import (
	pathpkg "path"
	"path/filepath"

	contractsai "github.com/goravel/framework/contracts/ai"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/errors"
)

type imageStorer struct{}

func NewImageStorer() contractsai.ImageStorer {
	return imageStorer{}
}

func (r imageStorer) Store(content []byte, name string, disk string) (string, error) {
	return r.storeContent(content, name, "", disk)
}

func (r imageStorer) StoreAs(content []byte, targetPath string, disk string) (string, error) {
	if targetPath == "" {
		return "", errors.AIImageStorePathRequired
	}

	name := filepath.Base(targetPath)
	if name == "." || name == string(filepath.Separator) || name == "" {
		return "", errors.AIImageNameRequired
	}

	directory := filepath.Dir(targetPath)
	if directory == "." {
		directory = ""
	}

	return r.storeContent(content, name, directory, disk)
}

func (imageStorer) storeContent(content []byte, name, path, disk string) (string, error) {
	if storageFacade == nil {
		return "", errors.StorageFacadeNotSet
	}
	if name == "" {
		return "", errors.AIImageNameRequired
	}

	resolvedPath := name
	if path != "" {
		resolvedPath = pathpkg.Join(path, name)
	}

	driver := contractsfilesystem.Driver(storageFacade)
	if disk != "" {
		driver = storageFacade.Disk(disk)
	}

	if err := driver.Put(resolvedPath, string(content)); err != nil {
		return "", err
	}

	return resolvedPath, nil
}
