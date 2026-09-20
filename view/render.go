package view

import (
	"html/template"
	"io/fs"
	"os"
	stdpath "path"
	"strings"
	"sync/atomic"
	"text/template/parse"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/paths"
	"github.com/goravel/framework/support"
)

// bodyHolder is the name a file's text is parsed under while it is taken apart. Parsing it
// under the file's own path would merge the text around the define blocks with a define block
// carrying that same path, which is the convention these views follow.
const bodyHolder = "\x00body"

// compiled is the result of parsing every view source, including a failure, so a broken
// view is not parsed again on every render.
type compiled struct {
	tmpl *template.Template
	err  error
	// sizes is keyed by the names a view can be rendered by and holds the last output size of
	// each, used to size the next buffer. Those names are not every name html/template ends up
	// associating with the set: a file that only holds define blocks also leaves an empty
	// template behind, and every set carries an unnamed root. Rendering either returns an empty
	// string, which must not pass for a view.
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

// renderable reports whether the view resolves to a template that can be rendered, which is
// the question First has to ask: a file can exist without being addressable under its own
// name, for example when it holds nothing but define blocks. When the sources cannot be
// parsed it falls back to Exists, so Render reports the parse error rather than First
// reporting that none of the views exist.
func (r *View) renderable(view string) bool {
	c := r.compile()
	if c.err != nil {
		return r.Exists(view)
	}

	_, ok := c.sizes[view]

	return ok
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
	var names []string
	c.tmpl, names, c.err = r.parse()
	for _, name := range names {
		c.sizes[name] = &atomic.Int64{}
	}
	r.compiled.Store(c)

	return c
}

func (r *View) resetCompiled() {
	r.compileMu.Lock()
	defer r.compileMu.Unlock()

	r.compiled.Store(nil)
}

// sources returns every place templates are loaded from, in precedence order: the
// application's resources/views, then directories registered via LoadViewsFrom, then
// filesystems registered via LoadViewsFromFS. The first source to contribute a name owns it,
// so a package can never replace a view the application, or a package registered before it,
// already provides.
func (r *View) sources() []fs.FS {
	var sources []fs.FS
	if dir := paths.Abs(support.Config.Paths.Resources, "views"); isDir(dir) {
		sources = append(sources, os.DirFS(dir))
	}
	for _, dir := range r.RegisteredViews() {
		if isDir(dir) {
			sources = append(sources, os.DirFS(dir))
		}
	}

	return append(sources, r.RegisteredViewFS()...)
}

// parse loads every source into one template set and returns the names the set can be
// rendered by. A file contributes each of its define blocks, and, when it has text of its own
// around them, its path within the source plus, for a nested file, its base name, which is
// what html/template.ParseFS and the route drivers address it by. Names are owned per source:
// views in one directory override each other the way html/template behaves on its own, while
// a later source is only used for the names no earlier source contributed. It returns a nil
// template when no source contributes one.
func (r *View) parse() (*template.Template, []string, error) {
	tmpl := template.New("")
	owned := make(map[string]bool)
	var names []string
	loaded := false

	for _, source := range r.sources() {
		contributed := make(map[string]bool)

		// claim moves one template into the set unless an earlier source owns the name. An alias
		// only ever fills a free name: it is a second way to address a file, never a view in its
		// own right, so it must not displace one.
		claim := func(name string, tree *parse.Tree, alias bool) error {
			if owned[name] || (alias && contributed[name]) {
				return nil
			}

			// AddParseTree keeps an existing body rather than letting an empty one replace it, so
			// a layout's {{ block }} placeholder cannot wipe the page that fills it, whichever of
			// the two is walked first.
			tree.Name = name
			if _, err := tmpl.AddParseTree(name, tree); err != nil {
				return err
			}
			if !contributed[name] {
				contributed[name] = true
				names = append(names, name)
			}

			return nil
		}

		err := fs.WalkDir(source, ".", func(name string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}

			content, err := fs.ReadFile(source, name)
			if err != nil {
				return err
			}
			text := string(content)

			// The file is parsed on its own and its templates are moved into the set one at a
			// time. Parsing it straight into the set would apply all of its define blocks at once,
			// letting a package take a name it does not own because another name in the same file
			// happened to be free.
			file, err := template.New(bodyHolder).Parse(text)
			if err != nil {
				return parseError(name, text, err)
			}
			loaded = true

			for _, sub := range file.Templates() {
				if sub.Name() == bodyHolder || sub.Tree == nil {
					continue
				}
				// Report positions against the file rather than the name it was parsed under.
				sub.Tree.ParseName = name
				if err := claim(sub.Name(), sub.Tree, false); err != nil {
					return err
				}
			}

			body := file.Lookup(bodyHolder)
			if body == nil || body.Tree == nil || parse.IsEmptyTree(body.Tree.Root) {
				return nil
			}
			body.Tree.ParseName = name

			if err := claim(name, body.Tree, false); err != nil {
				return err
			}
			if base := stdpath.Base(name); base != name {
				return claim(base, body.Tree.Copy(), true)
			}

			return nil
		})
		if err != nil {
			return nil, nil, err
		}

		for name := range contributed {
			owned[name] = true
		}
	}

	if !loaded {
		return nil, nil, nil
	}

	return tmpl, names, nil
}

// parseError reports a parse failure under the file's path, which is not the name the text was
// parsed under and so not the name the failure would otherwise carry.
func parseError(name, text string, err error) error {
	if _, named := template.New(name).Parse(text); named != nil {
		return named
	}

	return err
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}
