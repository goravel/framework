package handler

import (
	"os"
	"path/filepath"

	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

type FileHandler struct {
	path    string
	minutes int
}

func NewFileHandler(path string, minutes int) *FileHandler {
	return &FileHandler{
		path:    path,
		minutes: minutes,
	}
}

func (f *FileHandler) Close() bool {
	return true
}

func (f *FileHandler) Destroy(id string) bool {
	err := file.Remove(f.path + "/" + id)
	return err == nil
}

func (f *FileHandler) Gc(maxLifetime int) int {
	cutoffTime := carbon.Now("UTC").SubSeconds(maxLifetime)
	deletedSessions := 0

	_ = filepath.Walk(f.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.ModTime().Unix() < cutoffTime.Timestamp() {
			err := os.Remove(path)
			if err == nil {
				deletedSessions++
			}
		}

		return nil
	})

	return deletedSessions
}

func (f *FileHandler) Open(path string, name string) bool {
	return true
}

func (f *FileHandler) Read(id string) string {
	path := f.path + "/" + id
	if file.Exists(path) {
		modified, err := file.LastModified(path, "UTC")
		if err != nil {
			return ""
		}

		if modified.Unix() >= carbon.Now().SubMinutes(f.minutes).Timestamp() {
			data, err := os.ReadFile(path)
			if err != nil {
				return ""
			}
			return string(data)
		}
	}

	return ""
}

func (f *FileHandler) Write(id string, data string) error {
	return file.Create(f.path+"/"+id, data)
}
