package filesystem

import (
	"fmt"
	"os"

	"github.com/goravel/framework/contracts/filesystem"
)

var _ filesystem.Fake = (*Fake)(nil)

// Fake is a local disk in a temporary directory that replaces a disk of
// Storage until Reset.
type Fake struct {
	*Local
	storage  *Storage
	disk     string
	original filesystem.Driver
}

func (r *Fake) AssertCount(path string, count int, recursive ...bool) bool {
	var (
		files []string
		err   error
	)
	if len(recursive) > 0 && recursive[0] {
		files, err = r.AllFiles(path)
	} else {
		files, err = r.Files(path)
	}

	return err == nil && len(files) == count
}

func (r *Fake) AssertDirectoryEmpty(path string) bool {
	files, err := r.AllFiles(path)

	return err == nil && len(files) == 0
}

func (r *Fake) AssertExists(path ...string) bool {
	for _, p := range path {
		if !r.Exists(p) {
			return false
		}
	}

	return true
}

func (r *Fake) AssertMissing(path ...string) bool {
	for _, p := range path {
		if r.Exists(p) {
			return false
		}
	}

	return true
}

func (r *Fake) Reset() error {
	r.storage.restore(r)

	return os.RemoveAll(r.root)
}

// Fake replaces the disk, the default one when none is given, with a local
// disk in a new empty temporary directory, see filesystem.Storage.
func (r *Storage) Fake(disk ...string) filesystem.Fake {
	name := r.defaultDisk
	if len(disk) > 0 && disk[0] != "" {
		name = disk[0]
	}

	root, err := os.MkdirTemp("", "goravel-storage-fake-")
	if err != nil {
		panic(err)
	}

	fake := &Fake{
		Local: &Local{
			config: r.config,
			root:   root,
			url:    r.config.GetString(fmt.Sprintf("filesystems.disks.%s.url", name)),
		},
		storage: r,
		disk:    name,
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	fake.original = r.drivers[name]
	if previous, ok := fake.original.(*Fake); ok {
		// The previous fake is replaced, Reset restores what it replaced.
		fake.original = previous.original
		_ = os.RemoveAll(previous.root)
	}

	r.setDriverLocked(name, fake)

	return fake
}

// restore puts back the disk the fake replaced, unless the fake was already
// replaced or reset.
func (r *Storage) restore(fake *Fake) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.drivers[fake.disk] != fake {
		return
	}

	// A disk that was not loaded yet is loaded again on its next use.
	r.setDriverLocked(fake.disk, fake.original)
}

func (r *Storage) setDriverLocked(disk string, driver filesystem.Driver) {
	if driver == nil {
		delete(r.drivers, disk)
	} else {
		r.drivers[disk] = driver
	}

	if disk == r.defaultDisk && driver != nil {
		r.Driver = driver
	}
}
