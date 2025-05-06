package modify

import (
	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"

	"github.com/goravel/framework/contracts/packages/match"
)

type Action func(cursor *dstutil.Cursor)

type File interface {
	Apply() error
}

type GoFile interface {
	File
	Find(matchers []match.GoNode) GoNode
}

type GoNode interface {
	Apply(node dst.Node) error
	Modify(actions ...Action) GoFile
}
