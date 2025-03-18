package packages

import (
	"reflect"
	"strconv"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"

	"github.com/goravel/framework/contracts/packages"
)

type (
	MatchGoNode struct {
		first, last bool
		match       func(node dst.Node) bool
	}
	MatchGoNodes []packages.GoNodeMatcher
)

func (gn MatchGoNode) MatchCursor(cursor *dstutil.Cursor) bool {
	if gn.first || gn.last {
		if gn.MatchNode(cursor.Node()) {
			pr := reflect.Indirect(reflect.ValueOf(cursor.Parent())).FieldByName(cursor.Name())
			if pr.Kind() == reflect.Slice || pr.Kind() == reflect.Array {
				if gn.first {
					return cursor.Index() == 0
				}

				if gn.last {
					return cursor.Index() == pr.Len()-1
				}
			}
		}

		return false
	}

	return gn.MatchNode(cursor.Node())
}

func (gn MatchGoNode) MatchNode(node dst.Node) bool {
	return gn.match(node)
}

func (gns MatchGoNodes) MatchNodes(nodes []dst.Node) bool {
	if len(gns) == 0 {
		return true
	}

	if len(nodes) != len(gns) {
		return false
	}

	for i := range nodes {
		if len(gns) > i {
			if !gns[i].MatchNode(nodes[i]) {
				return false
			}
		}
	}

	return true
}

func MatchAnyNode() packages.GoNodeMatcher {
	return &MatchGoNode{
		match: func(node dst.Node) bool {
			return true
		},
	}
}

func MatchAnyNodes() MatchGoNodes {
	return MatchGoNodes{}
}

func MatchArrayType(elt, l packages.GoNodeMatcher) packages.GoNodeMatcher {
	return MatchGoNode{
		match: func(n dst.Node) bool {
			if e, ok := n.(*dst.ArrayType); ok {
				return elt.MatchNode(e.Elt) && l.MatchNode(e.Len)
			}

			return false
		},
	}
}

func MatchBasicLit(value string) packages.GoNodeMatcher {
	return MatchGoNode{
		match: func(n dst.Node) bool {
			if e, ok := n.(*dst.BasicLit); ok {
				return e.Value == value
			}

			return false
		},
	}
}

func MatchCallExpr(fun packages.GoNodeMatcher, args MatchGoNodes) packages.GoNodeMatcher {
	return MatchGoNode{
		match: func(n dst.Node) bool {
			if e, ok := n.(*dst.CallExpr); ok {
				var nodes = make([]dst.Node, len(e.Args))
				for i := range e.Args {
					nodes[i] = e.Args[i]
				}

				return fun.MatchNode(e.Fun) && args.MatchNodes(nodes)
			}

			return false
		},
	}
}

func MatchCompositeLit(t packages.GoNodeMatcher) packages.GoNodeMatcher {
	return MatchGoNode{
		match: func(n dst.Node) bool {
			if e, ok := n.(*dst.CompositeLit); ok {
				return t.MatchNode(e.Type)
			}

			return false
		},
	}
}

func MatchEqualNode(n dst.Node) packages.GoNodeMatcher {
	return MatchGoNode{
		match: func(node dst.Node) bool {
			return dstNodeEq(n, node)
		},
	}
}

func MatchEqualStatement(statement string) packages.GoNodeMatcher {
	return MatchGoNode{
		match: func(node dst.Node) bool {
			return dstNodeEq(MustParseStatement(statement), node)
		},
	}
}

func MatchFuncDecl(name packages.GoNodeMatcher) packages.GoNodeMatcher {
	return MatchGoNode{
		match: func(n dst.Node) bool {
			if e, ok := n.(*dst.FuncDecl); ok {
				return name.MatchNode(e.Name)
			}

			return false
		},
	}
}

func MatchIdent(name string) packages.GoNodeMatcher {
	return MatchGoNode{
		match: func(n dst.Node) bool {
			if ident, ok := n.(*dst.Ident); ok {
				return ident.Name == name
			}

			return false
		},
	}
}

func MatchImportSpec(path string, name ...string) packages.GoNodeMatcher {
	return MatchGoNode{
		match: func(n dst.Node) bool {
			if im, ok := n.(*dst.ImportSpec); ok {
				if im.Path.Value == strconv.Quote(path) {
					if len(name) > 0 {
						if im.Name != nil {
							return im.Name.Name == name[0]
						}
					}
					return true
				}
			}
			return false
		},
	}
}

func MatchKeyValueExpr(key, value packages.GoNodeMatcher) packages.GoNodeMatcher {
	return MatchGoNode{
		match: func(n dst.Node) bool {
			if e, ok := n.(*dst.KeyValueExpr); ok {
				return key.MatchNode(e.Key) && value.MatchNode(e.Value)
			}

			return false
		},
	}
}

func MatchLastOf(n packages.GoNodeMatcher) packages.GoNodeMatcher {
	return MatchGoNode{
		last:  true,
		match: n.MatchNode,
	}
}

func MatchSelectorExpr(x, sel packages.GoNodeMatcher) packages.GoNodeMatcher {
	return MatchGoNode{
		match: func(n dst.Node) bool {
			if e, ok := n.(*dst.SelectorExpr); ok {
				return x.MatchNode(e.X) && sel.MatchNode(e.Sel)
			}

			return false
		},
	}
}

func MatchTypeOf[T any](_ T) packages.GoNodeMatcher {
	return MatchGoNode{
		match: func(node dst.Node) bool {
			_, ok := node.(T)
			return ok
		},
	}
}

func dstNodeEq(x, y dst.Node) bool {
	switch x := x.(type) {
	case dst.Expr:
		y, ok := y.(dst.Expr)
		return ok && dstExprEq(x, y)
	case *dst.ImportSpec:
		y, ok := y.(*dst.ImportSpec)
		return ok && dstImportSpecEq(x, y)

	default:
		panic("unhandled node type, please add it to dstNodeEq")
	}
}

func dstExprEq(x, y dst.Expr) bool {
	if x == nil || y == nil {
		return x == y
	}

	dstutil.Unparen(x)

	switch x := x.(type) {
	case *dst.ArrayType:
		y, ok := y.(*dst.ArrayType)
		return ok && dstArrayTypeEq(x, y)
	case *dst.BasicLit:
		y, ok := y.(*dst.BasicLit)
		return ok && dstBasicLitEq(x, y)
	case *dst.CompositeLit:
		y, ok := y.(*dst.CompositeLit)
		return ok && dstCompositeLitEq(x, y)
	case *dst.Ident:
		y, ok := y.(*dst.Ident)
		return ok && dstIdentEq(x, y)
	case *dst.KeyValueExpr:
		y, ok := y.(*dst.KeyValueExpr)
		return ok && dstKeyValueExprEq(x, y)
	case *dst.MapType:
		y, ok := y.(*dst.MapType)
		return ok && dstMapTypeEq(x, y)
	case *dst.SelectorExpr:
		y, ok := y.(*dst.SelectorExpr)
		return ok && dstSelectorExprEq(x, y)
	case *dst.UnaryExpr:
		y, ok := y.(*dst.UnaryExpr)
		return ok && dstUnaryExprEq(x, y)

	default:
		panic("unhandled node type, please add it to dstExprEq")
	}
}

func dstArrayTypeEq(x, y *dst.ArrayType) bool {
	if x == nil || y == nil {
		return x == y
	}

	return dstExprEq(x.Elt, y.Elt) && dstExprEq(x.Len, y.Len)
}

func dstBasicLitEq(x, y *dst.BasicLit) bool {
	if x == nil || y == nil {
		return x == y
	}

	return x.Kind == y.Kind && x.Value == y.Value
}

func dstCompositeLitEq(x, y *dst.CompositeLit) bool {
	if x == nil || y == nil {
		return x == y
	}

	return dstExprEq(x.Type, y.Type) && dstExprSliceEq(x.Elts, y.Elts)
}

func dstExprSliceEq(xs, ys []dst.Expr) bool {
	if len(xs) != len(ys) {
		return false
	}

	for i := range xs {
		if !dstExprEq(xs[i], ys[i]) {
			return false
		}
	}

	return true
}

func dstIdentEq(x, y *dst.Ident) bool {
	if x == nil || y == nil {
		return x == y
	}

	return x.Name == y.Name
}

func dstImportSpecEq(x, y *dst.ImportSpec) bool {
	if x == nil || y == nil {
		return x == y
	}

	return x.Path.Value == y.Path.Value && dstIdentEq(x.Name, y.Name)
}

func dstKeyValueExprEq(x, y *dst.KeyValueExpr) bool {
	if x == nil || y == nil {
		return x == y
	}

	return dstExprEq(x.Key, y.Key) && dstExprEq(x.Value, y.Value)
}

func dstMapTypeEq(x, y *dst.MapType) bool {
	if x == nil || y == nil {
		return x == y
	}

	return dstExprEq(x.Key, y.Key) && dstExprEq(x.Value, y.Value)
}

func dstSelectorExprEq(x, y *dst.SelectorExpr) bool {
	if x == nil || y == nil {
		return x == y
	}

	return dstExprEq(x.X, y.X) && dstIdentEq(x.Sel, y.Sel)
}

func dstUnaryExprEq(x, y *dst.UnaryExpr) bool {
	if x == nil || y == nil {
		return x == y
	}

	return x.Op == y.Op && dstExprEq(x.X, y.X)
}
