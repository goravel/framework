package ai

import (
	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
)

type imageStorer struct{}

var imageStorePathErrors = storePathErrors{
	pathRequired: errors.AIImageStorePathRequired,
	nameRequired: errors.AIImageNameRequired,
	pathInvalid:  errors.AIImageStorePathInvalid,
}

func NewImageStorer() contractsai.ImageStorer {
	return imageStorer{}
}

func (r imageStorer) Store(content []byte, name string, disk string) (string, error) {
	return storeContent(content, name, "", disk, errors.AIImageNameRequired)
}

func (r imageStorer) StoreAs(content []byte, targetPath string, disk string) (string, error) {
	name, directory, err := normalizeStoreTargetPath(targetPath, imageStorePathErrors)
	if err != nil {
		return "", err
	}

	return storeContent(content, name, directory, disk, errors.AIImageNameRequired)
}
