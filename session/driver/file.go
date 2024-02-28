package driver

import (
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/support/carbon"
)

type FileDriver struct {
	files   filesystem.Storage
	path    string
	minutes int
}

func NewFileDriver(files filesystem.Storage, path string, minutes int) *FileDriver {
	return &FileDriver{
		files:   files,
		path:    path,
		minutes: minutes,
	}
}

func (f *FileDriver) Close() bool {
	return true
}

func (f *FileDriver) Destroy(id string) error {
	return f.files.Delete(f.path + "/" + id)
}

func (f *FileDriver) Gc(maxLifetime int) int {
	cutoffTime := carbon.Now().SubSeconds(maxLifetime)
	deletedSessions := 0

	files, err := f.files.Files(f.path)
	if err != nil {
		return 0
	}

	for _, file := range files {
		modified, err := f.files.LastModified(f.path + "/" + file)
		if err != nil {
			continue
		}

		if modified.Before(cutoffTime.StdTime()) {
			err = f.files.Delete(f.path + "/" + file)
			if err == nil {
				deletedSessions++
			}
		}
	}

	return deletedSessions
}

func (f *FileDriver) Open(string, string) bool {
	return true
}

func (f *FileDriver) Read(id string) string {
	path := f.path + "/" + id
	if f.files.Exists(path) {
		modified, err := f.files.LastModified(path)
		if err != nil {
			return ""
		}

		if modified.After(carbon.Now().SubMinutes(f.minutes).StdTime()) {
			data, err := f.files.Get(path)
			if err != nil {
				return ""
			}
			return data
		}
	}

	return ""
}

func (f *FileDriver) Write(id string, data string) error {
	return f.files.Put(f.path+"/"+id, data)
}
