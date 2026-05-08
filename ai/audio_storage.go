package ai

import (
	pathpkg "path"
	"strings"

	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/errors"
)

type audioStorer struct{}

func (r audioStorer) Store(content []byte, name string, disk string) (string, error) {
	return r.storeContent(content, name, "", disk)
}

func (r audioStorer) StoreAs(content []byte, targetPath string, disk string) (string, error) {
	if targetPath == "" {
		return "", errors.AIAudioStorePathRequired
	}

	normalizedPath := pathpkg.Clean(strings.ReplaceAll(targetPath, "\\", "/"))
	if normalizedPath == "." || strings.HasSuffix(targetPath, "/") || strings.HasSuffix(targetPath, "\\") {
		return "", errors.AIAudioNameRequired
	}
	if pathpkg.IsAbs(normalizedPath) || hasParentPathSegment(normalizedPath) {
		return "", errors.AIAudioStorePathInvalid
	}

	name := pathpkg.Base(normalizedPath)
	if name == "." || name == "/" || name == "" {
		return "", errors.AIAudioNameRequired
	}

	directory := pathpkg.Dir(normalizedPath)
	if directory == "." {
		directory = ""
	}

	return r.storeContent(content, name, directory, disk)
}

func (audioStorer) storeContent(content []byte, name, path, disk string) (string, error) {
	if storageFacade == nil {
		return "", errors.StorageFacadeNotSet
	}
	if name == "" {
		return "", errors.AIAudioNameRequired
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
