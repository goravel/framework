package modular

import (
	"io"
	"io/fs"
	"strings"
	"time"
)

var (
	_ fs.ReadDirFS  = (*prefixFS)(nil)
	_ fs.ReadFileFS = (*prefixFS)(nil)
	_ fs.StatFS     = (*prefixFS)(nil)
)

// prefixFS mounts inner at name: "blog/index.tmpl" is inner's "index.tmpl".
type prefixFS struct {
	name  string
	inner fs.FS
}

// NewPrefixFS mounts inner at name, a single path element.
func NewPrefixFS(name string, inner fs.FS) fs.FS {
	return &prefixFS{name: name, inner: inner}
}

func (r *prefixFS) Open(name string) (fs.File, error) {
	inner, kind, err := r.resolve("open", name)
	if err != nil {
		return nil, err
	}

	switch kind {
	case pathRoot:
		info, err := r.mountInfo()
		if err != nil {
			return nil, &fs.PathError{Op: "open", Path: name, Err: err}
		}

		return &rootDir{info: rootInfo{modTime: info.ModTime()}, entries: []fs.DirEntry{fs.FileInfoToDirEntry(info)}}, nil
	case pathMount:
		file, err := r.inner.Open(".")
		if err != nil {
			return nil, &fs.PathError{Op: "open", Path: name, Err: err}
		}

		return &renamedFile{File: file, name: r.name}, nil
	default:
		return r.inner.Open(inner)
	}
}

func (r *prefixFS) ReadDir(name string) ([]fs.DirEntry, error) {
	inner, kind, err := r.resolve("readdir", name)
	if err != nil {
		return nil, err
	}

	if kind == pathRoot {
		info, err := r.mountInfo()
		if err != nil {
			return nil, &fs.PathError{Op: "readdir", Path: name, Err: err}
		}

		return []fs.DirEntry{fs.FileInfoToDirEntry(info)}, nil
	}

	return fs.ReadDir(r.inner, inner)
}

func (r *prefixFS) ReadFile(name string) ([]byte, error) {
	inner, kind, err := r.resolve("readfile", name)
	if err != nil {
		return nil, err
	}
	if kind != pathInner {
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: fs.ErrInvalid}
	}

	return fs.ReadFile(r.inner, inner)
}

func (r *prefixFS) Stat(name string) (fs.FileInfo, error) {
	inner, kind, err := r.resolve("stat", name)
	if err != nil {
		return nil, err
	}

	switch kind {
	case pathRoot:
		info, err := r.mountInfo()
		if err != nil {
			return nil, &fs.PathError{Op: "stat", Path: name, Err: err}
		}

		return rootInfo{modTime: info.ModTime()}, nil
	case pathMount:
		return r.mountInfo()
	default:
		return fs.Stat(r.inner, inner)
	}
}

type pathKind int

const (
	pathRoot  pathKind = iota // "."
	pathMount                 // name
	pathInner                 // name/x
)

// resolve classifies name and returns inner's path for pathInner.
func (r *prefixFS) resolve(op, name string) (string, pathKind, error) {
	if !fs.ValidPath(name) {
		return "", 0, &fs.PathError{Op: op, Path: name, Err: fs.ErrInvalid}
	}

	switch {
	case name == ".":
		return "", pathRoot, nil
	case name == r.name:
		return ".", pathMount, nil
	case strings.HasPrefix(name, r.name+"/"):
		return strings.TrimPrefix(name, r.name+"/"), pathInner, nil
	default:
		return "", 0, &fs.PathError{Op: op, Path: name, Err: fs.ErrNotExist}
	}
}

// mountInfo is inner's root, renamed to the mount point.
func (r *prefixFS) mountInfo() (fs.FileInfo, error) {
	info, err := fs.Stat(r.inner, ".")
	if err != nil {
		return nil, err
	}

	return renamedInfo{FileInfo: info, name: r.name}, nil
}

// rootInfo describes the synthetic root.
type rootInfo struct{ modTime time.Time }

func (i rootInfo) Name() string       { return "." }
func (i rootInfo) Size() int64        { return 0 }
func (i rootInfo) Mode() fs.FileMode  { return fs.ModeDir | 0o555 }
func (i rootInfo) ModTime() time.Time { return i.modTime }
func (i rootInfo) IsDir() bool        { return true }
func (i rootInfo) Sys() any           { return nil }

// rootDir lists only the mount point.
type rootDir struct {
	info    rootInfo
	entries []fs.DirEntry
	offset  int
}

func (d *rootDir) Stat() (fs.FileInfo, error) { return d.info, nil }

func (d *rootDir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: ".", Err: fs.ErrInvalid}
}

func (d *rootDir) Close() error { return nil }

func (d *rootDir) ReadDir(n int) ([]fs.DirEntry, error) {
	return readDirEntries(d.entries, &d.offset, n)
}

// renamedInfo overrides a file's name.
type renamedInfo struct {
	fs.FileInfo
	name string
}

func (i renamedInfo) Name() string { return i.name }

// renamedFile is inner's root, named after the mount point.
type renamedFile struct {
	fs.File
	name string
}

func (f *renamedFile) Stat() (fs.FileInfo, error) {
	info, err := f.File.Stat()
	if err != nil {
		return nil, err
	}

	return renamedInfo{FileInfo: info, name: f.name}, nil
}

func (f *renamedFile) ReadDir(n int) ([]fs.DirEntry, error) {
	dir, ok := f.File.(fs.ReadDirFile)
	if !ok {
		return nil, &fs.PathError{Op: "readdir", Path: f.name, Err: fs.ErrInvalid}
	}

	return dir.ReadDir(n)
}

// readDirEntries implements ReadDir over a fixed list.
func readDirEntries(entries []fs.DirEntry, offset *int, n int) ([]fs.DirEntry, error) {
	rest := entries[*offset:]
	if n <= 0 {
		*offset = len(entries)

		return rest, nil
	}
	if len(rest) == 0 {
		return nil, io.EOF
	}
	if n > len(rest) {
		n = len(rest)
	}
	*offset += n

	return rest[:n], nil
}
