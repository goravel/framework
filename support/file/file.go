package file

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

func ClientOriginalExtension(file string) string {
	return strings.ReplaceAll(path.Ext(file), ".", "")
}

func Contain(file string, search string) bool {
	if Exists(file) {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return false
		}
		return strings.Contains(string(data), search)
	}

	return false
}

func Create(file string, content string) {
	err := os.MkdirAll(path.Dir(file), os.ModePerm)
	if err != nil {
		panic(err.Error())
	}

	f, err := os.Create(file)
	defer func() {
		f.Close()
	}()

	if err != nil {
		panic(err.Error())
	}

	_, err = f.WriteString(content)
	if err != nil {
		panic(err.Error())
	}
}

func Exists(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

//Extension Supported types: https://github.com/gabriel-vasile/mimetype/blob/master/supported_mimes.md
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

func Remove(file string) bool {
	fi, err := os.Stat(file)
	if err != nil {
		return false
	}

	if fi.IsDir() {
		dir, err := ioutil.ReadDir(file)

		if err != nil {
			return false
		}

		for _, d := range dir {
			err := os.RemoveAll(path.Join([]string{file, d.Name()}...))
			if err != nil {
				return false
			}
		}
	}

	err = os.Remove(file)

	return err == nil
}

func Size(file string) (int64, error) {
	fileInfo, err := os.Open(file)
	if err != nil {
		return 0, err
	}

	fi, err := fileInfo.Stat()
	if err != nil {
		return 0, err
	}

	return fi.Size(), nil
}
