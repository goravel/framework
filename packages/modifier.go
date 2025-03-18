package packages

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"

	"github.com/goravel/framework/contracts/packages"
	"github.com/goravel/framework/support/file"
)

type ModifyGoFile struct {
	File      string
	Modifiers []packages.GoNodeModifier
}
type ModifyGoNode struct {
	Action   func(cursor *dstutil.Cursor)
	Matchers []packages.GoNodeMatcher
}

func (g ModifyGoFile) Apply(dir string) error {
	fp := filepath.Join(dir, g.File)
	source, err := file.GetContent(fp)
	if err != nil {
		return err
	}

	df, err := decorator.Parse(source)
	if err != nil {
		return err
	}

	for i := range g.Modifiers {
		if err = g.Modifiers[i].Apply(df); err != nil {
			return fmt.Errorf("error modifying file %s: %v", g.File, err)
		}
	}

	var buf bytes.Buffer
	err = decorator.Fprint(&buf, df)
	if err != nil {
		return err
	}

	return file.PutContent(fp, buf.String())
}

func (g ModifyGoNode) Apply(node dst.Node) (err error) {
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

		if g.Matchers[current].MatchCursor(cursor) {
			matchedNodes[cursor.Node()] = true
			current++
			if current == len(g.Matchers) {
				// apply the action after all matchers are matched
				g.Action(cursor)
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
		return fmt.Errorf("%d out of %d matchers did not match", len(g.Matchers)-current, len(g.Matchers))
	}

	return nil
}
