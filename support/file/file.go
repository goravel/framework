package file

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

func ClientOriginalExtension(file string) string {
	return strings.ReplaceAll(filepath.Ext(file), ".", "")
}

func Contain(file string, search string) bool {
	if Exists(file) {
		data, err := os.ReadFile(file)
		if err != nil {
			return false
		}
		return strings.Contains(string(data), search)
	}

	return false
}

func Create(file string, content string) error {
	if err := os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(content); err != nil {
		return err
	}

	return nil
}

func Exists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

// Extension Supported types: https://github.com/gabriel-vasile/mimetype/blob/master/supported_mimes.md
func Extension(file string, originalWhenUnknown ...bool) (string, error) {
	mtype, err := mimetype.DetectFile(file)
	if err != nil {
		return "", err
	}

	if mtype.String() == "" {
		if len(originalWhenUnknown) > 0 {
			if originalWhenUnknown[0] {
				return ClientOriginalExtension(file), nil
			}
		}

		return "", errors.New("unknown file extension")
	}

	return strings.TrimPrefix(mtype.Extension(), "."), nil
}

func LastModified(file, timezone string) (time.Time, error) {
	fileInfo, err := os.Stat(file)
	if err != nil {
		return time.Time{}, err
	}

	l, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}

	return fileInfo.ModTime().In(l), nil
}

func MimeType(file string) (string, error) {
	mtype, err := mimetype.DetectFile(file)
	if err != nil {
		return "", err
	}

	return mtype.String(), nil
}

func Remove(file string) error {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	return os.RemoveAll(file)
}

func Size(file string) (int64, error) {
	fileInfo, err := os.Open(file)
	if err != nil {
		return 0, err
	}
	defer fileInfo.Close()

	fi, err := fileInfo.Stat()
	if err != nil {
		return 0, err
	}

	return fi.Size(), nil
}
