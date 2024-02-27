package driver

import (
	"os"
	"path/filepath"

	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

type FileDriver struct {
	path    string
	minutes int
}

func NewFileDriver(path string, minutes int) *FileDriver {
	return &FileDriver{
		path:    path,
		minutes: minutes,
	}
}

func (f *FileDriver) Close() bool {
	return true
}

func (f *FileDriver) Destroy(id string) error {
	return file.Remove(f.path + "/" + id)
}

func (f *FileDriver) Gc(maxLifetime int) int {
	cutoffTime := carbon.Now().SubSeconds(maxLifetime)
	deletedSessions := 0

	_ = filepath.Walk(f.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.ModTime().Before(cutoffTime.StdTime()) {
			err := os.Remove(path)
			if err == nil {
				deletedSessions++
			}
		}

		return nil
	})

	return deletedSessions
}

func (f *FileDriver) Open(string, string) bool {
	return true
}

func (f *FileDriver) Read(id string) string {
	path := f.path + "/" + id
	if file.Exists(path) {
		modified, err := file.LastModified(path, "UTC")
		if err != nil {
			return ""
		}

		if modified.Unix() >= carbon.Now(carbon.UTC).SubMinutes(f.minutes).StdTime().Unix() {
			data, err := os.ReadFile(path)
			if err != nil {
				return ""
			}
			return string(data)
		}
	}

	return ""
}

func (f *FileDriver) Write(id string, data string) error {
	return file.Create(f.path+"/"+id, data)
}
