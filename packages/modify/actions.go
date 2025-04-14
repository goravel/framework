package modify

import (
	"fmt"
	"go/token"
	"strconv"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"

	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/support/color"
)

// AddConfig adds a configuration key with the given expression to the config file.
func AddConfig(name, expression string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		var value *dst.CompositeLit
		switch node := cursor.Node().(type) {
		case *dst.KeyValueExpr:
			value = node.Value.(*dst.CompositeLit)
		case *dst.CallExpr:
			value = node.Args[1].(*dst.CompositeLit)
		}
		key := WrapNewline(&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(name)})
		if exists, _ := KeyExists(key, value.Elts); exists {
			color.Warningln(fmt.Sprintf("key [%s] already exists,using ReplaceConfig instead if you want to update it", name))

			return
		}

		// add config
		value.Elts = append(value.Elts, WrapNewline(&dst.KeyValueExpr{
			Key:   key,
			Value: WrapNewline(MustParseExpr(expression)).(dst.Expr),
		}))
	}
}

// AddImport adds an import statement to the file.
func AddImport(path string, name ...string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		node := cursor.Node().(*dst.GenDecl)

		// import spec
		im := &dst.ImportSpec{
			Path: &dst.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote(path),
			},
		}
		if len(name) > 0 {
			im.Name = &dst.Ident{
				Name: name[0],
			}
		}

		// Insert third-party imports at the top and others at the bottom.
		// When formatting the source code, this helps group and sort imports
		// into stdlib, third-party, and local packages.
		if isThirdParty(path) {
			node.Specs = append([]dst.Spec{WrapNewline(im)}, node.Specs...)

			return
		}
		node.Specs = append(node.Specs, WrapNewline(im))
	}
}

// AddProvider adds a provider to the service providers array.
func AddProvider(expression string, before ...string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		provider := MustParseExpr(expression).(dst.Expr)
		node := cursor.Node().(*dst.CompositeLit)
		if exist, _ := ExprExists(provider, node.Elts); !exist {
			if len(before) > 0 {
				beforeExpr := MustParseExpr(before[0]).(dst.Expr)

				// check if beforeExpr is existing and insert provider before it
				if has, i := ExprExists(beforeExpr, node.Elts); has {
					node.Elts = append(node.Elts[:i], append([]dst.Expr{provider}, node.Elts[i:]...)...)

					return
				}

				color.Warningln(fmt.Sprintf("provider [%s] not found, cannot insert before it", before[0]))
			}

			// insert provider at the end
			node.Elts = append(node.Elts, provider)

			return
		}
		color.Warningln(fmt.Sprintf("provider [%s] already exists", expression))
	}
}

// RemoveConfig removes a configuration key from the config file.
func RemoveConfig(name string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		var value *dst.CompositeLit
		switch node := cursor.Node().(type) {
		case *dst.KeyValueExpr:
			value = node.Value.(*dst.CompositeLit)
		case *dst.CallExpr:
			value = node.Args[1].(*dst.CompositeLit)
		}
		key := WrapNewline(&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(name)})

		// remove config
		if exist, i := KeyExists(key, value.Elts); exist {
			value.Elts = append(value.Elts[:i], value.Elts[i+1:]...)

			return
		}
	}
}

// RemoveImport removes an import statement from the file.
func RemoveImport(path string, name ...string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		node := cursor.Node().(*dst.GenDecl)
		for i, spec := range node.Specs {
			// remove import spec
			if match.ImportSpec(path, name...).MatchNode(spec) {
				node.Specs = append(node.Specs[:i], node.Specs[i+1:]...)

				return
			}
		}
	}
}

// RemoveProvider removes a provider from the service providers array.
func RemoveProvider(expression string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		provider := MustParseExpr(expression).(dst.Expr)
		node := cursor.Node().(*dst.CompositeLit)

		// remove provider
		if exist, i := ExprExists(provider, node.Elts); exist {
			node.Elts = append(node.Elts[:i], node.Elts[i+1:]...)
		}
	}
}

func ReplaceConfig(name, expression string) modify.Action {
	return func(cursor *dstutil.Cursor) {
		var value *dst.CompositeLit
		switch node := cursor.Node().(type) {
		case *dst.KeyValueExpr:
			value = node.Value.(*dst.CompositeLit)
		case *dst.CallExpr:
			value = node.Args[1].(*dst.CompositeLit)
		}
		key := WrapNewline(&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(name)})

		// replace config
		if exist, i := KeyExists(key, value.Elts); exist {
			value.Elts[i] = WrapNewline(&dst.KeyValueExpr{
				Key:   key,
				Value: WrapNewline(MustParseExpr(expression)).(dst.Expr),
			})

			return
		}

	}
}
