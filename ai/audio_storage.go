package ai

import (
	"github.com/goravel/framework/errors"
)

type audioStorer struct{}

var audioStorePathErrors = storePathErrors{
	pathRequired: errors.AIAudioStorePathRequired,
	nameRequired: errors.AIAudioNameRequired,
	pathInvalid:  errors.AIAudioStorePathInvalid,
}

func (r audioStorer) Store(content []byte, name string, disk string) (string, error) {
	return storeContent(content, name, "", disk, errors.AIAudioNameRequired)
}

func (r audioStorer) StoreAs(content []byte, targetPath string, disk string) (string, error) {
	name, directory, err := normalizeStoreTargetPath(targetPath, audioStorePathErrors)
	if err != nil {
		return "", err
	}

	return storeContent(content, name, directory, disk, errors.AIAudioNameRequired)
}
