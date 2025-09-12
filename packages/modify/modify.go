package modify

import (
	"bytes"
	"fmt"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/packages/match"
	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
	supportfile "github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path/internals"
	"github.com/goravel/framework/support/str"
)

func File(path string) modify.File {
	return &file{path: path}
}

func GoFile(file string) modify.GoFile {
	return &goFile{file: file}
}

func WhenFacade(facade string, applies ...modify.Apply) modify.Apply {
	return &whenFacadeModifier{
		facade:  facade,
		applies: applies,
	}
}

func WhenNoFacades(facades []string, applies ...modify.Apply) modify.Apply {
	return &whenNoFacadesModifier{
		facades: facades,
		applies: applies,
	}
}

func generateOptions(options []modify.Option) map[string]any {
	result := make(map[string]any)
	for _, option := range options {
		option(result)
	}
	return result
}

type file struct {
	path string
}

func (r *file) Overwrite(content string) modify.Apply {
	return &overwriteFile{
		path:    r.path,
		content: content,
	}
}

func (r *file) Remove() modify.Apply {
	return &removeFile{path: r.path}
}

type overwriteFile struct {
	path    string
	content string
}

func (r *overwriteFile) Apply(options ...modify.Option) error {
	generatedOptions := generateOptions(options)

	if supportfile.Exists(r.path) && !cast.ToBool(generatedOptions["force"]) {
		color.Warningln(errors.ConsoleFileAlreadyExists.Args(r.path))
		return nil
	}

	return supportfile.PutContent(r.path, r.content)
}

type removeFile struct {
	path string
}

func (r *removeFile) Apply(...modify.Option) error {
	return supportfile.Remove(r.path)
}

type goFile struct {
	file      string
	modifiers []modify.GoNode
}

func (r goFile) Apply(...modify.Option) error {
	source, err := supportfile.GetContent(r.file)
	if err != nil {
		return err
	}

	df, err := decorator.Parse(source)
	if err != nil {
		return err
	}

	for i := range r.modifiers {
		if err = r.modifiers[i].Apply(df); err != nil {
			return errors.PackageModifyGoFileFail.Args(r.file, err)
		}
	}

	var buf bytes.Buffer
	err = decorator.Fprint(&buf, df)
	if err != nil {
		return err
	}

	return supportfile.PutContent(r.file, buf.String())
}

func (r goFile) Find(matchers []match.GoNode) modify.GoNode {
	modifier := &goNode{
		matchers: matchers,
		goFile:   &r,
	}
	r.modifiers = append(r.modifiers, modifier)
	return modifier
}

type goNode struct {
	actions  []modify.Action
	goFile   *goFile
	matchers []match.GoNode
}

func (r goNode) Apply(node dst.Node) (err error) {
	var (
		current      int
		matched      bool
		matchedNodes = make(map[dst.Node]bool)
	)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	// match the node and apply the action
	dstutil.Apply(node, func(cursor *dstutil.Cursor) bool {
		// if already modified, skip the rest of the nodes
		if matched {
			return false
		}

		if r.matchers[current].MatchCursor(cursor) {
			matchedNodes[cursor.Node()] = true
			current++
			if current == len(r.matchers) {
				// apply the actions after all matchers are matched
				for _, action := range r.actions {
					action(cursor)
				}
				matched = true

				return false
			}
		}

		return true
	}, func(cursor *dstutil.Cursor) bool {
		if nd := cursor.Node(); nd != nil && matchedNodes[nd] {
			return false
		}

		return true
	})

	if !matched {
		count := len(r.matchers)
		return errors.PackageMatchGoNodeFail.Args(count-current, count)
	}

	return nil
}

func (r *goNode) Modify(actions ...modify.Action) modify.GoFile {
	r.actions = actions

	return r.goFile
}

type whenFacadeModifier struct {
	facade  string
	applies []modify.Apply
}

func (r whenFacadeModifier) Apply(options ...modify.Option) error {
	generatedOptions := generateOptions(options)

	if r.facade != generatedOptions["facade"] {
		return nil
	}

	for _, apply := range r.applies {
		if err := apply.Apply(options...); err != nil {
			return err
		}
	}

	return nil
}

type whenNoFacadesModifier struct {
	facades []string
	applies []modify.Apply
}

func (r whenNoFacadesModifier) Apply(options ...modify.Option) error {
	var exist bool
	generatedOptions := generateOptions(options)

	for _, facade := range r.facades {
		if facade == generatedOptions["facade"] {
			continue
		}

		if supportfile.Exists(facadeToFilepath(facade)) {
			exist = true
			break
		}
	}

	if !exist {
		for _, apply := range r.applies {
			if err := apply.Apply(options...); err != nil {
				return err
			}
		}
	}

	return nil
}

func facadeToFilepath(facade string) string {
	return internals.FacadesPath(str.Of(facade).Snake().String() + ".go")
}
