package session

import (
	"github.com/gookit/color"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
	"os"
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

func (f *FileHandler) Gc(maxLifetime int) (int, bool) {
	return 0, true
}

func (f *FileHandler) Open(path string, name string) bool {
	return true
}

func (f *FileHandler) Read(id string) string {
	path := f.path + "/" + id
	color.Yellowln(path, file.Exists(path))
	if file.Exists(path) {
		modified, err := file.LastModified(path, "UTC")
		if err != nil {
			return ""
		}
		color.Greenln(modified, err, modified.Unix() >= carbon.Now().SubMinutes(f.minutes).Timestamp())
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
