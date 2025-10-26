package match

import (
	"go/token"
	"strconv"
	"strings"

	"github.com/dave/dst"

	"github.com/goravel/framework/contracts/packages/match"
)

func Config(key string) []match.GoNode {
	keys := strings.Split(key, ".")
	matchers := []match.GoNode{
		Func(Ident("init")),
		CallExpr(
			SelectorExpr(
				AnyOf(
					Ident("config"),
					CallExpr(
						SelectorExpr(
							Ident("facades"),
							Ident("Config"),
						),
						AnyNodes(),
					),
				),
				Ident("Add"),
			),
			GoNodes{
				BasicLit(strconv.Quote(keys[0])),
				AnyNode(),
			},
		),
	}

	for _, k := range keys[1:] {
		matchers = append(matchers, KeyValueExpr(BasicLit(strconv.Quote(k)), AnyNode()))
	}

	return matchers
}

func Commands() []match.GoNode {
	return []match.GoNode{
		Func(Ident("Commands")),
		TypeOf(&dst.ReturnStmt{}),
		CompositeLit(
			ArrayType(
				SelectorExpr(
					Ident("console"),
					Ident("Command"),
				),
				AnyNode(),
			),
		),
	}
}

func Imports() []match.GoNode {
	return []match.GoNode{
		GoNode{
			match: func(n dst.Node) bool {
				if block, ok := n.(*dst.GenDecl); ok {
					return block.Tok == token.IMPORT
				}

				return false
			},
		},
	}
}

func Jobs() []match.GoNode {
	return []match.GoNode{
		Func(Ident("Jobs")),
		TypeOf(&dst.ReturnStmt{}),
		CompositeLit(
			ArrayType(
				SelectorExpr(
					Ident("queue"),
					Ident("Job"),
				),
				AnyNode(),
			),
		),
	}
}

func Migrations() []match.GoNode {
	return []match.GoNode{
		Func(Ident("Migrations")),
		TypeOf(&dst.ReturnStmt{}),
		CompositeLit(
			ArrayType(
				SelectorExpr(
					Ident("schema"),
					Ident("Migration"),
				),
				AnyNode(),
			),
		),
	}
}

func Providers() []match.GoNode {
	return []match.GoNode{
		Func(
			Ident("Providers"),
		),
		CompositeLit(
			ArrayType(
				SelectorExpr(
					Ident("foundation"),
					Ident("ServiceProvider"),
				),
				AnyNode(),
			),
		),
	}
}

// Deprecated: ProvidersFallback represents the old logic of registering
// service providers inside the `config/app.go` file.
//
// This pattern is deprecated and will be removed in future versions.
// Please migrate to the new Providers() function-based registration approach.
// NOTE: in the current internal setup files, ProvidersFallback is only used for
// uninstall (not for install). It will continue to serve as an uninstall-only
// fallback for the next few releases — during which both Providers() and
// ProvidersFallback matchers will be supported to ensure backward compatibility
// with older versions. After that deprecation period, ProvidersFallback will
// be removed entirely.
func ProvidersFallback() []match.GoNode {
	return []match.GoNode{
		Func(Ident("init")),
		CallExpr(
			SelectorExpr(
				Ident("config"),
				Ident("Add"),
			),
			GoNodes{
				BasicLit(strconv.Quote("app")),
				AnyNode(),
			},
		),
		KeyValueExpr(BasicLit(strconv.Quote("providers")), AnyNode()),
		CompositeLit(
			ArrayType(
				SelectorExpr(
					Ident("foundation"),
					Ident("ServiceProvider"),
				),
				AnyNode(),
			),
		),
	}
}

func RegisterFunc() []match.GoNode {
	return []match.GoNode{Func(Ident("Register"))}
}

func BootFunc() []match.GoNode {
	return []match.GoNode{Func(Ident("Boot"))}
}

func Seeders() []match.GoNode {
	return []match.GoNode{
		Func(Ident("Seeders")),
		TypeOf(&dst.ReturnStmt{}),
		CompositeLit(
			ArrayType(
				SelectorExpr(
					Ident("seeder"),
					Ident("Seeder"),
				),
				AnyNode(),
			),
		),
	}
}

func ValidationRules() []match.GoNode {
	return []match.GoNode{
		Func(Ident("rules")),
		TypeOf(&dst.ReturnStmt{}),
		CompositeLit(
			ArrayType(
				SelectorExpr(
					Ident("validation"),
					Ident("Rule"),
				),
				AnyNode(),
			),
		),
	}
}

func ValidationFilters() []match.GoNode {
	return []match.GoNode{
		Func(Ident("filters")),
		TypeOf(&dst.ReturnStmt{}),
		CompositeLit(
			ArrayType(
				SelectorExpr(
					Ident("validation"),
					Ident("Filter"),
				),
				AnyNode(),
			),
		),
	}
}

// FoundationSetup matches the Boot function containing foundation.Setup() chain calls.
// It matches both patterns:
//   - foundation.Setup().WithConfig(...).Run()
//   - foundation.Setup().WithMiddleware(...).WithConfig(...).Run()
//   - foundation.Setup().WithCommand(...).Run()
//
// Example usage:
//
//	GoFile("bootstrap/app.go").
//	    Find(match.FoundationSetup()).
//	    Modify(foundationSetupMiddleware(middleware)).
//	    Apply()
func FoundationSetup() []match.GoNode {
	return []match.GoNode{
		Func(Ident("Boot")),
		TypeOf(&dst.ExprStmt{}),
	}
}
