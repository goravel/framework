package modify

import (
	"slices"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"

	"github.com/goravel/framework/packages/match"
)

func ExprExists(x []dst.Expr, y dst.Expr) bool {
	return ExprIndex(x, y) >= 0
}

func ExprIndex(x []dst.Expr, y dst.Expr) int {
	return slices.IndexFunc(x, func(expr dst.Expr) bool {
		return match.EqualNode(y).MatchNode(expr)
	})
}

func IsUsingImport(df *dst.File, path string, name ...string) bool {
	if len(name) == 0 {
		split := strings.Split(path, "/")
		name = append(name, split[len(split)-1])
	}

	var used bool
	dst.Inspect(df, func(n dst.Node) bool {
		sel, ok := n.(*dst.SelectorExpr)
		if ok && isTopName(sel.X, name[0]) {
			used = true

			return false
		}
		return true
	})

	return used
}

func KeyExists(kvs []dst.Expr, key dst.Expr) bool {
	return KeyIndex(kvs, key) >= 0
}

func KeyIndex(kvs []dst.Expr, key dst.Expr) int {
	return slices.IndexFunc(kvs, func(expr dst.Expr) bool {
		if kv, ok := expr.(*dst.KeyValueExpr); ok {
			return match.EqualNode(key).MatchNode(kv.Key)
		}
		return false
	})
}

func MustParseExpr(x string) (node dst.Node) {
	src := "package p\nvar _ = " + x
	file, err := decorator.Parse(src)
	if err != nil {
		panic(err)
	}

	spec := file.Decls[0].(*dst.GenDecl).Specs[0].(*dst.ValueSpec)
	expr := spec.Values[0]

	// handle outer comments for expr
	expr.Decorations().Start = file.Decls[0].(*dst.GenDecl).Decorations().Start
	expr.Decorations().End = file.Decls[0].(*dst.GenDecl).Decorations().End

	return WrapNewline(expr)
}

func WrapNewline[T dst.Node](node T) T {
	dst.Inspect(node, func(n dst.Node) bool {
		switch v := n.(type) {
		case *dst.KeyValueExpr, *dst.UnaryExpr:
			v.Decorations().After = dst.NewLine
			v.Decorations().Before = dst.NewLine
		case *dst.FuncType:
			v.Results.Decorations().After = dst.NewLine
			v.Results.Decorations().Before = dst.NewLine
		}

		return true
	})

	return node
}

func isThirdParty(importPath string) bool {
	// Third party package import path usually contains "." (".com", ".org", ...)
	// This logic is taken from golang.org/x/tools/imports package.
	return strings.Contains(importPath, ".")
}

// isTopName returns true if n is a top-level unresolved identifier with the given name.
func isTopName(n dst.Expr, name string) bool {
	id, ok := n.(*dst.Ident)
	return ok && id.Name == name && id.Obj == nil
}
