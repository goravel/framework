package file

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func Create(file string, content string) {
	err := os.MkdirAll(path.Dir(file), os.ModePerm)

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

func Exist(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
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
	if err != nil {
		return false
	}

	return true
}

func Contain(file string, search string) bool {
	if Exist(file) {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return false
		}
		return strings.Contains(string(data), search)
	}

	return false
}
