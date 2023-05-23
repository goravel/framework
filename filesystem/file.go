package filesystem

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	supportfile "github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type File struct {
	disk string
	path string
	name string
}

func NewFile(file string) (*File, error) {
	if !supportfile.Exists(file) {
		return nil, errors.New("file doesn't exist")
	}

	disk := facades.Config.GetString("filesystems.default")

	return &File{disk: disk, path: file, name: path.Base(file)}, nil
}

func NewFileFromRequest(fileHeader *multipart.FileHeader) (*File, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	tempFileName := fmt.Sprintf("%s_*%s", facades.Config.GetString("app.name"), path.Ext(fileHeader.Filename))
	tempFile, err := ioutil.TempFile(os.TempDir(), tempFileName)
	if err != nil {
		return nil, err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, src)
	if err != nil {
		return nil, err
	}

	disk := facades.Config.GetString("filesystems.default")

	return &File{disk: disk, path: tempFile.Name(), name: fileHeader.Filename}, nil
}

func (f *File) Disk(disk string) filesystem.File {
	f.disk = disk

	return f
}

func (f *File) Extension() (string, error) {
	return supportfile.Extension(f.path)
}

func (f *File) File() string {
	return f.path
}

func (f *File) GetClientOriginalName() string {
	return f.name
}

func (f *File) GetClientOriginalExtension() string {
	return supportfile.ClientOriginalExtension(f.name)
}

func (f *File) HashName(path ...string) string {
	var realPath string
	if len(path) > 0 {
		realPath = strings.TrimRight(path[0], "/") + "/"
	}

	extension, _ := supportfile.Extension(f.path, true)
	if extension == "" {
		return realPath + str.Random(40)
	}

	return realPath + str.Random(40) + "." + extension
}

func (f *File) LastModified() (time.Time, error) {
	return supportfile.LastModified(f.path, facades.Config.GetString("app.timezone"))
}

func (f *File) MimeType() (string, error) {
	return supportfile.MimeType(f.path)
}

func (f *File) Size() (int64, error) {
	return supportfile.Size(f.path)
}

func (f *File) Store(path string) (string, error) {
	return facades.Storage.Disk(f.disk).PutFile(path, f)
}

func (f *File) StoreAs(path string, name string) (string, error) {
	return facades.Storage.Disk(f.disk).PutFileAs(path, f, name)
}
