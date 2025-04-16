package modify

import (
	"bytes"
	"fmt"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"

	"github.com/goravel/framework/contracts/packages/match"
	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/file"
)

type goFile struct {
	file      string
	modifiers []modify.GoNode
}

func GoFile(file string) modify.GoFile {
	return &goFile{file: file}
}

func (r goFile) Apply() error {
	source, err := file.GetContent(r.file)
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

	return file.PutContent(r.file, buf.String())
}

func (r goFile) Find(matchers ...match.GoNode) modify.GoNode {
	modifier := &GoNode{
		matchers: matchers,
		file:     &r,
	}
	r.modifiers = append(r.modifiers, modifier)
	return modifier
}

type GoNode struct {
	actions  []modify.Action
	file     *goFile
	matchers []match.GoNode
}

func (r GoNode) Apply(node dst.Node) (err error) {
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

func (r *GoNode) Modify(actions ...modify.Action) modify.GoFile {
	r.actions = actions

	return r.file
}
