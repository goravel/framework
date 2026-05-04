package ai

import (
	"path/filepath"
	pathpkg "path"

	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/errors"
)

func StoreImageContent(content []byte, name string, args ...string) (string, error) {
	if storageFacade == nil {
		return "", errors.StorageFacadeNotSet
	}
	if name == "" {
		return "", errors.AIImageNameRequired
	}
	if len(args) > 2 {
		return "", errors.AIImageStoreTooManyPaths
	}

	var path, disk string
	if len(args) > 0 {
		path = args[0]
	}
	if len(args) > 1 {
		disk = args[1]
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

func StoreImage(content []byte, name string, disk ...string) (string, error) {
	if len(disk) > 1 {
		return "", errors.AIImageStoreTooManyPaths
	}

	if len(disk) == 1 {
		return StoreImageContent(content, name, "", disk[0])
	}

	return StoreImageContent(content, name)
}

func StoreImageContentAs(content []byte, targetPath string, disk ...string) (string, error) {
	if targetPath == "" {
		return "", errors.AIImageStorePathRequired
	}
	if len(disk) > 1 {
		return "", errors.AIImageStoreTooManyPaths
	}

	name := filepath.Base(targetPath)
	if name == "." || name == string(filepath.Separator) || name == "" {
		return "", errors.AIImageNameRequired
	}

	directory := filepath.Dir(targetPath)
	if directory == "." {
		directory = ""
	}

	args := []string{directory}
	if len(disk) == 1 {
		args = append(args, disk[0])
	}

	return StoreImageContent(content, name, args...)
}
