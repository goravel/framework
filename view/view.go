package view

import (
	"io/fs"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/paths"
	"github.com/goravel/framework/support"
)

type View struct {
	mu          sync.RWMutex
	paths       []string
	filesystems []fs.FS
	shared      sync.Map
}

func NewView() *View {
	return &View{}
}

// Exists reports whether a view file with the given name is available from any source.
// Only files count: directories, "" and "." (which would resolve to a source root) do not match.
func (r *View) Exists(view string) bool {
	if isFile(paths.Abs(support.Config.Paths.Resources, "views", view)) {
		return true
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.paths {
		if isFile(paths.Abs(p, view)) {
			return true
		}
	}

	if len(r.filesystems) == 0 {
		return false
	}

	name := normalizeFSPath(view)
	for _, fsys := range r.filesystems {
		if info, err := fs.Stat(fsys, name); err == nil && !info.IsDir() {
			return true
		}
	}

	return false
}

func (r *View) LoadViewsFrom(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paths = append(r.paths, path)
}

func (r *View) LoadViewsFromFS(fsys fs.FS, root string) {
	if fsys == nil {
		panic(errors.ViewFSRequired)
	}

	name := normalizeFSPath(root)

	// Root the filesystem once at registration so consumers can treat every
	// registered source uniformly, with template paths relative to ".".
	sub, err := fs.Sub(fsys, name)
	if err != nil {
		panic(errors.ViewInvalidFSRoot.Args(root, err))
	}

	// fs.Sub only validates the path syntax; make sure the root really is a directory
	// so a typo does not register a source that silently never matches anything.
	info, err := fs.Stat(fsys, name)
	if err != nil {
		panic(errors.ViewInvalidFSRoot.Args(root, err))
	}
	if !info.IsDir() {
		panic(errors.ViewFSRootNotDirectory.Args(root))
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.filesystems = append(r.filesystems, sub)
}

func (r *View) RegisteredViews() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, len(r.paths))
	copy(out, r.paths)
	return out
}

func (r *View) RegisteredViewFS() []fs.FS {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]fs.FS, len(r.filesystems))
	copy(out, r.filesystems)
	return out
}

func (r *View) Share(key string, value any) {
	r.shared.Store(key, value)
}

func (r *View) Shared(key string, def ...any) any {
	value, ok := r.shared.Load(key)
	if !ok {
		if len(def) > 0 {
			return def[0]
		}

		return nil
	}

	return value
}

func (r *View) GetShared() map[string]any {
	shared := make(map[string]any)
	r.shared.Range(func(key, value any) bool {
		shared[key.(string)] = value
		return true
	})

	return shared
}

func isFile(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// normalizeFSPath converts an OS-style or loosely written path ("", "./views",
// "/views/", `views\admin`) into the slash-separated, unrooted form required by io/fs.
func normalizeFSPath(p string) string {
	p = strings.TrimPrefix(path.Clean(strings.ReplaceAll(p, `\`, "/")), "/")
	if p == "" {
		return "."
	}

	return p
}
