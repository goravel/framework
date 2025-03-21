package packages

import (
	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
)

type FileModifier interface {
	Apply() error
}

type Setup interface {
	Install() error
	Uninstall() error
}

type GoNodeMatcher interface {
	MatchNode(node dst.Node) bool
	MatchCursor(cursor *dstutil.Cursor) bool
}

type GoNodeModifier interface {
	Apply(node dst.Node) error
}
