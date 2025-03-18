package packages

import (
	"fmt"
	"go/token"
	"strconv"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"

	"github.com/goravel/framework/contracts/packages"
	"github.com/goravel/framework/support/color"
)

var providerMatchers = []packages.GoNodeMatcher{
	MatchFuncDecl(MatchIdent("init")),
	MatchCallExpr(
		MatchSelectorExpr(
			MatchIdent("config"),
			MatchIdent("Add"),
		),
		MatchGoNodes{
			MatchBasicLit(strconv.Quote("app")),
			MatchAnyNode(),
		},
	),
	MatchKeyValueExpr(MatchBasicLit(strconv.Quote("providers")), MatchAnyNode()),
	MatchCompositeLit(
		MatchArrayType(
			MatchSelectorExpr(
				MatchIdent("foundation"),
				MatchIdent("ServiceProvider"),
			),
			MatchAnyNode(),
		),
	),
}

// AddConfigSpec adds a configuration key to the config file.
func AddConfigSpec(path, name, statement string) packages.GoNodeModifier {
	return ModifyGoNode{
		Matchers: buildConfigMatchers(strings.Split(path, ".")),
		Action: func(cursor *dstutil.Cursor) {
			var mv *dst.CompositeLit
			switch node := cursor.Node().(type) {
			case *dst.KeyValueExpr:
				mv = node.Value.(*dst.CompositeLit)
			case *dst.CallExpr:
				mv = node.Args[1].(*dst.CompositeLit)
			}
			key := WrapNewline(&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(name)})
			if KeyExists(key, mv.Elts) {
				color.Warningln(fmt.Sprintf("Key [%s] already exists in [%s]", name, path))
				return
			}

			mv.Elts = append(mv.Elts, WrapNewline(&dst.KeyValueExpr{
				Key:   key,
				Value: WrapNewline(MustParseStatement(statement)).(dst.Expr),
			}))

		},
	}
}

// AddImportSpec adds an import statement to the file.
func AddImportSpec(path string, name ...string) packages.GoNodeModifier {
	matcher := MatchLastOf(MatchTypeOf(&dst.ImportSpec{}))
	if isThirdParty(path) {
		matcher = MatchTypeOf(&dst.ImportSpec{})
	}

	return ModifyGoNode{
		Matchers: []packages.GoNodeMatcher{matcher},
		Action: func(cursor *dstutil.Cursor) {
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

			cursor.InsertAfter(WrapNewline(im))
		},
	}
}

// AddProviderSpec adds a provider to the service providers array if it doesn't already exist.
func AddProviderSpec(statement string) packages.GoNodeModifier {
	return providerSpecInsertTo(
		append(providerMatchers, MatchLastOf(MatchTypeOf(dst.Expr(nil)))),
		statement,
		false,
	)
}

// AddProviderSpecAfter adds a provider to the service providers array after a specific provider.
func AddProviderSpecAfter(statement, after string) packages.GoNodeModifier {
	return providerSpecInsertTo(
		append(providerMatchers, MatchEqualStatement(after)),
		statement,
		false,
	)
}

// AddProviderSpecBefore adds a provider to the service providers array before a specific provider.
func AddProviderSpecBefore(statement, before string) packages.GoNodeModifier {
	return providerSpecInsertTo(
		append(providerMatchers, MatchEqualStatement(before)),
		statement,
		true,
	)
}

// RemoveConfigSpec removes a configuration key from the config file.
func RemoveConfigSpec(path string) packages.GoNodeModifier {
	return ModifyGoNode{
		Matchers: buildConfigMatchers(strings.Split(path, ".")),
		Action: func(cursor *dstutil.Cursor) {
			cursor.Delete()
		},
	}
}

// RemoveImportSpec removes an import statement from the file.
func RemoveImportSpec(path string, name ...string) packages.GoNodeModifier {
	return ModifyGoNode{
		Matchers: []packages.GoNodeMatcher{
			MatchImportSpec(path, name...),
		},
		Action: func(cursor *dstutil.Cursor) {
			cursor.Delete()
		},
	}
}

// RemoveProviderSpec removes a provider from the service providers array.
func RemoveProviderSpec(statement string) packages.GoNodeModifier {
	return ModifyGoNode{
		Matchers: append(providerMatchers, MatchEqualStatement(statement)),
		Action: func(cursor *dstutil.Cursor) {
			cursor.Delete()
		},
	}
}

// ReplaceConfigSpec replaces a configuration key in the config file.
func ReplaceConfigSpec(path, statement string) packages.GoNodeModifier {
	return ModifyGoNode{
		Matchers: buildConfigMatchers(strings.Split(path, ".")),
		Action: func(cursor *dstutil.Cursor) {
			cursor.Node().(*dst.KeyValueExpr).Value = WrapNewline(MustParseStatement(statement)).(dst.Expr)
		},
	}
}

func buildConfigMatchers(keys []string) []packages.GoNodeMatcher {
	var matchers = []packages.GoNodeMatcher{
		MatchFuncDecl(MatchIdent("init")),
		MatchCallExpr(
			MatchSelectorExpr(
				MatchIdent("config"),
				MatchIdent("Add"),
			),
			MatchGoNodes{
				MatchBasicLit(strconv.Quote(keys[0])),
				MatchAnyNode(),
			},
		),
	}

	for _, k := range keys[1:] {
		matchers = append(matchers, MatchKeyValueExpr(MatchBasicLit(strconv.Quote(k)), MatchAnyNode()))
	}

	return matchers
}

func providerSpecInsertTo(matchers []packages.GoNodeMatcher, statement string, before bool) packages.GoNodeModifier {
	return ModifyGoNode{
		Matchers: matchers,
		Action: func(cursor *dstutil.Cursor) {
			provider := MustParseStatement(statement).(dst.Expr)
			if ExprExists(provider, cursor.Parent().(*dst.CompositeLit).Elts) {
				color.Warningln(fmt.Sprintf("Provider [%s] already exists", statement))
				return
			}

			if before {
				cursor.InsertBefore(provider)
			} else {
				cursor.InsertAfter(provider)
			}
		},
	}
}

func isThirdParty(importPath string) bool {
	// Third party package import path usually contains "." (".com", ".org", ...)
	// This logic is taken from golang.org/x/tools/imports package.
	return strings.Contains(importPath, ".")
}
