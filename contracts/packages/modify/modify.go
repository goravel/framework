package modify

import (
	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"

	"github.com/goravel/framework/contracts/packages/match"
)

type Action func(cursor *dstutil.Cursor)

type Option func(map[string]any)

type Apply interface {
	Apply(...Option) error
}

type Facade interface {
	File(path string) File
}

type File interface {
	Overwrite(content string) Apply
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
