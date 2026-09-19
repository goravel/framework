package view

import (
	"html/template"
	"io/fs"
	"os"
	stdpath "path"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/paths"
	"github.com/goravel/framework/support"
)

var defineRe = regexp.MustCompile(`\{\{\s*define\s+"([^"]+)"`)

// compiled is the result of parsing every view source, including a failure, so a broken
// view is not parsed again on every render.
type compiled struct {
	tmpl *template.Template
	err  error
	// sizes holds the last output size of each template, used to size the next buffer.
	sizes map[string]*atomic.Int64
}

// render executes the named template with the given values.
func (r *View) render(name string, values map[string]any) (string, error) {
	c := r.compile()
	if c.err != nil {
		return "", c.err
	}

	size, ok := c.sizes[name]
	if !ok {
		return "", errors.ViewTemplateNotExist.Args(name)
	}

	var buf strings.Builder
	buf.Grow(int(size.Load()))
	if err := c.tmpl.ExecuteTemplate(&buf, name, values); err != nil {
		return "", err
	}
	size.Store(int64(buf.Len()))

	return buf.String(), nil
}

// compile returns the parsed view sources, parsing them on first use. The result is kept
// until a new source is registered.
func (r *View) compile() *compiled {
	if c := r.compiled.Load(); c != nil {
		return c
	}

	r.compileMu.Lock()
	defer r.compileMu.Unlock()

	if c := r.compiled.Load(); c != nil {
		return c
	}

	c := &compiled{sizes: make(map[string]*atomic.Int64)}
	c.tmpl, c.err = r.parse()
	if c.tmpl != nil {
		for _, t := range c.tmpl.Templates() {
			c.sizes[t.Name()] = &atomic.Int64{}
		}
	}
	r.compiled.Store(c)

	return c
}

func (r *View) resetCompiled() {
	r.compileMu.Lock()
	defer r.compileMu.Unlock()

	r.compiled.Store(nil)
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
