package view

import (
	"bytes"
	"html/template"
	"io/fs"
	"os"
	stdpath "path"
	"regexp"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/paths"
	"github.com/goravel/framework/support"
)

var defineRe = regexp.MustCompile(`\{\{\s*define\s+"([^"]+)"`)

// render executes the named template with the given values.
func (r *View) render(name string, values map[string]any) (string, error) {
	tmpl, err := r.templates()
	if err != nil {
		return "", err
	}
	if tmpl == nil || tmpl.Lookup(name) == nil {
		return "", errors.ViewTemplateNotExist.Args(name)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, values); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// templates returns the compiled template set, parsing every view source on first use.
// The set is reset whenever a new source is registered.
func (r *View) templates() (*template.Template, error) {
	r.tmplMu.Lock()
	defer r.tmplMu.Unlock()

	if r.tmpl != nil {
		return r.tmpl, nil
	}

	tmpl, err := r.parse()
	if err != nil {
		return nil, err
	}
	r.tmpl = tmpl

	return tmpl, nil
}

func (r *View) resetTemplate() {
	r.tmplMu.Lock()
	defer r.tmplMu.Unlock()

	r.tmpl = nil
}

// parse loads every source into one template set in precedence order: the application's
// resources/views, then directories registered via LoadViewsFrom, then filesystems registered
// via LoadViewsFromFS. The first source to define a template name wins. It returns nil when
// no source contributes a template.
func (r *View) parse() (*template.Template, error) {
	var sources []fs.FS
	if dir := paths.Abs(support.Config.Paths.Resources, "views"); isDir(dir) {
		sources = append(sources, os.DirFS(dir))
	}
	for _, dir := range r.RegisteredViews() {
		if isDir(dir) {
			sources = append(sources, os.DirFS(dir))
		}
	}
	sources = append(sources, r.RegisteredViewFS()...)

	tmpl := template.New("")
	defined := make(map[string]bool)
	loaded := false

	for _, source := range sources {
		err := fs.WalkDir(source, ".", func(name string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}

			content, err := fs.ReadFile(source, name)
			if err != nil {
				return err
			}
			text := string(content)

			// A file without a define block is addressed by its base name, the name it is parsed as.
			templateName := stdpath.Base(name)
			if matches := defineRe.FindStringSubmatch(text); len(matches) > 1 {
				templateName = matches[1]
			}
			if defined[templateName] {
				return nil
			}
			defined[templateName] = true

			if _, err := tmpl.New(stdpath.Base(name)).Parse(text); err != nil {
				return err
			}
			loaded = true

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	if !loaded {
		return nil, nil
	}

	return tmpl, nil
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}
