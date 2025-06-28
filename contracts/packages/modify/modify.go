package modify

import (
	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"

	"github.com/goravel/framework/contracts/packages/match"
)

type Action func(cursor *dstutil.Cursor)

type Apply interface {
	Apply() error
}

type File interface {
	Overwrite(content string, forces ...bool) Apply
	Remove() Apply
}

type GoFile interface {
	Apply
	Find(matchers []match.GoNode) GoNode
}

type GoNode interface {
	Apply(node dst.Node) error
	Modify(actions ...Action) GoFile
}
