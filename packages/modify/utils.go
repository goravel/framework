package modify

import (
	"go/parser"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"

	"github.com/goravel/framework/packages/match"
)

func ExprExists(x dst.Expr, y []dst.Expr) (bool, int) {
	for i := range y {
		if match.EqualNode(x).MatchNode(y[i]) {
			return true, i
		}
	}

	return false, -1
}

func KeyExists(key dst.Expr, kvs []dst.Expr) (bool, int) {
	for i := range kvs {
		if kv, ok := kvs[i].(*dst.KeyValueExpr); ok {
			if match.EqualNode(key).MatchNode(kv.Key) {
				return true, i
			}
		}
	}

	return false, -1
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
