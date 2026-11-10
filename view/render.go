package view

import (
	"html/template"
	"strings"
	"sync/atomic"
	"text/template/parse"

	"github.com/goravel/framework/errors"
)

// maxSizeHint caps the buffer reserved for a view, so a one-off spike in output size cannot
// size every later render of the same view.
const maxSizeHint = 1 << 20

// compiled is the result of compiling every view source, including a failure, so a broken
// view is not compiled again on every render.
type compiled struct {
	template *template.Template
	err      error
	// names is the set of renderable views. Not every name html/template ends up with is one: a
	// file that only holds define blocks leaves an empty template behind, every set carries an
	// unnamed root, and an empty {{ block }} placeholder stays empty until a body for it is
	// parsed.
	names map[string]struct{}
	// sizes holds the output size of the last render under each name, used to size the next
	// buffer.
	sizes map[string]*atomic.Int64
}

// render executes the named view with the given values.
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
	if n := size.Load(); n > 0 {
		buf.Grow(int(min(n, maxSizeHint)))
	}
	if err := c.template.ExecuteTemplate(&buf, name, values); err != nil {
		return "", err
	}
	size.Store(int64(buf.Len()))

	return buf.String(), nil
}

// renderable reports whether the name resolves to a view that can be rendered, which is the
// question First has to ask: a file can exist without being addressable under its own name,
// for example when it holds nothing but define blocks. When the sources cannot be compiled
// there is no telling which names resolve, so every candidate counts as renderable and Render
// reports the compile error rather than First reporting that none of the views exist.
func (r *View) renderable(name string) bool {
	c := r.compile()
	if c.err != nil {
		return true
	}

	_, ok := c.names[name]

	return ok
}

// compile returns the compiled view set, compiling it on first use. The result is kept until a
// view source or a template function is registered. Compiling goes through Compile, so Make
// resolves views exactly the way the route drivers do.
func (r *View) compile() *compiled {
	if c := r.compiled.Load(); c != nil {
		return c
	}

	r.compileMu.Lock()
	defer r.compileMu.Unlock()

	if c := r.compiled.Load(); c != nil {
		return c
	}

	tmpl, err := Compile(CompileOptions{
		FuncMap: r.Funcs(),
		Paths:   r.RegisteredViews(),
		FS:      r.RegisteredViewFS(),
	})
	names, sizes := renderableNames(tmpl)
	c := &compiled{template: tmpl, err: err, names: names, sizes: sizes}
	r.compiled.Store(c)

	return c
}

// resetCompiled drops the compiled set so the next render picks up a new registration.
// compile() takes compileMu and then reads the registry under mu.RLock, so a caller must not
// hold mu when calling this.
func (r *View) resetCompiled() {
	r.compileMu.Lock()
	defer r.compileMu.Unlock()

	r.compiled.Store(nil)
}

// renderableNames returns every template name in the set that holds something to render, and
// the buffer hint kept for each.
func renderableNames(tmpl *template.Template) (map[string]struct{}, map[string]*atomic.Int64) {
	names := make(map[string]struct{})
	sizes := make(map[string]*atomic.Int64)
	if tmpl == nil {
		return names, sizes
	}

	for _, t := range tmpl.Templates() {
		if t.Name() == "" || t.Tree == nil || parse.IsEmptyTree(t.Tree.Root) {
			continue
		}

		names[t.Name()] = struct{}{}
		sizes[t.Name()] = &atomic.Int64{}
	}

	return names, sizes
}
