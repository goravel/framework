package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"
)

type Local struct {
	root string
	url  string
}

func NewLocal(disk string) (*Local, error) {
	return &Local{
		root: facades.Config.GetString(fmt.Sprintf("filesystems.disks.%s.root", disk)),
		url:  facades.Config.GetString(fmt.Sprintf("filesystems.disks.%s.url", disk)),
	}, nil
}

func (r *Local) WithContext(ctx context.Context) filesystem.Driver {
	return r
}

func (r *Local) Put(file, content string) error {
	file = r.fullPath(file)
	if err := os.MkdirAll(path.Dir(file), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(file)
	defer f.Close()
	if err != nil {
		return err
	}

	if _, err = f.WriteString(content); err != nil {
		return err
	}

	return nil
}

func (r *Local) PutFile(filePath string, source filesystem.File) (string, error) {
	return r.PutFileAs(filePath, source, str.Random(40))
}

func (r *Local) PutFileAs(filePath string, source filesystem.File, name string) (string, error) {
	data, err := ioutil.ReadFile(source.File())
	if err != nil {
		return "", err
	}

	fullPath, err := fullPathOfFile(filePath, source, name)
	if err != nil {
		return "", err
	}

	if err := r.Put(fullPath, string(data)); err != nil {
		return "", err
	}

	return fullPath, nil
}

func (r *Local) Get(file string) (string, error) {
	data, err := ioutil.ReadFile(r.fullPath(file))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (r *Local) Size(file string) (int64, error) {
	fileInfo, err := os.Open(r.fullPath(file))
	if err != nil {
		return 0, err
	}

	fi, err := fileInfo.Stat()
	if err != nil {
		return 0, err
	}

	return fi.Size(), nil
}

func (r *Local) Path(file string) string {
	var abPath string
	_, filename, _, ok := runtime.Caller(1)
	if ok {
		abPath = path.Dir(filename)
	}

	return abPath + "/" + strings.TrimPrefix(strings.TrimPrefix(r.fullPath(file), "/"), "./")
}

func (r *Local) Exists(file string) bool {
	_, err := os.Stat(r.fullPath(file))
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func (r *Local) Missing(file string) bool {
	return !r.Exists(file)
}

func (r *Local) Url(file string) string {
	return strings.TrimSuffix(r.url, "/") + "/" + strings.TrimPrefix(file, "/")
}

func (r *Local) TemporaryUrl(file string, time time.Time) (string, error) {
	return r.Url(file), nil
}

func (r *Local) Copy(originFile, targetFile string) error {
	content, err := r.Get(originFile)
	if err != nil {
		return err
	}

	return r.Put(targetFile, content)
}

func (r *Local) Move(oldFile, newFile string) error {
	newFile = r.fullPath(newFile)
	if err := os.MkdirAll(path.Dir(newFile), os.ModePerm); err != nil {
		return err
	}

	if err := os.Rename(r.fullPath(oldFile), newFile); err != nil {
		return err
	}

	return nil
}

func (r *Local) Delete(files ...string) error {
	for _, file := range files {
		fileInfo, err := os.Stat(r.fullPath(file))
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			return errors.New("can't delete directory, please use DeleteDirectory")
		}
	}

	for _, file := range files {
		if err := os.Remove(r.fullPath(file)); err != nil {
			return err
		}
	}

	return nil
}

func (r *Local) MakeDirectory(directory string) error {
	return os.MkdirAll(path.Dir(r.fullPath(directory)+"/"), os.ModePerm)
}

func (r *Local) DeleteDirectory(directory string) error {
	return os.RemoveAll(r.fullPath(directory))
}

func (r *Local) fullPath(path string) string {
	return strings.TrimSuffix(r.root, "/") + "/" + strings.TrimPrefix(path, "/")
}
