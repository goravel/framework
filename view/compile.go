package view

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	stdpath "path"
	"path/filepath"
	"regexp"
	"slices"
	"sync"

	contractslog "github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/paths"
	"github.com/goravel/framework/support"
)

// Delims are custom template action delimiters.
type Delims struct {
	Left  string
	Right string
}

// CompileOptions configures Compile.
type CompileOptions struct {
	// Delims replaces the default "{{" and "}}" action delimiters.
	Delims *Delims
	// FuncMap is installed on the template set before any file is parsed.
	FuncMap template.FuncMap
	// Paths are package view directories in precedence order, typically
	// View.RegisteredViews(). A path that is missing or not a directory is skipped.
	Paths []string
	// FS are package view filesystems in precedence order, typically
	// View.RegisteredViewFS(). A nil or unreadable filesystem is skipped with a warning.
	FS []fs.FS
	// Extensions limits parsing to files with one of these extensions, dot
	// included. Empty parses every file.
	Extensions []string
	// Log receives warnings about name collisions and skipped filesystems. nil
	// silences them.
	Log contractslog.Log
}

// Compile parses the application's resources/views directory and every package
// source in options into one template set, in precedence order: application
// views first, then options.Paths, then options.FS, each in the order given.
//
// Every file is parsed under its slash-separated path relative to its source
// root, and additionally under the name of its {{ define }} block when it has
// one. A name already taken by a higher-precedence source causes the whole file
// to be skipped: silently when the application owns the name, with a warning
// when another package does. Route drivers wrap the returned set in their own
// renderer type.
//
// Compile returns (nil, nil) when no source contributed a file.
func Compile(options CompileOptions) (*template.Template, error) {
	instance := template.New("")
	leftDelim := "{{"
	if options.Delims != nil {
		instance.Delims(options.Delims.Left, options.Delims.Right)
		leftDelim = options.Delims.Left
	}
	if options.FuncMap != nil {
		instance.Funcs(options.FuncMap)
	}

	defines := &viewDefines{app: make(map[string]string), pkg: make(map[string]string), log: options.Log}
	loaded := false

	for _, source := range viewSources(options) {
		contributed, err := loadSource(instance, source, leftDelim, options.Extensions, defines)
		if err != nil {
			return nil, err
		}
		loaded = loaded || contributed
	}

	if !loaded {
		return nil, nil
	}

	return instance, nil
}

var (
	defineRe        = regexp.MustCompile(`\{\{\s*define\s+"([^"]+)"`)
	customDefineRes sync.Map
)

// viewTier is the precedence class of a view source. Lower-numbered tiers take
// precedence, in the order tierApp, tierDir, tierFS.
type viewTier int

const (
	tierApp viewTier = iota
	tierDir
	tierFS
)

// viewSource is a filesystem that contributes templates at a given precedence tier.
type viewSource struct {
	fsys fs.FS
	tier viewTier
	// label identifies the source in warnings and errors: the directory path for
	// tierApp and tierDir, "fs[i]" (i being the registration index) for tierFS.
	label string
}

// pathOf renders name, a slash-separated path inside the source, for display.
func (s viewSource) pathOf(name string) string {
	if s.tier == tierFS {
		return stdpath.Join(s.label, name)
	}

	return filepath.Join(s.label, filepath.FromSlash(name))
}

// viewDefines tracks the template name each precedence tier has already claimed,
// so the first source to define a name wins.
type viewDefines struct {
	app map[string]string
	pkg map[string]string
	log contractslog.Log
}

// claim records templateNames for source and reports whether source may contribute
// the file. A name already claimed by the application is dropped silently; one
// already claimed by an earlier package source is dropped with a warning. Nothing
// is recorded unless every name is free, so a skipped file never blocks a later one.
func (d *viewDefines) claim(source viewSource, name string, templateNames ...string) bool {
	fullPath := source.pathOf(name)

	if source.tier == tierApp {
		for _, templateName := range templateNames {
			d.app[templateName] = fullPath
		}

		return true
	}
	for _, templateName := range templateNames {
		if _, ok := d.app[templateName]; ok {
			return false
		}
	}
	for _, templateName := range templateNames {
		if prevFile, ok := d.pkg[templateName]; ok {
			if d.log != nil {
				d.log.Warningf("view collision: %q defined in %q and %q, using first", templateName, prevFile, fullPath)
			}

			return false
		}
	}

	for _, templateName := range templateNames {
		d.pkg[templateName] = fullPath
	}

	return true
}

// viewSources returns every existing template source in precedence order. A
// path that is missing or is not a directory is skipped: walking it through
// os.DirFS would fail the whole compile. Filesystem labels keep the registration
// index, so a skipped filesystem still consumes its slot and warnings point at
// the position the package registered.
func viewSources(options CompileOptions) []viewSource {
	var sources []viewSource

	if dir := paths.Abs(support.Config.Paths.Resources, "views"); isDir(dir) {
		sources = append(sources, viewSource{fsys: os.DirFS(dir), tier: tierApp, label: dir})
	}

	for _, dir := range options.Paths {
		if isDir(dir) {
			sources = append(sources, viewSource{fsys: os.DirFS(dir), tier: tierDir, label: dir})
		}
	}

	for i, fsys := range options.FS {
		if fsys == nil {
			if options.Log != nil {
				options.Log.Warningf("view source fs[%d] is nil, skipping", i)
			}
			continue
		}
		if _, err := fs.Stat(fsys, "."); err != nil {
			if options.Log != nil {
				options.Log.Warningf("view source fs[%d] is unreadable, skipping: %v", i, err)
			}
			continue
		}
		sources = append(sources, viewSource{fsys: fsys, tier: tierFS, label: fmt.Sprintf("fs[%d]", i)})
	}

	return sources
}

func isDir(dir string) bool {
	info, err := os.Stat(dir)

	return err == nil && info.IsDir()
}

// loadSource walks source and parses every file it contributes into instance,
// reporting whether it contributed any. Each file is read once and parsed
// immediately, so only one file's content is held at a time. A file is parsed
// under its slash-separated path relative to the source root, so files sharing a
// base name in different directories stay distinct, and is claimed under that
// path as well as its define name, so a define-less file cannot override a
// template of the same path from a higher-precedence source.
func loadSource(instance *template.Template, source viewSource, leftDelim string, extensions []string, defines *viewDefines) (bool, error) {
	contributed := false

	err := fs.WalkDir(source.fsys, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.ViewSourceWalkFailed.Args(source.pathOf(name), err)
		}
		if d.IsDir() {
			return nil
		}
		if len(extensions) > 0 && !slices.Contains(extensions, stdpath.Ext(name)) {
			return nil
		}

		content, err := fs.ReadFile(source.fsys, name)
		if err != nil {
			return errors.ViewReadFailed.Args(source.pathOf(name), err)
		}
		text := string(content)

		// The define name is checked before the path so a collision is reported by
		// the name templates are rendered with.
		var templateNames []string
		if defineName := extractDefineName(text, leftDelim); defineName != "" && defineName != name {
			templateNames = append(templateNames, defineName)
		}
		templateNames = append(templateNames, name)
		if !defines.claim(source, name, templateNames...) {
			return nil
		}

		// The content is parsed directly instead of through ParseFS, which names
		// each file after its base and, because it resolves names through fs.Glob,
		// mangles or fails on any name containing "[", "*" or "?".
		if _, err := instance.New(name).Parse(text); err != nil {
			return errors.ViewParseFailed.Args(source.pathOf(name), err)
		}
		contributed = true

		return nil
	})
	if err != nil {
		return false, err
	}

	return contributed, nil
}

// extractDefineName returns the name declared by the first {{ define }} block in
// content, or "" when there is none. Regexps for custom left delimiters are cached.
func extractDefineName(content string, leftDelim string) string {
	re := defineRe
	if leftDelim != "" && leftDelim != "{{" {
		if cached, ok := customDefineRes.Load(leftDelim); ok {
			re = cached.(*regexp.Regexp)
		} else {
			re = regexp.MustCompile(regexp.QuoteMeta(leftDelim) + `\s*define\s+"([^"]+)"`)
			customDefineRes.Store(leftDelim, re)
		}
	}

	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}
