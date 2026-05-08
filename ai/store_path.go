package ai

import (
	pathpkg "path"
	"strings"

	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/errors"
)

type storePathErrors struct {
	pathRequired error
	nameRequired error
	pathInvalid  error
}

func normalizeStoreTargetPath(targetPath string, pathErrors storePathErrors) (string, string, error) {
	if targetPath == "" {
		return "", "", pathErrors.pathRequired
	}

	normalizedPath := pathpkg.Clean(strings.ReplaceAll(targetPath, "\\", "/"))
	if normalizedPath == "." || strings.HasSuffix(targetPath, "/") || strings.HasSuffix(targetPath, "\\") {
		return "", "", pathErrors.nameRequired
	}
	if pathpkg.IsAbs(normalizedPath) || hasParentPathSegment(normalizedPath) {
		return "", "", pathErrors.pathInvalid
	}

	name := pathpkg.Base(normalizedPath)
	if name == "." || name == "/" || name == "" {
		return "", "", pathErrors.nameRequired
	}

	directory := pathpkg.Dir(normalizedPath)
	if directory == "." {
		directory = ""
	}

	return name, directory, nil
}

func storeContent(content []byte, name, path, disk string, nameRequired error) (string, error) {
	if storageFacade == nil {
		return "", errors.StorageFacadeNotSet
	}
	if name == "" {
		return "", nameRequired
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

func hasParentPathSegment(path string) bool {
	for _, segment := range strings.Split(path, "/") {
		if segment == ".." {
			return true
		}
	}

	return false
}
