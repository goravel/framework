package packages

import (
	"go/parser"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func ExprExists(x dst.Expr, y []dst.Expr) bool {
	for i := range y {
		if MatchEqualNode(x).MatchNode(y[i]) {
			return true
		}
	}

	return false
}

func KeyExists(key dst.Expr, kvs []dst.Expr) bool {
	for i := range kvs {
		if kv, ok := kvs[i].(*dst.KeyValueExpr); ok {
			if MatchEqualNode(key).MatchNode(kv.Key) {
				return true
			}
		}
	}

	return false
}

func MustParseStatement(x string) (node dst.Node) {
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
