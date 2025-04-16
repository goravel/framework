package modify

import (
	"go/parser"
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
	exp, err := parser.ParseExpr(x)
	if err == nil {
		node, err = decorator.Decorate(nil, exp)
	}
	if err != nil {
		panic(err)
	}

	return WrapNewline(node)
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
