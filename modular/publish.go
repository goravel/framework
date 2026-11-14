package modular

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"maps"
	"os"
	stdpath "path"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/modular"
	"github.com/goravel/framework/errors"
)

const (
	// PublishMarker records what modular:publish wrote into a module directory.
	PublishMarker = ".modular-publish.json"
	// PublishNotice tells readers the directory is generated.
	PublishNotice = "DO-NOT-EDIT.md"

	// Publish states of a module, as shown by modular:list.
	PublishLocal     = "local"     // in the project, nothing is copied
	PublishNone      = "none"      // third-party without sources or views
	PublishPublished = "published" // copied from the linked version
	PublishStale     = "stale"     // differs from the linked version
	PublishMissing   = "missing"   // not copied yet
	PublishConflict  = "conflict"  // the directory holds something else
)

// Dependency is the Go module a third-party module is linked from.
type Dependency struct {
	Path    string `json:"module"`
	Version string `json:"version"`
	Sum     string `json:"sum"`
}

type publishMarker struct {
	Dependency
	Files []string `json:"files"`
}

// DependencyOf returns the Go module of a third-party module. ok is false for a
// module of the application and for one that sets its own Path().
func DependencyOf(module modular.Module, info *debug.BuildInfo) (Dependency, bool) {
	if _, located := module.(modular.Located); located || info == nil {
		return Dependency{}, false
	}

	pkg := packagePath(module)
	if pkg == "" || pkg == "main" || within(pkg, info.Main.Path) {
		return Dependency{}, false
	}

	// Without a matching dependency (workspace builds) the package stands in for it.
	dependency := Dependency{Path: pkg}
	found := ""
	for _, dep := range info.Deps {
		if dep == nil || !within(pkg, dep.Path) || len(dep.Path) <= len(found) {
			continue
		}

		found = dep.Path
		linked := dep
		if dep.Replace != nil {
			linked = dep.Replace
		}
		dependency = Dependency{Path: dep.Path, Version: linked.Version, Sum: linked.Sum}
	}

	return dependency, true
}

// PublishState reports whether dir holds the copy of module published from dependency.
func PublishState(module modular.Module, dependency Dependency, dir string) string {
	files, err := publishFiles(module, dependency)
	if err != nil || len(files) == 0 {
		return PublishNone
	}

	marker, owned := readPublishMarker(dir)
	if !owned || marker.Path != dependency.Path {
		if emptyDir(dir) {
			return PublishMissing
		}

		return PublishConflict
	}
	if marker.Dependency != dependency || !slices.Equal(marker.Files, slices.Sorted(maps.Keys(files))) {
		return PublishStale
	}
	// The version alone misses a replaced module and edited copies: compare the files too.
	for name, content := range files {
		current, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
		if err != nil || !bytes.Equal(current, content) {
			return PublishStale
		}
	}

	return PublishPublished
}

// Publish copies the module's sources, and its views unless the sources hold
// them, into dir. It reports whether anything was written or removed.
func Publish(module modular.Module, dependency Dependency, dir string) (bool, error) {
	files, err := publishFiles(module, dependency)
	if err != nil || len(files) == 0 {
		return false, err
	}

	marker, owned := readPublishMarker(dir)
	if owned && marker.Path != dependency.Path {
		owned = false
	}
	if !owned && !emptyDir(dir) {
		return false, errors.ModularPublishTargetNotOwned.Args(module.Name(), dir, dependency.Path)
	}

	names := slices.Sorted(maps.Keys(files))

	changed := marker.Dependency != dependency || !slices.Equal(marker.Files, names)
	for _, name := range names {
		target := filepath.Join(dir, filepath.FromSlash(name))
		if current, err := os.ReadFile(target); err == nil && bytes.Equal(current, files[name]) {
			continue
		}

		changed = true
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return false, err
		}
		if err := os.WriteFile(target, files[name], 0o644); err != nil {
			return false, err
		}
	}

	// Remove what an earlier publish wrote and the module no longer ships.
	for _, name := range marker.Files {
		if _, ok := files[name]; ok {
			continue
		}
		if clean := stdpath.Clean(name); !fs.ValidPath(clean) || clean == "." {
			continue
		}

		target := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			return false, err
		}
		removeEmptyDirs(filepath.Dir(target), dir)
	}

	if !changed {
		return false, nil
	}

	content, err := json.MarshalIndent(publishMarker{Dependency: dependency, Files: names}, "", "  ")
	if err != nil {
		return false, err
	}

	return true, os.WriteFile(filepath.Join(dir, PublishMarker), append(content, '\n'), 0o644)
}

// publishFiles returns the files to write by slash path, empty when the module ships none.
func publishFiles(module modular.Module, dependency Dependency) (map[string][]byte, error) {
	files := make(map[string][]byte)

	if p, ok := module.(modular.ProvidesSources); ok {
		if err := readFiles(p.Sources(), "", files); err != nil {
			return nil, err
		}
	}

	if p, ok := module.(modular.ProvidesViews); ok && !hasPrefix(files, viewsDirectory+"/") {
		if err := readFiles(p.Views(), viewsDirectory, files); err != nil {
			return nil, err
		}
	}

	if len(files) == 0 {
		return files, nil
	}
	delete(files, PublishMarker)
	files[PublishNotice] = []byte(publishNotice(module.Name(), dependency))

	return files, nil
}

// viewsDirectory is where a module's views live inside its directory.
const viewsDirectory = "resources/views"

func readFiles(fsys fs.FS, prefix string, files map[string][]byte) error {
	if fsys == nil {
		return nil
	}

	return fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || isGoSource(name) {
			return err
		}

		content, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}
		files[stdpath.Join(prefix, name)] = content

		return nil
	})
}

// isGoSource reports files that would be compiled into the application, or cut the
// directory out of it: they are never published.
func isGoSource(name string) bool {
	base := stdpath.Base(name)

	return strings.HasSuffix(base, ".go") || base == "go.mod" || base == "go.sum"
}

func readPublishMarker(dir string) (publishMarker, bool) {
	var marker publishMarker
	content, err := os.ReadFile(filepath.Join(dir, PublishMarker))
	if err != nil || json.Unmarshal(content, &marker) != nil {
		return publishMarker{}, false
	}

	return marker, true
}

func publishNotice(name string, dependency Dependency) string {
	source := dependency.Path
	if dependency.Version != "" {
		source += " " + dependency.Version
	}

	return "# Do not edit\n\n" +
		"This directory is generated by `artisan modular:publish` from the module `" + name + "` (" + source + ").\n" +
		"Changes are overwritten by the next publish; run `go generate` after upgrading the module.\n" +
		"The files are here for tools that read sources (Vite, Tailwind); views are rendered from the copy embedded in the module.\n"
}

// within reports whether pkg is the Go module root or a package inside it.
func within(pkg, root string) bool {
	return root != "" && (pkg == root || strings.HasPrefix(pkg, root+"/"))
}

func hasPrefix(files map[string][]byte, prefix string) bool {
	for name := range files {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	return false
}

// emptyDir reports whether dir is missing or has no entries.
func emptyDir(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return os.IsNotExist(err)
	}

	return len(entries) == 0
}

// removeEmptyDirs removes dir and its empty parents, stopping at root.
func removeEmptyDirs(dir, root string) {
	for dir != root && strings.HasPrefix(dir, root) {
		if os.Remove(dir) != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}
