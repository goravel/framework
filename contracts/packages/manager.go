package packages

import (
	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
)

type FileModifier interface {
	Apply(dir string) error
}

type Manager interface {
	Install(dir string) error
	Uninstall(dir string) error
}

type GoNodeMatcher interface {
	MatchNode(node dst.Node) bool
	MatchCursor(cursor *dstutil.Cursor) bool
}

type GoNodeModifier interface {
	Apply(node dst.Node) error
}
