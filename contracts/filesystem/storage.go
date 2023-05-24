package filesystem

import (
	"context"
	"time"
)

//go:generate mockery --name=Storage
type Storage interface {
	Driver
	Disk(disk string) Driver
}

//go:generate mockery --name=Driver
type Driver interface {
	AllDirectories(path string) ([]string, error)
	AllFiles(path string) ([]string, error)
	Copy(oldFile, newFile string) error
	Delete(file ...string) error
	DeleteDirectory(directory string) error
	Directories(path string) ([]string, error)
	Exists(file string) bool
	Files(path string) ([]string, error)
	Get(file string) (string, error)
	LastModified(file string) (time.Time, error)
	MakeDirectory(directory string) error
	MimeType(file string) (string, error)
	Missing(file string) bool
	Move(oldFile, newFile string) error
	Path(file string) string
	Put(file, content string) error
	PutFile(path string, source File) (string, error)
	PutFileAs(path string, source File, name string) (string, error)
	Size(file string) (int64, error)
	TemporaryUrl(file string, time time.Time) (string, error)
	WithContext(ctx context.Context) Driver
	Url(file string) string
}

//go:generate mockery --name=File
type File interface {
	Disk(disk string) File
	Extension() (string, error)
	File() string
	GetClientOriginalName() string
	GetClientOriginalExtension() string
	HashName(path ...string) string
	LastModified() (time.Time, error)
	MimeType() (string, error)
	Size() (int64, error)
	Store(path string) (string, error)
	StoreAs(path string, name string) (string, error)
}
